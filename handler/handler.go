package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/recordex/backend/api/gen"
	"github.com/recordex/backend/config"
	"github.com/recordex/backend/infrastructure/ethereum"
	GCPInfrastructure "github.com/recordex/backend/infrastructure/gcp"
	"github.com/recordex/backend/lib"
)

// GetDiffPDF はクライアントから送られたファイルと、ブロックチェーンに記録されている最新のファイルハッシュのファイルの差分を色付けした PDF を新たに作成し、クライアントに返します
// 比較対象のファイルのハッシュ値も返します
func GetDiffPDF(c echo.Context) error {
	ctx := context.Background()

	// ファイルヘッダを取得
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ファイルの取得に失敗しました。：%+v", err))
	}

	var eg errgroup.Group
	eg.Go(func() error {
		uploadedFile, err := fileHeader.Open()
		defer func(src multipart.File) {
			err := src.Close()
			if err != nil {
				log.Printf("ファイルの close に失敗しました。fileName -> %s: %+v", lib.SanitizeInput(fileHeader.Filename), err)
			}
		}(uploadedFile)
		if err != nil {
			return xerrors.Errorf("ファイルのオープンに失敗しました。fileName -> %s: %+v", lib.SanitizeInput(fileHeader.Filename), err)
		}

		file, err := os.Create(lib.SanitizeInput(fileHeader.Filename))
		if err != nil {
			return xerrors.Errorf("ファイルの作成に失敗しました。fileName -> %s: %+v", lib.SanitizeInput(fileHeader.Filename), err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Printf("ファイルの close に失敗しました。fileName -> %s: %+v", lib.SanitizeInput(fileHeader.Filename), err)
			}
		}(file)

		_, err = io.Copy(file, uploadedFile)
		if err != nil {
			return xerrors.Errorf("ファイルの書き込みに失敗しました。fileName -> %s: %+v", lib.SanitizeInput(fileHeader.Filename), err)
		}

		return nil
	})

	// ブロックチェーンに記録されている最新のファイルを取得するためのチャネルを作成
	newestFileNameCh := make(chan string, 1)
	defer close(newestFileNameCh)
	eg.Go(func() error {
		// ファイルの最新バージョンのハッシュ値をブロックチェーンから取得
		newestFileMetadata, err := ethereum.GetNewestFileMetadata(ctx, fileHeader.Filename)
		if err != nil {
			return xerrors.Errorf("最新のファイルメタデータの取得に失敗しました。fileName -> %s: %+v", lib.SanitizeInput(fileHeader.Filename), err)
		}

		// newsFileHash がファイルネームとなる
		newestFileHash := string(newestFileMetadata.Hash[:])
		newestFileNameCh <- newestFileHash
		// GCS から最新のファイルを取得
		rc, err := GCPInfrastructure.GetStorageClient().Bucket(config.Get().CloudStorageBucketName).Object(newestFileHash).NewReader(ctx)
		defer func(rc *storage.Reader) {
			err := rc.Close()
			if err != nil {
				log.Printf("cloud storage reader の close に失敗しました。fileName -> %s: %+v", newestFileHash, err)
			}
		}(rc)
		if err != nil {
			return xerrors.Errorf("GCS からの最新ファイルのダウンロードに失敗しました。fileName -> %s: %+v", newestFileHash, err)
		}

		// ローカルのファイルに書き込む
		file, err := os.Create(newestFileHash)
		if err != nil {
			return xerrors.Errorf("ファイルの作成に失敗しました。fileName -> %s: %+v", newestFileHash, err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Printf("ファイルの close に失敗しました。fileName -> %s: %+v", newestFileHash, err)
			}
		}(file)

		_, err = io.Copy(file, rc)
		if err != nil {
			return xerrors.Errorf("ファイルの書き込みに失敗しました。fileName -> %s: %+v", newestFileHash, err)
		}

		return nil
	})

	err = eg.Wait()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	newestFileName := <-newestFileNameCh
	// ファイルの差分を色付けした PDF を作成
	diffFileName, err := lib.DiffPDF(ctx, newestFileName, fileHeader.Filename)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("PDF の差分ファイルの作成に失敗しました。: %+v", err))
	}

	go func() {
		err := os.Remove(lib.SanitizeInput(fileHeader.Filename))
		if err != nil {
			log.Printf("ファイルの削除に失敗しました。fileName -> %s: %+v", fileHeader.Filename, err)
		}
	}()
	go func() {
		err := os.Remove(lib.SanitizeInput(newestFileName))
		if err != nil {
			log.Printf("ファイルの削除に失敗しました。fileName -> %s: %+v", newestFileName, err)
		}
	}()

	return c.File(diffFileName)
}

// PostRecord はトランザクション ID が正しいかどうかをブロックチェーンに記録されているデータからチェックし、ファイルを GCS にアップロードする
func PostRecord(c echo.Context) error {
	ctx := context.Background()

	transactionHash := c.FormValue("transaction_hash")
	var req gen.PostRecordRequest
	req.TransactionHash = transactionHash
	// リクエストのバリデーション
	err := req.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("request が不正です。request -> %s：%+v", req.String(), err))
	}

	// ファイルヘッダを取得
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ファイルの取得に失敗しました。：%+v", err))
	}

	file, err := fileHeader.Open()
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Printf("ファイルの close に失敗しました。fileName -> %s: %+v", lib.SanitizeInput(fileHeader.Filename), err)
		}
	}(file)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ファイルのオープンに失敗しました。fileName -> %s：%+v", lib.SanitizeInput(fileHeader.Filename), err))
	}

	fileHash, err := lib.CalculateFileHash(fileHeader)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ファイルのハッシュ値の計算に失敗しました。：%+v", err))
	}
	log.Printf("ファイルのハッシュ値 fileName -> %s: %s", lib.SanitizeInput(fileHeader.Filename), fileHash)

	var isTransactionHashValid bool
	doneChan, errChan := make(chan struct{}), make(chan error)
	// タイムアウトは20秒に設定
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	go func(transactionHash string, fileHash string) {
		defer close(doneChan)
		// トラザクションが進行中かを1秒ごとに確認
		for isPending, err := ethereum.IsTransactionPending(ctx, transactionHash); isPending; isPending, err = ethereum.IsTransactionPending(ctx, transactionHash) {
			if err != nil {
				errChan <- err
				return
			}
			time.Sleep(1 * time.Second)
		}
		isTransactionHashValid, err = ethereum.IsRecordTransactionHashValid(ctx, transactionHash, fileHash)
		errChan <- err
	}(transactionHash, fileHash)
	select {
	case <-ctx.Done():
		// 5秒以内に処理が終わらなかった場合はエラーを返す
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("transactionHash の検証がタイムアウトしました。transactionHash -> %s：%+v", transactionHash, err))
	case err := <-errChan:
		// IsRecordTransactionHashValid がエラーを返した場合もエラーを返す
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("transactionHash の検証に失敗しました。transactionHash -> %s：%+v", transactionHash, err))
		}
	case <-doneChan:
		// IsRecordTransactionHashValid が正常に終了した場合は何もしない
	}

	// トランザクション ID が正しいかどうかをチェック
	if !isTransactionHashValid {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("transactionHash が不正です。transactionHash -> %s", lib.SanitizeInput(transactionHash)))
	}
	log.Printf("transactionHash の検証に成功しました。transactionHash -> %s", lib.SanitizeInput(transactionHash))

	// cloud storage にファイルをアップロード
	// アップロードするファイルのファイル名は、そのファイルのハッシュ値にする
	// 参照: https://cloud.google.com/storage/docs/samples/storage-upload-file?hl=ja#storage_upload_file-go
	wc := GCPInfrastructure.GetStorageClient().Bucket(config.Get().CloudStorageBucketName).Object(fileHash).NewWriter(context.Background())
	defer func(wc *storage.Writer) {
		err := wc.Close()
		if err != nil {
			log.Printf("cloud storage writer の close に失敗しました。fileName -> %s: %+v", lib.SanitizeInput(wc.Name), err)
		}
	}(wc)
	// ファイルを cloud storage にアップロード
	_, err = io.Copy(wc, file)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("cloud storage へのファイルのアップロードに失敗しました。fileName -> %s：%+v", lib.SanitizeInput(wc.Name), err))
	}

	log.Printf("cloud storage へのファイルのアップロードに成功しました。fileName -> %s", lib.SanitizeInput(wc.Name))
	resp := gen.PostRecordResponse{
		FileName:        fileHeader.Filename,
		TransactionHash: transactionHash,
	}
	err = resp.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("不正な response が作成されました。fileName -> %s, transactionHash -> %s：%+v", lib.SanitizeInput(fileHeader.Filename), transactionHash, err))
	}

	return c.JSON(http.StatusOK, resp)
}
