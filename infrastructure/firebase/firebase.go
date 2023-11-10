package firebase

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var firebaseApp *firebase.App

func init() {
	opt := option.WithCredentialsJSON([]byte(os.Getenv("FIREBASE_KEYFILE_JSON")))
	var err error
	firebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("firebase app の初期化に失敗しました: %v", err)
	}
}

func GetFirebaseApp() *firebase.App {
	return firebaseApp
}
