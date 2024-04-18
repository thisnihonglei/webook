package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"webook/internal/domain"
	"webook/internal/repository"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户不存在或密码错误")
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Edit(ctx context.Context, user domain.User) error
	FindById(ctx *gin.Context, id int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWeChat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByName(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *userService) Edit(ctx context.Context, user domain.User) error {
	return svc.repo.Edit(ctx, user)
}

func (svc *userService) FindById(ctx *gin.Context, id int64) (domain.User, error) {
	return svc.repo.FindById(ctx, id)
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 查询用户是否存在
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		// 两种情况
		// 1.err=nil u是可用的
		// 2.err!=nil 系统错误
		return u, err
	}
	//用户没有找到
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	// 两种可能 唯一索引冲突（phone）
	// 一种是err！=nil
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) FindOrCreateByWeChat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	// 查询用户是否存在
	u, err := svc.repo.FindByWeChat(ctx, wechatInfo.OpenId)
	if err != repository.ErrUserNotFound {
		// 两种情况
		// 1.err=nil u是可用的
		// 2.err!=nil 系统错误
		return u, err
	}
	//用户没有找到
	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: wechatInfo,
	})

	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	return svc.repo.FindByWeChat(ctx, wechatInfo.OpenId)
}
