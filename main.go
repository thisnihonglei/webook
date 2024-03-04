package main

import (
	//"github.com/gin-contrib/sessions"
	//"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/sms"
	"webook/internal/service/sms/localsms"
	"webook/internal/web"
	"webook/internal/web/middleware"
)

func main() {

	db := initDB()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	server := initWebServer()
	codeSvc := initCodeSvc(redisClient)
	initUser(db, redisClient, codeSvc, server)
	//server := gin.Default()
	//server.GET("/hello", func(ctx *gin.Context) {
	//	ctx.String(http.StatusOK, "hello,Kubernetes 启动成功了！")
	//})
	server.Run(":8080")
}

func initUser(db *gorm.DB, redisClient redis.Cmdable, codeSvc *service.CodeService, server *gin.Engine) {
	userCache := cache.NewUserCache(redisClient)
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud, userCache)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us, codeSvc)
	hdl.RegisterRoutes(server)
}

func initSMSService() sms.Service {
	return localsms.NewService()
}

func initCodeSvc(redisClient redis.Cmdable) *service.CodeService {
	cc := cache.NewCodeCache(redisClient)
	cRepo := repository.NewCodeRepository(cc)
	cSms := initSMSService()
	return service.NewCodeService(cRepo, cSms)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:3306)/webook?charset=utf8mb4&parseTime=True&loc=Local"))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"x-jwt-token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "http:127.0.0.1")
		},
		MaxAge: 12 * time.Hour,
	}))
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: "127.0.0.1:6379",
	//})
	//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	useJWT(server)
	//useSession(server)

	return server
}

func useJWT(server *gin.Engine) {
	login := &middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

//func useSession(server *gin.Engine) {
//	//store := cookie.NewStore([]byte("secret"))
//	store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
//		[]byte("cgWrzQrzH2tfJngYC59iuqh3Dix235FX"),
//		[]byte("h1R8JhnYVKVnajhK9HsYhwTmM1FNo7bS"),
//	)
//	if err != nil {
//		panic(err)
//	}
//	server.Use(sessions.Sessions("ssid", store))
//	login := &middleware.LoginMiddlewareBuilder{}
//	server.Use(login.CheckLogin())
//}
