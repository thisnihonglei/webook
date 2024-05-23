package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"
)

func TestCachedArticleRepository_SyncV1(t *testing.T) {
	testCase := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (dao.ArticleReaderDAO, dao.ArticleAuthorDAO)
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "作者新建帖子同步读者成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleReaderDAO, dao.ArticleAuthorDAO) {
				authorRepo := daomocks.NewMockArticleAuthorDAO(ctrl)
				authorRepo.EXPECT().Create(gomock.Any(), dao.Article{
					Title:    "新标题",
					Content:  "新内容",
					AuthorId: 123,
				}).Return(int64(1), nil)
				readRepo := daomocks.NewMockArticleReaderDAO(ctrl)
				readRepo.EXPECT().Upsert(gomock.Any(), dao.Article{
					Id:       1,
					Title:    "新标题",
					Content:  "新内容",
					AuthorId: 123,
				}).Return(nil)
				return readRepo, authorRepo
			},
			art: domain.Article{
				Title:   "新标题",
				Content: "新内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "作者修改帖子同步读者成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleReaderDAO, dao.ArticleAuthorDAO) {
				authorRepo := daomocks.NewMockArticleAuthorDAO(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), dao.Article{
					Id:       11,
					Title:    "新标题",
					Content:  "新内容",
					AuthorId: 123,
				}).Return(nil)
				readRepo := daomocks.NewMockArticleReaderDAO(ctrl)
				readRepo.EXPECT().Upsert(gomock.Any(), dao.Article{
					Id:       11,
					Title:    "新标题",
					Content:  "新内容",
					AuthorId: 123,
				}).Return(nil)
				return readRepo, authorRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "新标题",
				Content: "新内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  11,
			wantErr: nil,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			readerDao, authorDao := tc.mock(ctrl)
			repo := NewCachedArticleRepositoryV2(readerDao, authorDao)
			id, err := repo.SyncV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
