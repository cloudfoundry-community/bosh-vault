package server

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"github.com/zipcar/bosh-vault/secret"
	"github.com/zipcar/bosh-vault/types"
	"io/ioutil"
	"net/http"
)

func healthCheckHandler(ctx echo.Context) error {
	context := ctx.(*BvContext)
	// todo: Should this also verify that no DEBUG properties are set and return a different status code if so?
	if context.Store.Healthy() {
		return ctx.JSON(http.StatusOK, &map[string]interface{}{
			"status":      http.StatusOK,
			"status_text": http.StatusText(http.StatusOK),
		})
	} else {
		return ctx.JSON(http.StatusInternalServerError, &map[string]interface{}{
			"status":      http.StatusInternalServerError,
			"status_text": fmt.Sprintf("%s your backend store is unhealthy, has it been initialized and unsealed?", http.StatusText(http.StatusInternalServerError)),
		})
	}

}

func dataGetByNameHandler(ctx echo.Context) error {
	context := ctx.(*BvContext)
	name := ctx.QueryParam("name")
	if name == "" {
		// this should never happen because of echo's router
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "name query param not passed to data?name handler"))
		return errors.New("name query param not passed to data?name handler")
	}
	context.Log.Debugf("request to GET %s?name=%s", dataUri, name)

	secretResponses, err := context.Store.GetByName(name)
	if err != nil {
		ctx.Error(echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("problem fetching secret by name: %s %s", name, err)))
		return err
	}

	// Check to see if the value is a flat string value and don't nest it if that's the case (password type, for example)
	// todo: look at checking for integer values too
	responseData := make([]secret.Secret, 0)
	for _, sr := range secretResponses {
		valString, ok := sr.Value.(map[string]interface{})["value"].(string)
		if ok {
			sr.Value = valString
		}
		responseData = append(responseData, sr)
	}

	return ctx.JSON(http.StatusOK, struct {
		Data []secret.Secret `json:"data"`
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
	context.Log.Debugf("request to %s/%s", dataUri, id)
	vaultSecretResponse, err := context.Store.GetById(id)
	if err != nil {
		ctx.Error(echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("problem fetching secret by id: %s %s", id, err)))
		return err
	}
	valString, ok := vaultSecretResponse.Value.(map[string]interface{})["value"].(string)
	if ok {
		vaultSecretResponse.Value = valString
	}
	return ctx.JSON(http.StatusOK, vaultSecretResponse)
}

func dataPostHandler(ctx echo.Context) error {
	context := ctx.(*BvContext)
	requestBody, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err))
		return err
	}

	context.Log.Debugf("request: %s", requestBody)

	credentialRequest, noOverrideMode, err := types.ParseCredentialGenerationRequest(requestBody)
	if err != nil {
		context.Log.Error("request error: ", err)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err.Error()))
		return err
	}

	if noOverrideMode && context.Store.Exists(credentialRequest.CredentialName()) {
		latest, err := context.Store.GetLatestByName(credentialRequest.CredentialName())
		if err != nil {
			context.Log.Errorf("problem getting latest in no-override mode: %s", err)
			ctx.Error(echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem getting latest in no-override mode: %s", err)))
		}
		return ctx.JSON(http.StatusOK, &latest)
	}

	credentialType := credentialRequest.CredentialType()
	ok := credentialRequest.Validate()
	if !ok {
		context.Log.Error("invalid credential request for ", credentialType)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid credential request for %s", credentialType)))
		return err
	}

	context.Log.Debugf("attempting to generate %s", credentialType)

	credential, err := credentialRequest.Generate(context.Store)
	if err != nil {
		context.Log.Error(err)
		ctx.Error(echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem generating %s: %s", credentialType, err)))
	}

	credentialResponse, err := credential.Store(context.Store, credentialRequest.CredentialName())
	if err != nil {
		context.Log.Error(err)
		context.Error(echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem storing %s: %s", credentialType, credentialRequest.CredentialName())))
	}

	return ctx.JSON(http.StatusCreated, &credentialResponse)
}

func dataDeleteHandler(ctx echo.Context) error {
	context := ctx.(*BvContext)
	name := ctx.QueryParam("name")
	if name == "" {
		// this should never happen because of echo's router
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "name query param not passed to data?name handler"))
		return errors.New("name query param not passed to data?name handler")
	}
	context.Log.Debugf("request to DELETE %s?name=%s", dataUri, name)
	err := context.Store.DeleteByName(name)
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
	setRequest, err := types.ParseCredentialSetRequest(requestBody)
	if err != nil {
		context.Log.Error("request error: ", err)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err.Error()))
		return err
	}
	response, err := setRequest.Record.Store(context.Store, setRequest.Name)
	if err != nil {
		context.Log.Error("server error: ", err)
		ctx.Error(echo.NewHTTPError(http.StatusInternalServerError, err.Error()))
		return err
	}
	return ctx.JSON(http.StatusOK, &response)
}
