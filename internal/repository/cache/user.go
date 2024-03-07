package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (c *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := c.key(id)
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}

	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (c *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func (c *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := c.key(du.Id)
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}
