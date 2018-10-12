package server

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/zipcar/vault-cfcs/config"
	"github.com/zipcar/vault-cfcs/logger"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"time"
)

type ConfigurationContext struct {
	echo.Context
	Config config.Configuration
	Log    *logrus.Logger
}

func ListenAndServe(vcfcsConfig config.Configuration) {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	// middleware function that sets a custom context exposing our configuration and logger to handler functions
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			configContext := &ConfigurationContext{
				Context: c,
				Config:  vcfcsConfig,
				Log:     logger.Log,
			}
			return h(configContext)
		}
	})

	e.Use(middleware.Secure())

	e.GET("/v1/health", healthHandler)
	e.POST("/v1/data", dataPostHandler)
	e.GET("/v1/data/:id", dataGetByIdHandler)

	// Start server
	go func() {
		logger.Log.Infof("starting vault-cfcs api server at %s", vcfcsConfig.ApiListenAddress)
		if err := e.StartTLS(vcfcsConfig.ApiListenAddress, "certs/local-dev.crt", "certs/local-dev.key"); err != nil {
			logger.Log.Info("shutting down the vault-cfcs api server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	// Gracefully shutdown the server if it has not shutdown within 10 seconds then force it to shutdown
	logger.Log.Info("received shutdown signal, shutting down the vault-cfcs api server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(vcfcsConfig.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Log.Error(err)
	}
}
