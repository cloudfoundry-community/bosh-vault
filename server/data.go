package server

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
	"github.com/zipcar/vault-cfcs/vault"
	"io/ioutil"
	"net/http"
)

func dataGetByNameHandler(ctx echo.Context) error {
	context := ctx.(*VcfcsContext)
	name := ctx.QueryParam("name")
	if name == "" {
		// this should never happen because of echo's router
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "name query param not passed to data?name handler"))
		return errors.New("name query param not passed to data?name handler")
	}
	context.Log.Debugf("request to GET /v1/data?name=%s", name)
	secretResponse, err := vault.FetchSecretByName(name)
	if err != nil {
		context.Log.Errorf("problem fetching secret by name: %s %s", name, err)
		return err
	}
	return ctx.JSON(http.StatusOK, secretResponse)
}

func dataGetByIdHandler(ctx echo.Context) error {
	context := ctx.(*VcfcsContext)
	id := ctx.Param("id")
	if id == "" {
		// this should never happen because of echo's router
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "id uri param not passed to data/:id handler"))
		return errors.New("id uri param not passed to data/:id handler")
	}
	context.Log.Debugf("request to /v1/data/%s", id)
	vaultSecretResponse, err := vault.FetchSecretById(id)
	if err != nil {
		context.Log.Errorf("problem fetching secret by id: %s %s", id, err)
		return err
	}

	secretResp := vcfcsTypes.ParseSecretResponse(vaultSecretResponse)

	return ctx.JSON(http.StatusOK, secretResp)
}

func dataPostHandler(ctx echo.Context) error {
	context := ctx.(*VcfcsContext)
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

	return ctx.JSON(http.StatusCreated, &credential)
}

func dataDeleteHandler(ctx echo.Context) error {
	context := ctx.(*VcfcsContext)
	name := ctx.QueryParam("name")
	if name == "" {
		// this should never happen because of echo's router
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "name query param not passed to data?name handler"))
		return errors.New("name query param not passed to data?name handler")
	}
	context.Log.Debugf("request to DELETE /v1/data?name=%s", name)
	err := vault.DeleteSecretByName(name)
	if err != nil {
		context.Log.Errorf("problem deleting secret by name: %s %s", name, err)
		return err
	}
	return ctx.NoContent(http.StatusNoContent)
}
