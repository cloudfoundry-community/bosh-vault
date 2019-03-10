package uaa

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/zipcar/bosh-vault/config"
	"github.com/zipcar/bosh-vault/logger"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const UaaAuthScheme = "bearer" // uaa doesn't use "Bearer" (the JWT default), but "bearer"

type Uaa struct {
	Config         config.UaaConfiguration
	Enabled        bool
	Endpoints      *UaaEndpoints
	httpClient     *http.Client
	SigningKeyData TokenKeyResponse
}

type MiddlewareConfig struct {
	Skipper middleware.Skipper
}

type UaaEndpoints struct {
	CheckToken string
	TokenKey   string
}

// @see: http://docs.cloudfoundry.org/api/uaa/version/release-candidate/#token-key-s
type TokenKeyResponse struct {
	Kid   string `json:"kid"`
	Alg   string `json:"alg"`
	Value string `json:"value"`
	Kty   string `json:"kty"`
	Use   string `json:"use"`
	N     string `json:"n"`
	E     string `json:"e"`
}

func GetUaa(bvConfig config.Configuration) *Uaa {

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		logger.Log.Error("problem reading system cert pool, if no UAA CA cert was passed in the config expect TLS errors")
		rootCAs = x509.NewCertPool()
	}

	if bvConfig.Uaa.Ca != "" {
		certs, err := ioutil.ReadFile(bvConfig.Uaa.Ca)
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", bvConfig.Uaa.Ca, err)
		}

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: bvConfig.Uaa.SkipVerify,
		RootCAs:            rootCAs,
	}

	// Setup a custom transport that trusts our UAA Ca as well as the system's trusted certs
	customTransport := &http.Transport{TLSClientConfig: tlsConfig}
	customHttpClient := &http.Client{
		Timeout:   time.Second * time.Duration(bvConfig.Uaa.Timeout),
		Transport: customTransport,
	}

	client := &Uaa{
		Enabled: bvConfig.Uaa.Enabled,
		Endpoints: &UaaEndpoints{
			CheckToken: fmt.Sprintf("%s/check_token", bvConfig.Uaa.Address),
			TokenKey:   fmt.Sprintf("%s/token_key", bvConfig.Uaa.Address),
		},
		httpClient: customHttpClient,
	}

	if bvConfig.Uaa.Enabled {
		// Update the key signing information for the UAA server once a day by default,
		// this will cut down on traffic to the UAA server
		ticker := time.NewTicker(time.Duration(bvConfig.Uaa.KeyRefreshInterval) * time.Second)
		go func() {
			for _ = range ticker.C {
				logger.Log.Debug("refreshing signing key info from UAA server")
				err := client.updateSigningKeyData()
				if err != nil {
					logger.Log.Error("error getting signing key info from UAA server, perhaps it's down? continuing to use cached signing key data...")
				}
			}
		}()
	}

	return client
}

func (uaa *Uaa) updateSigningKeyData() error {
	var signingKeyResp TokenKeyResponse

	resp, err := uaa.httpClient.Get(uaa.Endpoints.TokenKey)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received a status code %v when requesting token signing info", resp.Status)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(responseBody, &signingKeyResp)
	if err != nil {
		return err
	}

	// cache response
	uaa.SigningKeyData = signingKeyResp

	return nil
}

func (uaa *Uaa) AuthMiddleware(config MiddlewareConfig) echo.MiddlewareFunc {

	if uaa.SigningKeyData.Value == "" || uaa.SigningKeyData.Alg == "" {
		err := uaa.updateSigningKeyData()
		if err != nil {
			// connection lost with UAA
			logger.Log.Fatal(err)
		}
	}

	publicKey, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(uaa.SigningKeyData.Value))

	// The JWT middleware will handle basic authentication, our success handler does broad based audience claim authorization
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    publicKey,
		SigningMethod: uaa.SigningKeyData.Alg,
		AuthScheme:    UaaAuthScheme,
		Skipper:       config.Skipper,
		// JWT middleware handles basic validity checks, this successhandler is our custom audience check since UAA
		// returns a []string for the aud claim so single users can access multiple resources, the consequence is we can't
		// use the built in methods of the JWT middleware to validate the audience for us since we don't know what additional
		// audiences a given user may have
		SuccessHandler: uaa.validateAudience,
	})
}

func (uaa *Uaa) validateAudience(ctx echo.Context) {
	user := ctx.Get("user").(*jwt.Token)
	if !uaa.validateAudClaim(user.Claims.(jwt.MapClaims)["aud"].([]interface{})) {
		errorText := fmt.Sprintf("valid JWT received but missing %s audience claim, closing connection", uaa.Config.ExpectedAudienceClaim)
		logger.Log.Error(errorText)
		ctx.Error(echo.NewHTTPError(http.StatusUnauthorized, errorText))
	}
}

func (uaa *Uaa) validateAudClaim(claims []interface{}) bool {
	for _, claim := range claims {
		if strings.Contains(claim.(string), uaa.Config.ExpectedAudienceClaim) {
			return true
		}
	}
	return false
}
