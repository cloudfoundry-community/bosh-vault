package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/zipcar/vault-cfcs/config"
	"github.com/zipcar/vault-cfcs/logger"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const UaaAuthScheme = "bearer" // uaa doesn't use "Bearer" (the JWT default), but "bearer"
const UaaExpectedAudience = "config_server"

type UaaClient struct {
	CheckTokenEndpoint string
	TokenKeyEndpoint   string
	Username           string
	Password           string
	httpClient         *http.Client
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

func GetUaaClient(vcfcsConfig config.Configuration) UaaClient {

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		logger.Log.Error("problem reading system cert pool, if no UAA CA cert was passed in the config expect TLS errors")
		rootCAs = x509.NewCertPool()
	}

	if vcfcsConfig.Uaa.Ca != "" {
		certs, err := ioutil.ReadFile(vcfcsConfig.Uaa.Ca)
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", vcfcsConfig.Uaa.Ca, err)
		}

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: vcfcsConfig.Uaa.SkipVerify,
		RootCAs:            rootCAs,
	}

	// Setup a custom transport that trusts our UAA Ca as well as the system's trusted certs
	customTransport := &http.Transport{TLSClientConfig: tlsConfig}

	return UaaClient{
		Username:           vcfcsConfig.Uaa.Username,
		Password:           vcfcsConfig.Uaa.Password,
		CheckTokenEndpoint: fmt.Sprintf("%s/check_token", vcfcsConfig.Uaa.Address),
		TokenKeyEndpoint:   fmt.Sprintf("%s/token_key", vcfcsConfig.Uaa.Address),
		httpClient: &http.Client{
			Timeout:   time.Second * time.Duration(vcfcsConfig.Uaa.Timeout),
			Transport: customTransport,
		},
	}
}

// todo: refactor so this method is callled and caching signing key information, updating on a timer
func (uaa *UaaClient) GetTokenSigningInfo() (TokenKeyResponse, error) {
	var signingKeyResp TokenKeyResponse

	resp, err := uaa.httpClient.Get(uaa.TokenKeyEndpoint)
	if err != nil {
		return signingKeyResp, err
	}
	if resp.StatusCode != http.StatusOK {
		return signingKeyResp, fmt.Errorf("received a status code %v when requesting token signing info", resp.Status)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return signingKeyResp, err
	}
	err = json.Unmarshal(responseBody, &signingKeyResp)
	if err != nil {
		return signingKeyResp, err
	}
	return signingKeyResp, nil
}

func (uaa *UaaClient) AuthMiddleware() echo.MiddlewareFunc {
	// todo: refactor to cache this signing info and update with a timer channel instead of crashing when UAA is down
	uaaKeyData, err := uaa.GetTokenSigningInfo()
	if err != nil {
		// connection lost with UAA
		logger.Log.Fatal(err)
	}

	publicKey, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(uaaKeyData.Value))

	// The JWT middleware will handle basic authentication, our success handler does broad based audience claim authorization
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    publicKey,
		SigningMethod: uaaKeyData.Alg,
		AuthScheme:    UaaAuthScheme,
		// JWT middleware handles basic validity checks, this successhandler is our custom audience check since UAA
		// returns a []string for the aud claim so single users can access multiple resources, the consequence is we can't
		// use the built in methods of the JWT middleware to validate the audience for us since we don't know what additional
		// audiences a given user may have
		SuccessHandler: uaa.validateUaaAudience,
	})
}

func (uaa *UaaClient) validateUaaAudience(ctx echo.Context) {
	user := ctx.Get("user").(*jwt.Token)
	if !validateAudClaim(user.Claims.(jwt.MapClaims)["aud"].([]interface{})) {
		errorText := fmt.Sprintf("valid JWT received but missing %s audience claim, closing connection", UaaExpectedAudience)
		logger.Log.Error(errorText)
		ctx.Error(echo.NewHTTPError(http.StatusUnauthorized, errorText))
	} else {
		logger.Log.Debugf("token claim audience contains %s audience claim, allow connection", UaaExpectedAudience)
	}
}

func validateAudClaim(claims []interface{}) bool {
	for _, claim := range claims {
		if strings.Contains(claim.(string), UaaExpectedAudience) {
			return true
		}
	}
	return false
}
