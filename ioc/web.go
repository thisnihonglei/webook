package ioc

import (
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/internal/web/middleware"
	"webook/pkg/ginx/middleware/ratelimit"
	"webook/pkg/limiter"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, wechat *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	wechat.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable, hdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			ExposedHeaders:   []string{"x-jwt-token", "x-refresh-token"},
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					return true
				}
				return strings.Contains(origin, "http:127.0.0.1")
			},
			MaxAge: 12 * time.Hour,
		}),

		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
		middleware.NewLoginJWTMiddlewareBuilder(hdl).CheckLogin(),
	}
}
