package auth

import "errors"

var ErrInvalidSignature = errors.New("invalid signature")
var ErrActionMissing = errors.New("action not found")
var ErrNotAllowed = errors.New("not allowed")
var ErrInvalidState = errors.New("invalid state")
var ErrInvalidStateKey = errors.New("invalid state key")
var ErrInvalidStateValue = errors.New("invalid state value")
var ErrInvalidStateType = errors.New("invalid state type")
