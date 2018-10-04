package server

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
	"io/ioutil"
	"net/http"
)

func dataPostHandler(ctx echo.Context) error {
	ctx.Logger().SetLevel(log.INFO)
	requestBody, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err))
		return err
	}

	credentialRequest, err := vcfcsTypes.ParseGenericCredentialRequest(requestBody)
	if err != nil {
		ctx.Logger().Error(err)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err.Error()))
	}

	credentialType := credentialRequest.CredentialType()
	ok := credentialRequest.Validate()
	if !ok {
		ctx.Logger().Error("Invalid credential request for ", credentialType)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "Invalid credential request for ", credentialType))
	}

	ctx.Logger().Info(fmt.Sprintf("Attempting to generate %s", credentialType))

	err = credentialRequest.Generate()
	if err != nil {
		ctx.Logger().Error(err)
		ctx.Error(echo.NewHTTPError(http.StatusInternalServerError, "Problem generating ", credentialType, err))
	}

	return ctx.JSON(http.StatusOK, &map[string]interface{}{
		"status":      http.StatusOK,
		"status_text": http.StatusText(http.StatusOK),
	})
}
