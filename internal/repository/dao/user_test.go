package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCase := []struct {
		name    string
		mock    func(t *testing.T) *sql.DB
		ctx     context.Context
		user    User
		wantErr error
	}{
		{
			name: "Insert执行成功",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				// 传入SQL的正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnResult(mockRes)
				return db
			},
			ctx: context.Background(),
			user: User{
				NickName: "Tom",
			},
			wantErr: nil,
		},
		{
			name: "Insert Failed 邮箱存在，导致冲突",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				// 传入SQL的正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(&mysqlDriver.MySQLError{
						Number: 1062,
					})
				return db
			},
			ctx: context.Background(),
			user: User{
				NickName: "Tom",
			},
			wantErr: ErrDuplicateEmail,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				// 传入SQL的正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(errors.New("数据库错误"))
				return db
			},
			ctx: context.Background(),
			user: User{
				NickName: "Tom",
			},
			wantErr: errors.New("数据库错误"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.mock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			assert.NoError(t, err)
			dao := NewUserDAO(db)
			err = dao.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
