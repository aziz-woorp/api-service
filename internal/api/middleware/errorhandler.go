package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// Only run if there are errors to handle
		if len(c.Errors) > 0 {
			// Use the last error
			err := c.Errors.Last()
			status := http.StatusInternalServerError
			if err.Type == gin.ErrorTypePublic {
				status = c.Writer.Status()
				if status < 400 {
					status = http.StatusBadRequest
				}
			}
			c.JSON(status, gin.H{
				"error":      err.Error(),
				"request_id": c.GetString("request_id"),
			})
		}
	}
}
