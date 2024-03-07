package repository

import (
	"context"
	"database/sql"
	"log"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByName(ctx context.Context, email string) (domain.User, error)
	Edit(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(dao dao.UserDAO, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (repo *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.toEntity(u))
}

func (repo *CachedUserRepository) FindByName(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByName(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CachedUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Birthday: u.Birthday,
		NickName: u.NickName,
		AboutMe:  u.AboutMe,
	}
}

func (repo *CachedUserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Birthday: u.Birthday,
		NickName: u.NickName,
		AboutMe:  u.AboutMe,
	}
}

func (repo *CachedUserRepository) Edit(ctx context.Context, u domain.User) error {
	return repo.dao.Edit(ctx, dao.User{
		Id:       u.Id,
		Birthday: u.Birthday,
		AboutMe:  u.AboutMe,
		NickName: u.NickName,
	})
}

func (repo *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, id)
	// 只要err为nil，就返回
	if err == nil {
		return du, err
	}

	// err 不为nil，再去查询数据库
	// 这里err有两种可能
	// 1.key不存在，redis是正常的，从redis中获取不到用户信息
	// 2.redis有问题，访问redis报错了，可能是网络问题，也可能是redis本身崩溃了

	u, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	du = repo.toDomain(u)
	go func() {
		err = repo.cache.Set(ctx, du)
		if err != nil {
			log.Println(err)
		}
	}()
	return du, nil
}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}
