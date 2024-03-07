package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode        string
	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证太频繁")
)

var ErrKeyNotExist = redis.Nil

type CodeRedisCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type RedisCodeCache struct {
	cmd redis.Cmdable
}

func NewCodeCache(cmd redis.Cmdable) CodeRedisCache {
	return &RedisCodeCache{
		cmd: cmd,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.cmd.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		//调用redis报错
		return err
	}
	switch res {
	case -2:
		return errors.New("验证码存在，但是没有过期时间")
	case -1:
		return ErrCodeSendTooMany
	default:
		return nil
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		//调用redis报错
		return false, err
	}
	switch res {
	case -2:
		return false, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	default:
		return true, nil
	}
}

func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type LocalCodeCache struct {
	cmd *lru.Cache
	// 读写锁
	lock       sync.Mutex
	expiration time.Duration
}

func NewLocalCodeCache(c *lru.Cache, expiration time.Duration) CodeCache {
	return &LocalCodeCache{
		cmd:        c,
		expiration: expiration,
	}
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	key := l.key(biz, phone)

	timeN := time.Now()

	val, ok := l.cmd.Get(key)
	if !ok {
		//缓存中没有找到短信验证码 Set短信验证码，验证码使用次数3次，Set过期时间
		l.cmd.Add(key, codeItem{
			code:   code,
			cnt:    3,
			expire: timeN.Add(l.expiration),
		})
		return nil
	}
	itm, ok := val.(codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return errors.New("系统错误")
	}

	if itm.expire.Sub(timeN) > time.Minute*9 {
		//不到一分钟，发送太频繁，返回报错
		return ErrCodeSendTooMany
	}
	l.cmd.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: timeN.Add(l.expiration),
	})
	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	key := l.key(biz, phone)
	val, ok := l.cmd.Get(key)
	if !ok {
		return false, ErrKeyNotExist
	}

	itm, ok := val.(codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return false, errors.New("系统错误")
	}
	if itm.cnt <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	itm.cnt--
	return itm.code == inputCode, nil
}

func (l *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type codeItem struct {
	code   string
	cnt    int
	expire time.Time
}
