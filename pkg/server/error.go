package gor

import "errors"

var (
	// ErrEngineRouterIsMissing indicates that Start was called before adding a router.
	ErrEngineRouterIsMissing = errors.New("engine routes is missing")
	// ErrEngineAlreadyStarted indicates that Start was called more than once.
	ErrEngineAlreadyStarted = errors.New("engine already started")
	// ErrEngineNotStarted indicates that a lifecycle operation requires Start first.
	ErrEngineNotStarted = errors.New("engine not started")
)
