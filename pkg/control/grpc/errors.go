package grpc

import (
	"errors"
)

var (
	ErrNoCACert          = errors.New("CA cert file is required for mutual TLS")
	ErrFailAppendCert    = errors.New("failed to append CA cert to pool")
	ErrCertKeyMismatch   = errors.New("certificate and key file count mismatch")
	ErrInsufficientCerts = errors.New("insufficient certificate count for GM mode")
	ErrEmptyCertPath     = errors.New("certificate file path is empty")
	ErrEmptyKeyPath      = errors.New("key file path is empty")
)
