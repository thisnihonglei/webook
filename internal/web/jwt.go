package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

type jwtHandler struct {
	signingMethod jwt.SigningMethod
	RefreshKey    []byte
}

func NewJWTHandler() jwtHandler {
	return jwtHandler{
		signingMethod: jwt.SigningMethodHS512,
		RefreshKey:    []byte("cgWrzQrzH2tfJngYC59iuqh3Dix246FE"),
	}
}

func (h *jwtHandler) setJWTToken(ctx *gin.Context, uid int64) {
	uc := UserClaims{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, tokenErr := token.SignedString(JWTKey)
	if tokenErr != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	ctx.Header("x-jwt-token", tokenStr)

}

func (h *jwtHandler) setRefreshToken(ctx *gin.Context, uid int64) error {
	uc := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid: uid,
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, tokenErr := token.SignedString(h.RefreshKey)
	if tokenErr != nil {
		return tokenErr
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

var JWTKey = []byte("cgWrzQrzH2tfJngYC59iuqh3Dix246FX")

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid int64
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid int64
}

func ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		// 没有登录，没有token，没有Authorization这个头部
		return authCode
	}
	seg := strings.Split(authCode, " ")
	if len(seg) != 2 {
		// 没登录 Authorization是乱传的
		return ""
	}
	return seg[1]
}
