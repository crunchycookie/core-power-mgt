package serviceerror

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, err := range c.Errors {
			id := uuid.New()
			log.Println(fmt.Sprintf("error id: %s - ", id.String()), err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Something went wrong. Check server logs for id: %s.", id.String())})
		}
	}
}
