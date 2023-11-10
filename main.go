package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	Middleware "github.com/recordex/backend/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
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
	auth.Use(Middleware.FirebaseAuth())

	// Routes
	e.GET("/health", health)
	auth.GET("/auth", authorization)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// Handler
func health(c echo.Context) error {
	return c.String(http.StatusOK, "Healthy.")
}

func authorization(c echo.Context) error {
	return c.String(http.StatusOK, "Authorized.")
}
