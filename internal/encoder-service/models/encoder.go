package models

import (
	"context"
	"os/exec"

	"go.uber.org/zap"
)

// StreamEncoder represents an active encoding process for a stream
type StreamEncoder struct {
	StreamKey string
	Cmd       *exec.Cmd
	Ctx       context.Context
	Cancel    context.CancelFunc
	Logger    *zap.Logger
}
