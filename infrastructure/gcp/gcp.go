package gcp

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
)

var storageClient *storage.Client

func init() {
	var err error
	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("storage client の初期化に失敗しました: %v", err)
	}
}

func GetStorageClient() *storage.Client {
	return storageClient
}
