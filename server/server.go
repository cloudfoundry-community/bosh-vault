package server

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/zipcar/vault-cfcs/auth"
	"github.com/zipcar/vault-cfcs/config"
	"github.com/zipcar/vault-cfcs/logger"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"time"
)

type VcfcsContext struct {
	echo.Context
	Config    config.Configuration
	Log       *logrus.Logger
	UaaClient *auth.UaaClient
}

func ListenAndServe(vcfcsConfig config.Configuration) {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	uaaClient := auth.GetUaaClient(vcfcsConfig)

	// Support UAA Authorization (and broad authentication for now too)
	// Will allow connections if JWT contains the expected audience claim
	e.Use(uaaClient.AuthMiddleware())

	e.Use(middleware.Secure())

	// middleware function that sets a custom context exposing our configuration, logger, and a UAA client to handler functions
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			configContext := &VcfcsContext{
				Context:   c,
				Config:    vcfcsConfig,
				Log:       logger.Log,
				UaaClient: uaaClient,
			}
			return h(configContext)
		}
	})

	e.Use(middleware.Secure())

	e.GET("/v1/health", healthHandler)
	e.POST("/v1/data", dataPostHandler)
	e.GET("/v1/data/:id", dataGetByIdHandler)
	e.GET("/v1/data", dataGetByNameHandler)

	if vcfcsConfig.Tls.Cert == "" || vcfcsConfig.Tls.Key == "" {
		logger.Log.Fatal("unable to start vault-cfcs without tls_cert_path and tls_key_path being set")
	}

	// Start server
	go func() {
		logger.Log.Infof("starting vault-cfcs api server at %s", vcfcsConfig.ApiListenAddress)
		if err := e.StartTLS(vcfcsConfig.ApiListenAddress, vcfcsConfig.Tls.Cert, vcfcsConfig.Tls.Key); err != nil {
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
