package domain

type User struct {
	Id       int64
	Email    string
	Password string
	NickName string
	Birthday int64
	AboutMe  string
	Phone    string

	WechatInfo WechatInfo
}
