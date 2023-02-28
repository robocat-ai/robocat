package robocat

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sakirsensoy/genv"
	"nhooyr.io/websocket"
)

var wsServerAddress string

var wsServerUsername = "test"
var wsServerPassword = "test"

func TestMain(m *testing.M) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %s", err)
	}

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

	var container *dockertest.Resource

	existing, found := pool.ContainerByName("robocat-test")
	if found {
		if genv.Key("DOCKERTEST_REUSE_CONTAINER").Bool() {
			container = existing
		} else if err := pool.Purge(existing); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	if container == nil {
		container, err = pool.BuildAndRunWithOptions("./../../Dockerfile", &dockertest.RunOptions{
			Name:         "robocat-test",
			ExposedPorts: []string{"80/tcp"},
			Env: []string{
				fmt.Sprintf("AUTH_USERNAME=%s", wsServerUsername),
				fmt.Sprintf("AUTH_PASSWORD=%s", wsServerPassword),
			},
		}, func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.Mounts = append(config.Mounts, docker.HostMount{
				Target: "/home/robocat/flow",
				Source: fmt.Sprintf("%s/test-flow", pwd),
				Type:   "bind",
			})
		})
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}
	}

	wsServerAddress = container.GetHostPort("80/tcp")

	if err := pool.Retry(func() error {
		u, err := url.Parse(fmt.Sprintf("http://%s", wsServerAddress))
		if err != nil {
			return err
		}

		if len(wsServerUsername) > 0 {
			u.User = url.UserPassword(wsServerUsername, wsServerPassword)
		}

		conn, _, err := websocket.Dial(
			context.Background(),
			u.String(),
			&websocket.DialOptions{
				Subprotocols: []string{"robocat"},
			})
		if err != nil {
			return err
		}

		return conn.Close(websocket.StatusNormalClosure, "")
	}); err != nil {
		log.Fatalf("Could not connect to the server: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// if err := pool.Purge(container); err != nil {
	// 	log.Fatalf("Could not purge resource: %s", err)
	// }

	os.Exit(code)
}
