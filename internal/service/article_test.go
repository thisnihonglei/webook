package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
	"webook/pkg/logger"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (repository.ArticleReaderRepository, repository.ArticleAuthorRepository)

		art domain.Article

		wantId    int64
		wantError error
	}{
		{
			name: "新建发表成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleReaderRepository, repository.ArticleAuthorRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 使用制作库的ID
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				})
				return readerRepo, authorRepo
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:    1,
			wantError: nil,
		},
		{
			name: "修改并新发表成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleReaderRepository, repository.ArticleAuthorRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 使用制作库的ID
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				})
				return readerRepo, authorRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:    11,
			wantError: nil,
		},
		{
			name: "修改并发表失败，重试成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleReaderRepository, repository.ArticleAuthorRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 使用制作库的ID
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(errors.New("mock db error"))
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 使用制作库的ID
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				return readerRepo, authorRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:    11,
			wantError: nil,
		},
		{
			name: "修改并发表失败，重试失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleReaderRepository, repository.ArticleAuthorRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 使用制作库的ID
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Times(3).Return(errors.New("mock db error"))
				return readerRepo, authorRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:    11,
			wantError: errors.New("保存到线上库失败，重试次数耗尽"),
		},
		{
			name: "修改并保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleReaderRepository, repository.ArticleAuthorRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(errors.New("mock db error"))
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				return readerRepo, authorRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantError: errors.New("mock db error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			readerRepo, authorRepo := tc.mock(ctrl)
			svc := NewArticleServiceV1(readerRepo, authorRepo, logger.NewNopLogger())
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantError, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
