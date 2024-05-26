package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var ctx = context.Background()

func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:            "127.0.0.1:6379",
		Password:        "",                // 密码
		DB:              0,                 // 使用哪个数据库
		MaxIdleConns:    50,                // 连接池存在最大空闲连接数，连接超出阈值，关闭旧的。PoolSize是超出阈值，开辟新的连接
		MaxActiveConns:  30,                // 连接池可用最大连接
		ConnMaxIdleTime: 300 * time.Second, //空闲连接的最大时间，超时将回收连接
	})
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // 没有密码，默认值
		DB:       0,  // 默认DB 0
	})

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(pong)
	return rdb
}
