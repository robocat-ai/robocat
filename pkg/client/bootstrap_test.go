package robocat

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sakirsensoy/genv"
)

var wsServerAddress string

var wsServerUsername = "test"
var wsServerPassword = "test"

func TestMain(m *testing.M) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %s", err)
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
		runOptions := &dockertest.RunOptions{
			Name:         "robocat-test",
			ExposedPorts: []string{"80/tcp"},
			Env: []string{
				"DEBUG=1",
				fmt.Sprintf("AUTH_USERNAME=%s", wsServerUsername),
				fmt.Sprintf("AUTH_PASSWORD=%s", wsServerPassword),
			},
		}

		if genv.Key("DOCKERTEST_RUN_AS_ROOT").Bool() {
			runOptions.User = "root"
		}

		container, err = pool.BuildAndRunWithOptions(
			"./../../Dockerfile",
			runOptions,
			func(config *docker.HostConfig) {
				config.AutoRemove = true
				config.Mounts = append(config.Mounts, docker.HostMount{
					Target: "/flow",
					Source: fmt.Sprintf("%s/test-flow", pwd),
					Type:   "bind",
				})
				config.Mounts = append(config.Mounts, docker.HostMount{
					Target: "/home/robocat/flow",
					Source: fmt.Sprintf("%s/test-flow", pwd),
					Type:   "bind",
				})
			},
		)
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}
	}

	wsServerAddress = container.GetHostPort("80/tcp")

	if err := pool.Retry(func() error {
		log.Printf("Trying to connect to %s...", wsServerAddress)

		client, err := Connect(
			fmt.Sprintf("ws://%s", wsServerAddress), Credentials{
				wsServerUsername, wsServerPassword,
			},
		)
		if err != nil {
			return err
		}

		defer client.Close()
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to the server: %s", err)
	}

	code := m.Run()

	os.Exit(code)
}
