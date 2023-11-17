package middleware

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	FirebaseInfrastructure "github.com/recordex/backend/infrastructure/firebase"
	"github.com/recordex/backend/lib"
)

func FirebaseAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		firebaseApp := FirebaseInfrastructure.GetFirebaseApp()
		authClient, err := firebaseApp.Auth(ctx)
		if err != nil {
			log.Println("firebase アプリケーションの auth クライアントの作成に失敗しました。")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": err.Error(),
			})
		}

		barerToken, err := lib.GetAuthorizationBarerTokenFromHeader(c.Request().Header)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}

		token, err := authClient.VerifyIDToken(ctx, barerToken)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": err.Error(),
			})
		}

		log.Printf("idToken の検証に成功しました。uid -> %s", token.UID)
		c.Set("userId", token.UID)
		return next(c)
	}
}
