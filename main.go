package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/recordex/backend/api/gen"
	"github.com/recordex/backend/config"
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
	transactionID := c.FormValue("transaction_id")
	var req gen.PostTransactionRequest
	req.TransactionId = transactionID
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
			log.Printf("ファイルの close に失敗しました。fileName -> %s: %+v", fileHeader.Filename, err)
		}
	}(file)
	if err != nil {
		return err
	}

	hashValue, err := lib.CalculateFileHash(fileHeader)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ファイルのハッシュ値の計算に失敗しました。：%+v", err))
	}
	log.Printf("ファイルのハッシュ値 fileName -> %s: %s", fileHeader.Filename, hashValue)

	// cloud storage にファイルをアップロード
	// 参照: https://cloud.google.com/storage/docs/samples/storage-upload-file?hl=ja#storage_upload_file-go
	wc := GCPInfrastructure.GetStorageClient().Bucket(config.Get().CloudStorageBucketName).Object(fileHeader.Filename).NewWriter(context.Background())
	defer func(wc *storage.Writer) {
		err := wc.Close()
		if err != nil {
			log.Printf("cloud storage writer の close に失敗しました。fileName -> %s: %+v", wc.Name, err)
		}
	}(wc)
	_, err = io.Copy(wc, file)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, fmt.Sprintf("ファイルのアップロードに成功しました。fileName -> %s", fileHeader.Filename))
}
