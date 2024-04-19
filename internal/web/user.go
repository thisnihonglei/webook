package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	bizLogin             = "login"
)

type UserHandler struct {
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	svc              service.UserService
	codeSvc          service.CodeService
	jwtHandler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:              svc,
		codeSvc:          codeSvc,
		jwtHandler:       NewJWTHandler(),
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("users")
	ug.POST("/signup", h.SignUp)
	//ug.POST("/login", h.Login)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)

	ug.GET("/refresh_token", h.RefreshToken)

	//手机验证码登录相关功能
	ug.POST("/login_sms/code/send", h.SendLoginSMSCode)
	ug.POST("/login_sms", h.LoginSMS)

}

func (h *UserHandler) LoginSMS(ctx *gin.Context) {
	type codeReq struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req codeReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := h.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}

	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码错误，请重新输入",
		})
		return
	}
	u, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = h.setRefreshToken(ctx, u.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	h.setJWTToken(ctx, u.Id)
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
}

func (h *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type SMSReq struct {
		Phone string `json:"phone"`
	}

	var req SMSReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "请输入手机号码",
		})
		return
	}

	err := h.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		//补日志的
	}
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式不正确")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码输入的不一样")
		return
	}

	isPassword, err := h.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符、并且长度不能小于8位")
		return
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		err = h.setRefreshToken(ctx, u.Id)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		h.setJWTToken(ctx, u.Id)
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码错误")
	default:
		ctx.String(http.StatusOK, "系统错误")

	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		session := sessions.Default(ctx)
		session.Set("userId", u.Id)
		session.Options(sessions.Options{
			MaxAge: 900,
		})
		err = session.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码错误")
	default:
		ctx.String(http.StatusOK, "系统错误")

	}
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		NickName string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"about_me"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	t, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		fmt.Println("解析日期失败:", err)
		return
	}

	// 使用 time.Unix() 函数将 time.Time 对象转换为时间戳（Unix 时间）
	timestamp := t.Unix()
	//sess := sessions.Default(ctx)
	//id := sess.Get("userId").(int64)

	userData, exists := ctx.Get("user")
	if !exists {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	uc, ok := userData.(UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	id := uc.Uid

	err = h.svc.Edit(ctx, domain.User{
		Id:       id,
		NickName: req.NickName,
		Birthday: timestamp,
		AboutMe:  req.AboutMe,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "更新成功")
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Id       int64
		Email    string
		NickName string
		Birthday string
		AboutMe  string
	}
	//sess := sessions.Default(ctx)
	//id := sess.Get("userId").(int64)

	userData, exists := ctx.Get("user")
	if !exists {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	uc, ok := userData.(UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	id := uc.Uid

	u, err := h.svc.FindById(ctx, id)
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 将时间戳转换为Time类型
	timeValue := time.Unix(u.Birthday, 0)

	// 将Time类型格式化为字符串
	dateString := timeValue.Format("2006-01-02")
	ctx.JSON(http.StatusOK, Profile{
		Id:       id,
		Email:    u.Email,
		NickName: u.NickName,
		Birthday: dateString,
		AboutMe:  u.AboutMe,
	})
}

func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	tokenStr := ExtractToken(ctx)
	var rc RefreshClaims

	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return h.RefreshKey, nil
	})

	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	h.setJWTToken(ctx, rc.Uid)
	ctx.JSON(http.StatusOK, "刷新成功")
}
