package gateway

import (
	"strings"

	"github.com/gin-gonic/gin"
	gRPCauth "github.com/mohmdsaalim/EngineX/api/gen/gRPCauth"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
	"github.com/mohmdsaalim/EngineX/pkg/response"
)

func AuthMiddleware(authClient gRPCauth.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == ""{
			response.Fail(c, apperr.New(apperr.CodeUnauthorized, "invalid authorization header required"))
			c.Abort()
			return 
	}
	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 || parts[0] != "Bearer"{
		response.Fail(c, apperr.New(apperr.CodeUnauthorized, "invalid authorixation format"))
		c.Abort()
		return 
	}

	resp, err := authClient.ValidateToken(c.Request.Context(), 
		&gRPCauth.ValidateTokenRequest{Token: parts[1]})

		if err != nil || !resp.Valid{
			response.Fail(c, apperr.New(apperr.CodeUnauthorized, " invalid or expires token"))
			c.Abort()
			return 
		}

		c.Set("userID", resp.UserId)
		c.Set("email", resp.Email)
		c.Next()
}}