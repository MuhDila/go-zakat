package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response adalah format standar response API
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Success mengirimkan response sukses standar
func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error mengirimkan response error standar
func Error(c *gin.Context, code int, message string, errors interface{}) {
	c.JSON(code, Response{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// ValidationError mengirimkan response error validasi
func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

// BadRequest mengirimkan response 400 Bad Request
func BadRequest(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusBadRequest, message, errors)
}

// Unauthorized mengirimkan response 401 Unauthorized
func Unauthorized(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusUnauthorized, message, errors)
}

// InternalServerError mengirimkan response 500 Internal Server Error
func InternalServerError(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusInternalServerError, message, errors)
}
