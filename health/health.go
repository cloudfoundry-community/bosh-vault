package health

import (
	"github.com/labstack/echo"
	"net/http"
)

const HealthCheckUri = "/v1/health"

func HealthCheckHandler(ctx echo.Context) error {
	// todo: Other than being up and able to respond we should have healthy checks, like confirming Vault connection
	return ctx.JSON(http.StatusOK, &map[string]interface{}{
		"status":      http.StatusOK,
		"status_text": http.StatusText(http.StatusOK),
	})
}
