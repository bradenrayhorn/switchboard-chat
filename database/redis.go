package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"log"
)

type RedisDB struct {
	Client *redis.Client
}

func MakeRedisClient() *RedisDB {
	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis_address"),
		Username: viper.GetString("redis_username"),
		Password: viper.GetString("redis_password"),
		DB:       viper.GetInt("redis_db"),
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Println("error connecting to redis")
		log.Println(err)
	}
	redisDB := &RedisDB{
		Client: client,
	}
	return redisDB
}
