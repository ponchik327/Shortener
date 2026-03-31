package rediscache_test

import (
	"context"
	"os"
	"testing"

	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	wbfredis "github.com/wb-go/wbf/redis"
)

var testRedisClient *wbfredis.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		panic("start redis container: " + err.Error())
	}

	endpoint, err := redisContainer.PortEndpoint(ctx, "6379/tcp", "")
	if err != nil {
		panic("get redis endpoint: " + err.Error())
	}

	testRedisClient, err = wbfredis.Connect(wbfredis.Options{Address: endpoint})
	if err != nil {
		panic("connect to redis: " + err.Error())
	}

	code := m.Run()

	_ = redisContainer.Terminate(ctx)

	os.Exit(code)
}
