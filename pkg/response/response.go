package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
)

// APIResponse is the envelope every endpoint returns.
// Success:  { "success": true,  "data": {...} }
// Failure:  { "success": false, "error": {"code": 400, "message": "..."} }
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OK sends 200 with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data})
}

// Created sends 201 with data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: data})
}

// Accepted sends 202 with data.
// Used by Gateway after publishing order to Kafka.
func Accepted(c *gin.Context, data interface{}) {
	c.JSON(http.StatusAccepted, APIResponse{Success: true, Data: data})
}

// Fail sends the correct HTTP status mapped from AppError code.
func Fail(c *gin.Context, err *apperr.AppError) {
	c.JSON(mapCodeToHTTP(err.Code), APIResponse{
		Success: false,
		Error:   &ErrorBody{Code: int(err.Code), Message: err.Message},
	})
}

func mapCodeToHTTP(code apperr.Code) int {
	switch code {
	case apperr.CodeInvalidInput:
		return http.StatusBadRequest
	case apperr.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperr.CodeForbidden:
		return http.StatusForbidden
	case apperr.CodeNotFound:
		return http.StatusNotFound
	case apperr.CodeConflict:
		return http.StatusConflict
	case apperr.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}