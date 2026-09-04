package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIResponse is the standard response envelope for all API endpoints.
// Every response — success or error — uses this format.
// WHY: Consistent format lets frontend handle responses uniformly.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    Meta        `json:"meta"`
}

// APIError contains error details.
// Code is machine-readable. Message is human-readable.
// Details contains field-level validation errors.
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta contains request metadata.
type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

// OK sends a 200 success response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

// Created sends a 201 created response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, code, message string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// ValidationError sends a 400 with field-level errors.
func ValidationError(c *gin.Context, details map[string]string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: details,
		},
		Meta: buildMeta(c),
	})
}

// Unauthorized sends a 401 error response.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// Forbidden sends a 403 error response.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "FORBIDDEN",
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "NOT_FOUND",
			Message: resource + " not found",
		},
		Meta: buildMeta(c),
	})
}

// Conflict sends a 409 error response.
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "CONFLICT",
			Message: message,
		},
		Meta: buildMeta(c),
	})
}

// InternalError sends a 500 error response.
// NEVER expose internal error details to client.
// Log the actual error separately.
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred. Please try again.",
		},
		Meta: buildMeta(c),
	})
}

// TooManyRequests sends a 429 rate limit response.
func TooManyRequests(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "RATE_LIMIT_EXCEEDED",
			Message: "Too many requests. Please slow down.",
		},
		Meta: buildMeta(c),
	})
}

// buildMeta creates response metadata.
// RequestID is set by middleware — falls back to new UUID if missing.
func buildMeta(c *gin.Context) Meta {
	requestID, exists := c.Get("request_id")
	if !exists {
		requestID = uuid.New().String()
	}
	return Meta{
		RequestID: requestID.(string),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}
