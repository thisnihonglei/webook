package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	cachemocks "webook/internal/repository/cache/mocks"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	testCase := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)

		ctx context.Context
		uid int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存 未命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{
					Id: 123,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "123456",
					Birthday: 100,
					AboutMe:  "自我介绍",
					Phone: sql.NullString{
						String: "15801088888",
						Valid:  true,
					},
					NickName: "nick_name",
					Ctime:    101,
					Utime:    102,
				}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: 100,
					AboutMe:  "自我介绍",
					Phone:    "15801088888",
					NickName: "nick_name",
				}).Return(nil)
				return d, c
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: 100,
				AboutMe:  "自我介绍",
				Phone:    "15801088888",
				NickName: "nick_name",
			},
			wantErr: nil,
		},
		{
			name: "查找成功，缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: 100,
					AboutMe:  "自我介绍",
					Phone:    "15801088888",
					NickName: "nick_name",
				}, nil)
				return d, c
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: 100,
				AboutMe:  "自我介绍",
				Phone:    "15801088888",
				NickName: "nick_name",
			},
			wantErr: nil,
		},
		{
			name: "查找失败，没有找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{}, dao.ErrRecordNotFound)
				return d, c
			},
			uid:      123,
			ctx:      context.Background(),
			wantUser: domain.User{},
			wantErr:  dao.ErrRecordNotFound,
		},
		{
			name: "回写缓存失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{
					Id: 123,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "123456",
					Birthday: 100,
					AboutMe:  "自我介绍",
					Phone: sql.NullString{
						String: "15801088888",
						Valid:  true,
					},
					NickName: "nick_name",
					Ctime:    101,
					Utime:    102,
				}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: 100,
					AboutMe:  "自我介绍",
					Phone:    "15801088888",
					NickName: "nick_name",
				}).Return(errors.New("redis error"))
				return d, c
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: 100,
				AboutMe:  "自我介绍",
				Phone:    "15801088888",
				NickName: "nick_name",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tc.mock(ctrl)
			svc := NewCachedUserRepository(ud, uc)
			user, err := svc.FindById(tc.ctx, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)

		})
	}
}
