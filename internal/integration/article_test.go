package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/internal/integration/startup"
	"webook/internal/repository/dao"
	ijwt "webook/internal/web/jwt"
)

type ArticleHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func (s *ArticleHandlerSuite) SetupSuite() {
	s.db = startup.InitDB()
	server := gin.Default()
	hdl := startup.InitArticleHandler()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", ijwt.UserClaims{
			Uid: 123,
		})
	})
	hdl.RegisterRoutes(server)
	s.server = server
}

func (s *ArticleHandlerSuite) TearDownTest() {
	s.db.Exec("truncate table `articles`")
}

func (s *ArticleHandlerSuite) TestEdit() {
	t := s.T()
	testCase := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)
		req    Article

		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("author_id=?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, "this title", art.Title)
				assert.Equal(t, "this content", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
			},
			req: Article{
				Title:   "this title",
				Content: "this content",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "更新帖子",
			before: func(t *testing.T) {
				err := s.db.Create(&dao.Article{
					Id:       2,
					Title:    "标题",
					Content:  "内容",
					AuthorId: 123,
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, "new title", art.Title)
				assert.Equal(t, "new content", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.True(t, art.Utime > 789)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "new title",
					Content:  "new content",
					AuthorId: 123,
					Ctime:    456,
				}, art)
			},
			req: Article{
				Id:      2,
				Title:   "new title",
				Content: "new content",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "修改帖子-别人的帖子",
			before: func(t *testing.T) {
				// 假装数据库已经有这个帖子
				err := s.db.Create(&dao.Article{
					Id:      22,
					Title:   "我的标题",
					Content: "我的内容",
					// 模拟别人
					AuthorId: 1024,
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 你要验证，保存到了数据库里面
				var art dao.Article
				err := s.db.Where("id=?", 22).
					First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       22,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 1024,
					Ctime:    456,
					Utime:    789,
				}, art)
			},
			req: Article{
				Id:      22,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Msg: "系统错误",
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)

			// 准备请求
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewReader(data))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)

			// 准备记录响应
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var result Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResult, result)

		})
	}
}

func TestArticleHandler(t *testing.T) {
	suite.Run(t, &ArticleHandlerSuite{})
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
