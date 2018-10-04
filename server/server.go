package server

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/zipcar/vault-cfcs/logger"
)

const serverAddr = ":1337"

func ListenAndServe() {
	logger.InitializeLogger()
	logger.Log.Print("Server starting")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())

	e.GET("/v1/health", healthHandler)
	e.POST("/v1/data", dataPostHandler)

	logger.Log.Fatal(e.Start(serverAddr))
}
