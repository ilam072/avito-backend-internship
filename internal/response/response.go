package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Err struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error Err `json:"error"`
}

func Error(c *gin.Context, status int, code string, message string) {
	errorResponse := ErrorResponse{Error: Err{
		Code:    code,
		Message: message,
	}}

	c.JSON(status, errorResponse)
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func Conflict(c *gin.Context, code string, message string) {
	Error(c, http.StatusConflict, code, message)
}

func NotFound(c *gin.Context) {
	Error(c, http.StatusConflict, "NOT_FOUND", "resource not found")
}

func InternalServerError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error, try again later")
}
