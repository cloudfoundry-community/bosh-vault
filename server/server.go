package server

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/zipcar/vault-cfcs/logger"
)

const serverAddr = "localhost:1337"

func ListenAndServe() {
	logger.InitializeLogger()
	logger.Log.Printf("Server starting at: %s", serverAddr)
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"level":"info","msg":"Request data",time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out}}` + "\n",
		Skipper: middleware.DefaultSkipper,
		Output:  logger.Log.Out,
	}))

	e.Use(middleware.Secure())

	e.GET("/v1/health", healthHandler)
	e.POST("/v1/data", dataPostHandler)

	logger.Log.Fatal(e.Start(serverAddr))
}
