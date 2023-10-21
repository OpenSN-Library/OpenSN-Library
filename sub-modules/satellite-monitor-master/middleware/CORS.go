package middleware

import "github.com/gin-gonic/gin"

func CORS(context *gin.Context) {
	context.Header("Access-Control-Allow-Origin", "*")
	context.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
	context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	context.Header("Access-Control-Allow-Credentials", "true")
	context.Next()
}
