package utils

import "github.com/gin-gonic/gin"

func JsonError(code int, error string, c *gin.Context) {
	c.JSON(code, gin.H{
		"error": error,
	})
}

type Json map[string]interface{}
