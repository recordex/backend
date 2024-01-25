package lib

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os/exec"
	"strings"

	"golang.org/x/xerrors"
)

// GetAuthorizationBarerTokenFromHeader は Authorization ヘッダーから Bearer トークンを取得します。
func GetAuthorizationBarerTokenFromHeader(header http.Header) (string, error) {
	authHeader := header.Get("Authorization")
	if authHeader == "" {
		return "", xerrors.Errorf("Authorization ヘッダーが設定されていません。")
	}

	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return "", xerrors.Errorf("token の形式が不正です。")
	}

	barerToken := splitToken[1]

	return barerToken, nil
}

// CalculateFileHash は引数で指定されたファイルの SHA256 ハッシュ値を計算します。
func CalculateFileHash(fileHeader *multipart.FileHeader) (string, error) {
	// ファイルの open
	file, err := fileHeader.Open()
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Printf("ファイルの close に失敗しました。fileName -> %s: %+v", SanitizeInput(fileHeader.Filename), err)
		}
	}(file)
	if err != nil {
		return "", xerrors.Errorf("ファイルの open に失敗しました。fileName -> %s: %+v", SanitizeInput(fileHeader.Filename), err)
	}

	// ハッシュ値の計算
	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", xerrors.Errorf("ハッシュ値の計算中にエラーが発生しました。: %+v", err)
	}

	hashValue := hex.EncodeToString(hasher.Sum(nil))
	return hashValue, nil
}

// SanitizeInput は引数で指定された文字列にエスケープ処理をします
func SanitizeInput(input string) string {
	// 改行文字とタブ文字のエスケープ
	input = strings.Replace(input, "\n", "\\n", -1)
	input = strings.Replace(input, "\r", "\\r", -1)
	input = strings.Replace(input, "\t", "\\t", -1)

	return input
}

// DiffPDF は引数で指定された 2 つの PDF ファイルの差分を見られるようにした PDF を新たに作成。
// 新たに作成した差分ファイルのファイル名を返します。
func DiffPDF(ctx context.Context, pdfPath string, changedPDFPath string) (string, error) {
	dstFileName := "diff.pdf"
	doneCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		defer close(doneCh)
		command := SanitizeInput(fmt.Sprintf("diff-pdf -mv --output-diff=%s %s %s", dstFileName, pdfPath, changedPDFPath))
		_, err := exec.Command(command).CombinedOutput()
		// diff-pdf コマンドは実効性工事でも exit status 1 を返すので、
		// exit status 1 以外のエラーが発生した場合のみエラーとして扱う
		if err != nil && err.Error() != "exit status 1" {
			errCh <- xerrors.Errorf("diff-pdf コマンドの実行に失敗しました。: %+v", err)
		}
	}()

	// context がタイムアウトしているか、goroutine の中でエラーが発生しているかを確認
	select {
	case <-ctx.Done():
		return "", xerrors.Errorf("差分ファイルの生成がタイムアウトしました。pdfPath -> %s, changedPDFPath -> %s: %+v", pdfPath, changedPDFPath, ctx.Err())
	case err := <-errCh:
		return "", err
	case <-doneCh:
		log.Println("差分が記録された PDF を生成しました。:", dstFileName)
		return dstFileName, nil
	}
}
