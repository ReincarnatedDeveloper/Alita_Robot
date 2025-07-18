// Package cache provides caching functionality for the Alita Robot.
//
// This package implements Redis-based caching with support for both string
// and marshal (object) caching. It provides optimized storage and retrieval
// for frequently accessed data to improve bot performance.
package cache

import (
	"context"
	"sync"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	redis_store "github.com/eko/gocache/store/redis/v4"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	"github.com/redis/go-redis/v9"
)

var (
	Context = context.Background()
	Marshal *marshaler.Marshaler
	Manager *cache.ChainCache[any]
)

/*
AdminCache represents the cached administrator information for a chat.

Fields:
  - ChatId:   The unique identifier for the chat.
  - UserInfo: A slice of merged chat member information for each admin.
  - Cached:   Indicates if the cache is valid and populated.
*/
type AdminCache struct {
	ChatId   int64
	UserInfo []gotgbot.MergedChatMember
	UserMap  map[int64]gotgbot.MergedChatMember
	Cached   bool
	mux      sync.RWMutex
}

/*
InitCache initializes the caching system for the application.

It sets up both Redis and Ristretto as cache backends, creates a chain cache manager,
and initializes the marshaler for serializing and deserializing cached data.
Panics if Ristretto cache initialization fails.
*/
func InitCache() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Password: config.RedisPassword, // no password set
		DB:       config.RedisDB,       // use default DB
	})
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	// initialize cache manager
	redisStore := redis_store.NewRedis(redisClient)
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
	cacheManager := cache.NewChain(cache.New[any](ristrettoStore), cache.New[any](redisStore))

	// Initializes marshaler
	Marshal = marshaler.New(cacheManager)
}
