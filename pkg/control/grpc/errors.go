package grpc

import (
	"errors"
)

var (
	ErrNoCACert       = errors.New("CA cert file is required for mutual TLS")
	ErrFailAppendCert = errors.New("failed to append CA cert to pool")
)
