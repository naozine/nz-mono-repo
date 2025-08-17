package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create an Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `
<!DOCTYPE html>
<html>
<head>
    <title>Simple Top Page</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>シンプルなトップページ</h1>
    <p>Echo フレームワークを使用したシンプルなトップページです。</p>
</body>
</html>`)
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
