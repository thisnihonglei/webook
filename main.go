package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
)

func main() {
	initViperRemote()
	server := InitWebServer()
	server.Run(viper.GetString("server.port"))
}

func initViper() {
	configFile := pflag.String("config", "config/dev.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*configFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperWatch() {
	configFile := pflag.String("config", "config/dev.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*configFile)
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println(viper.GetString("test.key"))
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println("远程配置中心发生变更")
	})
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
	//go func() {
	//	for {
	//		err = viper.WatchRemoteConfig()
	//		if err != nil {
	//			panic(err)
	//		}
	//		log.Println("watch", viper.GetString("test.key"))
	//		time.Sleep(time.Second * 3)
	//	}
	//
	//}()
}

//func initUser(db *gorm.DB, redisClient redis.Cmdable, codeSvc service.CodeService, server *gin.Engine) {
//	userCache := cache.NewUserCache(redisClient)
//	ud := dao.NewUserDAO(db)
//	ur := repository.NewCachedUserRepository(ud, userCache)
//	us := service.NewUserService(ur)
//	hdl := web.NewUserHandler(us, codeSvc)
//	hdl.RegisterRoutes(server)
//}

//func initSMSService() sms.Service {
//	return localsms.NewService()
//}
//
//func initCodeSvc(redisClient redis.Cmdable) service.CodeService {
//	cc := cache.NewCodeCache(redisClient)
//	cRepo := repository.NewCodeRepository(cc)
//	cSms := initSMSService()
//	return service.NewCodeService(cRepo, cSms)
//}
//
//func initWebServer() *gin.Engine {
//	server := gin.Default()
//	server.Use(cors.New(cors.Config{
//		AllowCredentials: true,
//		AllowedHeaders:   []string{"Content-Type", "Authorization"},
//		ExposedHeaders:   []string{"x-jwt-token"},
//		AllowOriginFunc: func(origin string) bool {
//			if strings.HasPrefix(origin, "http://localhost") {
//				return true
//			}
//			return strings.Contains(origin, "http:127.0.0.1")
//		},
//		MaxAge: 12 * time.Hour,
//	}))
//	//redisClient := redis.NewClient(&redis.Options{
//	//	Addr: "127.0.0.1:6379",
//	//})
//	//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
//
//	useJWT(server)
//	//useSession(server)
//
//	return server
//}
//
//func useJWT(server *gin.Engine) {
//	login := &middleware.LoginJWTMiddlewareBuilder{}
//	server.Use(login.CheckLogin())
//}

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
