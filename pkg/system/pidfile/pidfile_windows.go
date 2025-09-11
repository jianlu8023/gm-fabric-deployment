//go:build windows

package pidfile

import (
	"errors"
)

var ErrNotSupported = errors.New("windows not supported")

func CreateOrUpdatePIDFile(filename string) error {
	return ErrNotSupported
}

func ReleasePID() {}
