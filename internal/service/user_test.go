package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
)

func TestPasswordEncrypt(t *testing.T) {
	password := []byte("Hello#world123")
	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	println(string(encrypted))
}

func Test_userService_Login(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		ctx      context.Context
		email    string
		password string

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByName(gomock.Any(), "123456@qq.com").Return(domain.User{
					Email: "123456@qq.com",
					// 加密后的密码
					Password: "$2a$10$N.edWE4zAEdb33BlrZiGe.R/yxjJSY2yhYIV2lWOstwxyGeLbMcuW",
					Phone:    "15811111111",
				}, nil)
				return repo
			},
			email:    "123456@qq.com",
			password: "Hello#world123",

			wantUser: domain.User{
				Email:    "123456@qq.com",
				Password: "$2a$10$N.edWE4zAEdb33BlrZiGe.R/yxjJSY2yhYIV2lWOstwxyGeLbMcuW",
				Phone:    "15811111111",
			},
			wantErr: nil,
		},
		{
			name: "用户未找到",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByName(gomock.Any(), "123456@qq.com").Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123456@qq.com",
			password: "Hello#world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByName(gomock.Any(), "123456@qq.com").Return(domain.User{}, errors.New("DB错误"))
				return repo
			},
			email:    "123456@qq.com",
			password: "Hello#world123",

			wantUser: domain.User{},
			wantErr:  errors.New("DB错误"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByName(gomock.Any(), "123456@qq.com").Return(domain.User{
					Email: "123456@qq.com",
					// 加密后的密码
					Password: "xxxxxx",
					Phone:    "15811111111",
				}, nil)
				return repo
			},
			email:    "123456@qq.com",
			password: "Hello#world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewUserService(repo)
			usvc, err := svc.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, usvc)
		})
	}
}
