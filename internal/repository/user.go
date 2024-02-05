package repository

import (
	"context"
	"github.com/gin-gonic/gin"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (repo *UserRepository) FindByName(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByName(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *UserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}

func (repo *UserRepository) Edit(ctx context.Context, u domain.User) error {
	return repo.dao.Edit(ctx, dao.User{
		Id:       u.Id,
		Birthday: u.Birthday,
		AboutMe:  u.AboutMe,
		NickName: u.NickName,
	})
}

func (repo *UserRepository) FindById(ctx *gin.Context, id int64) (domain.User, error) {
	u, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		NickName: u.NickName,
		Email:    u.Email,
		Birthday: u.Birthday,
		AboutMe:  u.AboutMe,
	}, nil
}
