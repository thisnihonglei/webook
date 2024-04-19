package middleware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			return
		}

		tokenStr := web.ExtractToken(ctx)
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//expires := uc.ExpiresAt
		//if expires.Sub(time.Now()) < 50*time.Second {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
		//	tokenStr, err = token.SignedString(web.JWTKey)
		//	ctx.Header("x-jwt-token", tokenStr)
		//	if err != nil {
		//		fmt.Println(err)
		//	}
		//}
		ctx.Set("user", uc)
	}
}
