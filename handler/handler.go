package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/labstack/echo/v4"

	"github.com/recordex/backend/api/gen"
	"github.com/recordex/backend/config"
	"github.com/recordex/backend/infrastructure/ethereum"
	GCPInfrastructure "github.com/recordex/backend/infrastructure/gcp"
	"github.com/recordex/backend/lib"
)

func PostRecord(c echo.Context) error {
	transactionHash := c.FormValue("transaction_hash")
	var req gen.PostTransactionRequest
	req.TransactionHash = transactionHash
	// リクエストのバリデーション
	err := req.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("request が不正です。request -> %s：%+v", req.String(), err))
	}

	// ファイルをフォームから取得
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
		return err
	}

	fileHash, err := lib.CalculateFileHash(fileHeader)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ファイルのハッシュ値の計算に失敗しました。：%+v", err))
	}
	log.Printf("ファイルのハッシュ値 fileName -> %s: %s", lib.SanitizeInput(fileHeader.Filename), fileHash)

	var isTransactionHashValid bool
	doneChan, errChan := make(chan struct{}), make(chan error)
	// タイムアウトは20秒に設定
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	go func(transactionHash string, fileHash string) {
		defer close(doneChan)
		// トラザクションが進行中かを1秒ごとに確認
		for isPending, err := ethereum.IsTransactionPending(transactionHash); isPending; isPending, err = ethereum.IsTransactionPending(transactionHash) {
			if err != nil {
				errChan <- err
				return
			}
			time.Sleep(1 * time.Second)
		}
		isTransactionHashValid, err = ethereum.IsRecordTransactionHashValid(transactionHash, fileHash)
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
	// 参照: https://cloud.google.com/storage/docs/samples/storage-upload-file?hl=ja#storage_upload_file-go
	wc := GCPInfrastructure.GetStorageClient().Bucket(config.Get().CloudStorageBucketName).Object(fileHeader.Filename).NewWriter(context.Background())
	defer func(wc *storage.Writer) {
		err := wc.Close()
		if err != nil {
			log.Printf("cloud storage writer の close に失敗しました。fileName -> %s: %+v", lib.SanitizeInput(wc.Name), err)
		}
	}(wc)
	// ファイルを cloud storage にアップロード
	_, err = io.Copy(wc, file)
	if err != nil {
		return err
	}

	log.Printf("cloud storage へのファイルのアップロードに成功しました。fileName -> %s, fileHash -> %s", lib.SanitizeInput(wc.Name), fileHash)
	resp := gen.PostTransactionResponse{
		FileName:        fileHeader.Filename,
		TransactionHash: transactionHash,
	}
	err = resp.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("不正な response が作成されました。fileName -> %s, transactionHash -> %s：%+v", lib.SanitizeInput(fileHeader.Filename), transactionHash, err))
	}

	return c.JSON(http.StatusOK, resp)
}
