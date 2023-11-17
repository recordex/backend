package main

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
	"github.com/labstack/echo/v4/middleware"

	"github.com/recordex/backend/api/gen"
	"github.com/recordex/backend/config"
	"github.com/recordex/backend/infrastructure/ethereum"
	GCPInfrastructure "github.com/recordex/backend/infrastructure/gcp"
	"github.com/recordex/backend/lib"
	Middleware "github.com/recordex/backend/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORSのミドルウェアを全許可の設定で追加
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	auth := e.Group("")
	auth.Use(Middleware.FirebaseAuth)

	e.GET("/health", health)
	e.POST("/record", record)

	auth.GET("/auth", authorize)
	auth.POST("/auth/record", record)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func health(c echo.Context) error {
	return c.String(http.StatusOK, "Healthy.")
}

func authorize(c echo.Context) error {
	return c.String(http.StatusOK, "Authorized.")
}

func record(c echo.Context) error {
	transactionHash := c.FormValue("transaction_hash")
	var req gen.PostTransactionRequest
	req.TransactionHash = transactionHash
	// リクエストのバリデーション
	err := req.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("request が不正です。：%+v", err))
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
	// タイムアウトは5秒に設定
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func(transactionHash string, fileHash string) {
		defer close(doneChan)
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

	return c.String(http.StatusOK, fmt.Sprintf("ファイルのアップロードに成功しました。fileName -> %s", lib.SanitizeInput(fileHeader.Filename)))
}
