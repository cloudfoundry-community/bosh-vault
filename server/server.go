package server

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/zipcar/bosh-vault/config"
	"github.com/zipcar/bosh-vault/logger"
	"github.com/zipcar/bosh-vault/secret"
	"github.com/zipcar/bosh-vault/store"
	"github.com/zipcar/bosh-vault/uaa"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const healthUri = "/v1/health"
const dataUri = "/v1/data"

type BvContext struct {
	echo.Context
	Config config.Configuration
	Log    *logrus.Logger
	Store  secret.Store
}

func ListenAndServe(bvConfig config.Configuration) {

	// config server ALWAYS needs TLS
	if (bvConfig.Tls.Cert == "" || bvConfig.Tls.Key == "") && !bvConfig.Debug.DisableTls {
		logger.Log.Fatal("unable to start bosh-vault without tls_cert_path and tls_key_path being set")
	}

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	storeClient := store.GetStore(bvConfig)

	// Support UAA Authorization if enabled (and broad authentication for now too)
	// Will allow connections if JWT contains the expected audience claim
	// Skip if UAA isn't enabled or the request is for the health endpoint
	if !bvConfig.Debug.DisableAuth {
		uaaClient := uaa.GetUaa(bvConfig)

		e.Use(uaaClient.AuthMiddleware(uaa.MiddlewareConfig{
			Skipper: func(c echo.Context) bool {
				return c.Request().RequestURI == healthUri
			},
		}))
	}

	// middleware function that sets a custom context exposing our configuration and logger to handler functions
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// return 500 if the store isn't healthy
			if !storeClient.Healthy() {
				return echo.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}
			configContext := &BvContext{
				Context: c,
				Config:  bvConfig,
				Log:     logger.Log,
				Store:   storeClient,
			}
			return next(configContext)
		}
	})

	e.Use(middleware.Secure())

	e.GET(healthUri, healthCheckHandler)

	e.POST(dataUri, dataPostHandler)
	e.PUT(dataUri, dataPutHandler)
	e.GET(fmt.Sprintf("%s/:id", dataUri), dataGetByIdHandler)
	e.GET(dataUri, dataGetByNameHandler)
	e.DELETE(dataUri, dataDeleteHandler)

	// Start server
	go func() {
		logger.Log.Infof("starting bosh-vault api server at %s", bvConfig.Api.Address)
		if bvConfig.Debug.DisableTls {
			logger.Log.Error("!!!!!!!!!! DEBUG MODE ACTIVE TLS DISABLED !!!!!!!!!")
			if err := e.Start(bvConfig.Api.Address); err != nil {
				logger.Log.Info("shutting down the bosh-vault api server")
			}
		} else {
			if err := e.StartTLS(bvConfig.Api.Address, bvConfig.Tls.Cert, bvConfig.Tls.Key); err != nil {
				logger.Log.Info("shutting down the bosh-vault api server")
			}
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
