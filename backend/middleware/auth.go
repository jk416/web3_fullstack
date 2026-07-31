package middleware

import (
	"net/http"
	"strings"
	"web3-wallet-exchange/global"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := global.Token.VerifyToken(token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Here you would typically validate the token
		c.Set("userID", claims.UserID)
		c.Next()
	}

}
