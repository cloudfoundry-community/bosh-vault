package server

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"github.com/zipcar/bosh-vault/store"
	bvTypes "github.com/zipcar/bosh-vault/types"
	"io/ioutil"
	"net/http"
)

func dataGetByNameHandler(ctx echo.Context) error {
	context := ctx.(*BvContext)
	name := ctx.QueryParam("name")
	if name == "" {
		// this should never happen because of echo's router
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "name query param not passed to data?name handler"))
		return errors.New("name query param not passed to data?name handler")
	}
	context.Log.Debugf("request to GET /v1/data?name=%s", name)
	secretResponses, err := store.GetAllByName(name)
	if err != nil {
		context.Log.Errorf("problem fetching secret by name: %s %s", name, err)
		return err
	}

	responseData := make([]*store.SecretResponse, 0)
	for _, secret := range secretResponses {
		responseData = append(responseData, bvTypes.ParseSecretResponse(secret))
	}

	return ctx.JSON(http.StatusOK, struct {
		Data []*store.SecretResponse `json:"data"`
	}{
		Data: responseData,
	})
}

func dataGetByIdHandler(ctx echo.Context) error {
	context := ctx.(*BvContext)
	id := ctx.Param("id")
	if id == "" {
		// this should never happen because of echo's router
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "id uri param not passed to data/:id handler"))
		return errors.New("id uri param not passed to data/:id handler")
	}
	context.Log.Debugf("request to /v1/data/%s", id)
	vaultSecretResponse, err := store.GetById(id)
	if err != nil {
		context.Log.Errorf("problem fetching secret by id: %s %s", id, err)
		return err
	}

	secretResp := bvTypes.ParseSecretResponse(vaultSecretResponse)

	return ctx.JSON(http.StatusOK, secretResp)
}

func dataPostHandler(ctx echo.Context) error {
	context := ctx.(*BvContext)
	requestBody, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err))
		return err
	}

	context.Log.Debugf("request: %s", requestBody)

	credentialRequest, err := bvTypes.ParseCredentialGenerationRequest(requestBody)
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
	context := ctx.(*BvContext)
	name := ctx.QueryParam("name")
	if name == "" {
		// this should never happen because of echo's router
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "name query param not passed to data?name handler"))
		return errors.New("name query param not passed to data?name handler")
	}
	context.Log.Debugf("request to DELETE /v1/data?name=%s", name)
	err := store.DeleteByName(name)
	if err != nil {
		context.Log.Errorf("problem deleting secret by name: %s %s", name, err)
		return err
	}
	return ctx.NoContent(http.StatusNoContent)
}

func dataPutHandler(ctx echo.Context) error {
	context := ctx.(*BvContext)
	requestBody, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err))
		return err
	}

	context.Log.Debugf("request: %s", requestBody)
	setRequest, err := bvTypes.ParseCredentialSetRequest(requestBody)
	if err != nil {
		context.Log.Error("request error: ", err)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err.Error()))
		return err
	}
	response, err := setRequest.Record.Store(setRequest.Name)
	if err != nil {
		context.Log.Error("server error: ", err)
		ctx.Error(echo.NewHTTPError(http.StatusInternalServerError, err.Error()))
		return err
	}
	return ctx.JSON(http.StatusOK, &response)
}
