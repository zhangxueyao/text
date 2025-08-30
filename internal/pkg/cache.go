package pkg

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

type CacheManager struct {
	// 本地缓存
	localCache map[string]interface{}

	// Redis客户端
	redisClient *redis.Client

	// 数据库客户端
	dbClient sqlx.SqlConn
}

func NewCacheManager(redisClient *redis.Client, dbClient sqlx.SqlConn) *CacheManager {
	return &CacheManager{
		localCache:  make(map[string]interface{}),
		redisClient: redisClient,
		dbClient:    dbClient,
	}
}

func (cm *CacheManager) GetData(key string) (interface{}, error) {
	// 1. 检查本地缓存
	if value, exists := cm.localCache[key]; exists {
		fmt.Println("Cache hit in local cache")
		return value, nil
	}

	// 2. 检查 Redis 缓存
	ctx := context.Background()
	value, err := cm.redisClient.Get(ctx, key).Result()
	if err == nil {
		// Redis缓存命中
		fmt.Println("Cache hit in Redis")
		// 将 Redis 数据写入本地缓存
		cm.localCache[key] = value
		return value, nil
	} else if err != redis.Nil {
		// Redis 错误
		return nil, err
	}

	// 3. Redis缓存未命中，查询数据库
	fmt.Println("Cache miss in Redis, querying database")
	dbValue, err := cm.queryDatabase(key)
	if err != nil {
		return nil, err
	}

	// 将数据保存到 Redis 和本地缓存
	cm.redisClient.Set(ctx, key, dbValue, time.Minute*10) // 设置 Redis 缓存过期时间
	cm.localCache[key] = dbValue
	return dbValue, nil
}

// queryDatabase 模拟数据库查询
func (cm *CacheManager) queryDatabase(key string) (string, error) {
	// 模拟数据库查询
	return "Database data for " + key, nil
}
