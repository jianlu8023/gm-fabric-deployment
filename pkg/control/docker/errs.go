package docker

import (
	"errors"
)

var ErrNoAliveDockerClient = errors.New("no alive docker client")

func IsNoAliveDockerClient(err error) bool {
	return errors.Is(err, ErrNoAliveDockerClient)
}
