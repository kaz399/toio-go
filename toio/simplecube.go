// **************************************************************
//
//	simplecube.go
//
//	Copyright 2026 Yabe.Kazuhiro
//
// **************************************************************
package toio

import (
	"context"
	"time"
)

type SimpleCube struct {
	Cube *cube.Cube
}

type SimpleOptions struct {
	NameSuffix3 string
	Timeout time.Duration
}

func NewSimpleCube(ctx context.Context, opt SimpleOptions) (*SimpleCube, error)


