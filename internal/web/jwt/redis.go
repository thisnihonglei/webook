package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
	"time"
)

type RedisJWTHandler struct {
	client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rtExpiration  time.Duration
}

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		client:        client,
		signingMethod: jwt.SigningMethodHS512,
		rtExpiration:  time.Hour * 24 * 7,
	}
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	logout, err := h.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if logout > 0 {
		return errors.New("token 无效")
	}

	return nil
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
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

var _ Handler = &RedisJWTHandler{}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = h.setRefreshToken(ctx, uid, ssid)
	return err
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, tokenErr := token.SignedString(JWTKey)
	if tokenErr != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rtExpiration)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, tokenErr := token.SignedString(RefreshKey)
	if tokenErr != nil {
		return tokenErr
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

// ClearToken 清除token
func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	return h.client.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid), "", h.rtExpiration).Err()
}

var JWTKey = []byte("cgWrzQrzH2tfJngYC59iuqh3Dix246FX")
var RefreshKey = []byte("cgWrzQrzH2tfJngYC59iuqh3Dix246FR")

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}
