package integrationtest

import (
	"fmt"
	"github.com/ory/dockertest/v3"
	"kingscomp/internal/repository/redis"
	"kingscomp/pkg/helpers"
	"os"
	"testing"
)

var redisPort string

func TestMain(m *testing.M) {
	if !helpers.IsIntegration() {
		return
	}

	pool := helpers.StartDockerPool()

	// set up the redis container for tests
	redisRes := helpers.StartDockerInstance(pool, "redis/redis-stack-server", "latest",
		func(res *dockertest.Resource) error {
			port := res.GetPort("6379/tcp")
			_, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", port))
			return err
		})

	redisPort = redisRes.GetPort("6379/tcp")

	// now run tests
	exitCode := m.Run()
	redisRes.Close()
	os.Exit(exitCode)
}
