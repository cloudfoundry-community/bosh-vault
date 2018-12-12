package server

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/zipcar/bosh-vault/auth"
	"github.com/zipcar/bosh-vault/config"
	"github.com/zipcar/bosh-vault/health"
	"github.com/zipcar/bosh-vault/logger"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"time"
)

type BvContext struct {
	echo.Context
	Config    config.Configuration
	Log       *logrus.Logger
	UaaClient *auth.UaaClient
}

func ListenAndServe(bvConfig config.Configuration) {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	uaaClient := auth.GetUaaClient(bvConfig)

	// Support UAA Authorization if enabled (and broad authentication for now too)
	// Will allow connections if JWT contains the expected audience claim
	e.Use(uaaClient.AuthMiddleware())

	// middleware function that sets a custom context exposing our configuration, logger, and a UAA client to handler functions
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			configContext := &BvContext{
				Context:   c,
				Config:    bvConfig,
				Log:       logger.Log,
				UaaClient: uaaClient,
			}
			return h(configContext)
		}
	})

	e.Use(middleware.Secure())

	e.GET(health.HealthCheckUri, health.HealthCheckHandler)

	e.POST("/v1/data", dataPostHandler)
	e.PUT("/v1/data", dataPutHandler)
	e.GET("/v1/data/:id", dataGetByIdHandler)
	e.GET("/v1/data", dataGetByNameHandler)
	e.DELETE("v1/data", dataDeleteHandler)

	if bvConfig.Tls.Cert == "" || bvConfig.Tls.Key == "" {
		logger.Log.Fatal("unable to start bosh-vault without tls_cert_path and tls_key_path being set")
	}

	// Start server
	go func() {
		logger.Log.Infof("starting bosh-vault api server at %s", bvConfig.Api.Address)
		if err := e.StartTLS(bvConfig.Api.Address, bvConfig.Tls.Cert, bvConfig.Tls.Key); err != nil {
			logger.Log.Info("shutting down the bosh-vault api server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	// Gracefully shutdown the server if it has not shutdown within 10 seconds then force it to shutdown
	logger.Log.Info("received shutdown signal, shutting down the bosh-vault api server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(bvConfig.Api.DrainTimeout)*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Log.Error(err)
	}
}
