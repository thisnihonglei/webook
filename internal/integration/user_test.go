package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/internal/integration/startup"
	"webook/internal/web"
)

func TestUserHandler_SendSMSCode(t *testing.T) {
	rdb := startup.InitRedis()
	server := startup.InitWebServer()
	testCase := []struct {
		name string

		before func(t *testing.T)

		after func(t *testing.T)

		phone string

		wantCode int

		wantBody web.Result
	}{
		{
			name: "发送成功的用例",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:15801088210"
				code, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, len(code) > 0)
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9+time.Second*50)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			phone:    "15801088210",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name: "未输入手机号码",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "请输入手机号码",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:15801088210"
				err := rdb.Set(ctx, key, "123456", time.Minute*9+time.Second*50).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:15801088210"
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
			},
			phone:    "15801088210",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "短信发送太频繁，请稍后再试",
			},
		},

		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:15801088210"
				err := rdb.Set(ctx, key, "123456", 0).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:15801088210"
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
			},
			phone:    "15801088210",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			// 准备请求
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send",
				bytes.NewReader([]byte(fmt.Sprintf(`{"phone":"%s"}`, tc.phone))))
			req.Header.Set("Content-Type", "application/json")

			assert.NoError(t, err)

			// 准备记录响应
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res web.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, res)
		})
	}
}
