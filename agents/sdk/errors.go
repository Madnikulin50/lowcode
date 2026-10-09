package sdk

import "errors"

var (
	errUseBackend  = errors.New("sdk: descriptor-only component; use Backend.StartJob")
	ErrJobNotFound = errors.New("sdk: job not found")
	ErrNoCall      = errors.New("sdk: sync calls not configured")
)
