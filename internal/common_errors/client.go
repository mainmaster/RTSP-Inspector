package common_errors

import "errors"

var (
	ErrBadCredentials = errors.New("wrong login or password")
)
