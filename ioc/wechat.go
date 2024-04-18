package ioc

import (
	"webook/internal/service/auth2/wechat"
)

func InitWechatService() wechat.Service {
	//appId, ok := os.LookupEnv("WECHAT_APP_ID")
	//if !ok {
	//	panic("找不到环境变量 WECHAT_APP_ID")
	//}

	//appSecret, ok := os.LookupEnv("WECHAT_APP_Secret")
	//if !ok {
	//	panic("找不到环境变量 WECHAT_APP_Secret")
	//}

	appId := "123456"
	appSecret := "theSecret"

	return wechat.NewService(appId, appSecret)
}
