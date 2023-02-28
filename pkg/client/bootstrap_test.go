package robocat

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var wsServerAddress string

var wsServerUsername = "test"
var wsServerPassword = "test"

// type testServerInfo struct {
// }

// var wsServerInfo testServerInfo

func TestMain(m *testing.M) {
	// pwd, err := os.Getwd()
	// if err != nil {
	// 	log.Fatalf("failed to get working directory: %s", err)
	// }

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	existing, found := pool.ContainerByName("robocat-test")
	if found {
		if err := pool.Purge(existing); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	resource, err := pool.BuildAndRunWithOptions("./../../Dockerfile", &dockertest.RunOptions{
		Name:         "robocat-test",
		ExposedPorts: []string{"80/tcp"},
		Env: []string{
			fmt.Sprintf("AUTH_USERNAME=%s", wsServerUsername),
			fmt.Sprintf("AUTH_PASSWORD=%s", wsServerPassword),
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		// config.Mounts = append(config.Mounts, docker.HostMount{
		// Target: "/data",
		// Source: fmt.Sprintf("%s/examples/data", pwd),
		// Type:   "bind",
		// })
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	wsServerAddress = resource.GetHostPort("80/tcp")

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		u, err := url.Parse(fmt.Sprintf("http://%s", wsServerAddress))
		if err != nil {
			return err
		}

		if len(wsServerUsername) > 0 {
			u.User = url.UserPassword(wsServerUsername, wsServerPassword)
		}

		resp, err := http.Get(u.String())
		if err != nil {
			return err
		}

		if resp.StatusCode != 426 {
			return errors.New("unexpected response code, expected 426")
		}

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to the server: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
