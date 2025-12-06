package cache

import (
	"context"
	"fmt"
	"time"

	"xboard/internal/config"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	rdb *redis.Client
	ctx context.Context
}

func New(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb, ctx: ctx}, nil
}

func (c *Client) Get(key string) (string, error) {
	return c.rdb.Get(c.ctx, key).Result()
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(c.ctx, key, value, expiration).Err()
}

func (c *Client) Del(key string) error {
	return c.rdb.Del(c.ctx, key).Err()
}

func (c *Client) Exists(key string) (bool, error) {
	n, err := c.rdb.Exists(c.ctx, key).Result()
	return n > 0, err
}

func (c *Client) Incr(key string) (int64, error) {
	return c.rdb.Incr(c.ctx, key).Result()
}

func (c *Client) IncrBy(key string, value int64) (int64, error) {
	return c.rdb.IncrBy(c.ctx, key, value).Result()
}

func (c *Client) HGet(key, field string) (string, error) {
	return c.rdb.HGet(c.ctx, key, field).Result()
}

func (c *Client) HSet(key, field string, value interface{}) error {
	return c.rdb.HSet(c.ctx, key, field, value).Err()
}

func (c *Client) HGetAll(key string) (map[string]string, error) {
	return c.rdb.HGetAll(c.ctx, key).Result()
}

// Cache key prefixes
const (
	KeyServerLastCheckAt = "SERVER_%s_LAST_CHECK_AT_%d"
	KeyServerLastPushAt  = "SERVER_%s_LAST_PUSH_AT_%d"
	KeyServerOnlineUser  = "SERVER_%s_ONLINE_USER_%d"
	KeyServerLoadStatus  = "SERVER_%s_LOAD_STATUS_%d"
	KeyUserOnline        = "USER_ONLINE_%d"
)

func ServerLastCheckAtKey(serverType string, serverID int64) string {
	return fmt.Sprintf(KeyServerLastCheckAt, serverType, serverID)
}

func ServerLastPushAtKey(serverType string, serverID int64) string {
	return fmt.Sprintf(KeyServerLastPushAt, serverType, serverID)
}

func ServerOnlineUserKey(serverType string, serverID int64) string {
	return fmt.Sprintf(KeyServerOnlineUser, serverType, serverID)
}

func ServerLoadStatusKey(serverType string, serverID int64) string {
	return fmt.Sprintf(KeyServerLoadStatus, serverType, serverID)
}
