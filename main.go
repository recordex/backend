package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/recordex/backend/handler"
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

	auth := e.Group("/auth")
	auth.Use(Middleware.FirebaseAuth)

	e.GET("/health", health)
	e.GET("/diff/pdf", handler.GetDiffPDF)
	e.POST("/record", handler.PostRecord)

	auth.GET("/", authorize)
	auth.POST("/record", handler.PostRecord)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func health(c echo.Context) error {
	return c.String(http.StatusOK, "Healthy.")
}

func authorize(c echo.Context) error {
	return c.String(http.StatusOK, "Authorized.")
}
