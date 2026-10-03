package response

import (
	"net/http"

	"calendar-booking/internal/common/validator"

	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func Error(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, ErrorBody{Error: msg})
}

// BindError: 422 for validation failures, 400 for malformed JSON.
func BindError(c *gin.Context, err error) {
	if details := validator.FormatErrors(err); details != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, ErrorBody{
			Error:   "validation failed",
			Details: details,
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
}
