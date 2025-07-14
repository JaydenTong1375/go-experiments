package redis

import (
	"time"

	"github.com/go-redsync/redsync/v4"
	redsync_goredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// Redis server
var rdb = redis.NewClient(&redis.Options{
	Addr: "redis:6379", // Use the service name defined in docker-compose.yaml as the hostname
})

var pool = redsync_goredis.NewPool(rdb)
var rs = redsync.New(pool)

zz
