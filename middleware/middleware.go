package middleware

import (
	"net/http"

	token "github.com/amanraghuvanshi/ecombackend/tokens"
	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		//  we are expecting the token in the request header, so we will get the access to the token
		ClientToken := c.Request.Header.Get("token")
		if ClientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "No Authorization header Provided"})
			c.Abort()
			return
		}
		claims, err := token.ValidateToken(ClientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("uid", claims.UID)
		// this tells us that the process can do the transition to the next step.
		c.Next()
	}
}
