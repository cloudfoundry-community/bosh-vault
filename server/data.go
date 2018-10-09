package server

import (
	"fmt"
	"github.com/labstack/echo"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
	"io/ioutil"
	"net/http"
)

func dataPostHandler(ctx echo.Context) error {
	context := ctx.(*ConfigurationContext)
	requestBody, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err))
		return err
	}

	context.Log.Debugf("request: %s", requestBody)

	credentialRequest, err := vcfcsTypes.ParseGenericCredentialRequest(requestBody)
	if err != nil {
		context.Log.Error("request error: ", err)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err.Error()))
		return err
	}

	credentialType := credentialRequest.CredentialType()
	ok := credentialRequest.Validate()
	if !ok {
		context.Log.Error("invalid credential request for ", credentialType)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid credential request for %s", credentialType)))
		return err
	}

	context.Log.Debugf("attempting to generate %s", credentialType)

	credential, err := credentialRequest.Generate()
	if err != nil {
		context.Log.Error(err)
		ctx.Error(echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem generating %s: %s", credentialType, err)))
	}

	return ctx.JSON(http.StatusOK, &credential)
}
