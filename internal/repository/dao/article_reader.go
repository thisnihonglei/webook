package dao

import (
	"context"
	"gorm.io/gorm"
)

type ArticleReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticle) error
}

type ArticleReaderGORMDAO struct {
	db *gorm.DB
}

func (a *ArticleReaderGORMDAO) UpsertV2(ctx context.Context, art PublishedArticle) error {
	//TODO implement me
	panic("implement me")
}

func (a *ArticleReaderGORMDAO) Upsert(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

func NewArticleReaderGORMDAO(db *gorm.DB) ArticleReaderDAO {
	return &ArticleReaderGORMDAO{db: db}
}
