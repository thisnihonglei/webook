//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

var thirdPartySet = wire.NewSet(InitDB, InitRedis,
	InitLogger)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		thirdPartySet,
		// Dao 部分
		dao.NewUserDAO,
		dao.NewArticleGORMDAO,
		// cache 部分
		cache.NewCodeCache, cache.NewUserCache,

		// Repository 部分
		repository.NewCachedUserRepository, repository.NewCodeRepository, repository.NewCachedArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// Handler 部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		web.NewArticleHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		dao.NewArticleGORMDAO,
		repository.NewCachedArticleRepository,
		web.NewArticleHandler,
		service.NewArticleService)
	return &web.ArticleHandler{}
}
