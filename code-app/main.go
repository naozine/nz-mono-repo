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
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        button { padding: 10px 20px; margin: 10px 0; background-color: #3490dc; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background-color: #2779bd; }
    </style>
</head>
<body>
    <h1>シンプルなトップページ</h1>
    <p>Echo フレームワークを使用したシンプルなトップページです。</p>
    
    <div x-data="{ message: 'こんにちは！', clicked: false }">
        <button @click="clicked = !clicked" x-text="clicked ? 'クリック済み' : 'ボタンをクリック'"></button>
        <p x-show="clicked" x-text="message"></p>
    </div>
</body>
</html>`)
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
