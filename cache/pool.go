package cache

import (
	"fmt"
	"github.com/apsystole/log"
	"github.com/gomodule/redigo/redis"
	"os"
)

var redisPool *redis.Pool

// Pool returns a pool to get Redis connections from
func Pool() (*redis.Pool, error) {
	if redisPool == nil {

		// Pre-declare err to avoid shadowing redisPool
		var err error
		redisPool, err = initializeRedis()
		if err != nil {
			log.Errorf("Redis initialization failed: %v", err)
			return nil, err
		}
	}
	return redisPool, nil
}

func getRedisConfig() *RedisConfig {
	var config RedisConfig
	_ = os.Setenv(os.Getenv("REDIS_HOST"), "localhost")
	_ = os.Setenv(os.Getenv("REDIS_PORT"), "6379")
	config.Host = os.Getenv("REDIS_HOST")
	config.Port = os.Getenv("REDIS_PORT")
	config.Password = os.Getenv("REDIS_PASSWORD")

	return &config
}

// initializeRedis initializes and returns a Redis connection pool
func initializeRedis() (*redis.Pool, error) {
	rc := getRedisConfig()
	redisAddr := fmt.Sprintf("%s:%s", rc.Host, rc.Port)

	var dialOptions []redis.DialOption
	dialOptions = append(dialOptions, redis.DialPassword(rc.Password))

	log.Debugf("Connected to Redis Address: %s", redisAddr)

	return &redis.Pool{
		Wait:    true,
		MaxIdle: 100,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", redisAddr, dialOptions...)
			if err != nil {
				return nil, fmt.Errorf("redis.Dial: %w", err)
			}
			return conn, err
		},
	}, nil
}
