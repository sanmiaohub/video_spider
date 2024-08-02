package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

func main() {
	e := echo.New()

	// 异常
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		_ = c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}

	// 中间件
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	// 接口
	e.GET("/", hello)
	e.POST("/analysis", analysis)
	// 文件
	e.Static("/video", "video")

	e.Logger.Fatal(e.Start(":8080"))
}
