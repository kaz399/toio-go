//**************************************************************
//
//  scanner.go
//
//  Copyright 2026 Yabe.Kazuhiro
//
//**************************************************************
package toio

import (
	"context"
	"time"
)

type Scanner struct {
	Timeout time.Duration
}

type Device struct {
	Name string
	Address string
	RSSI int
}

func (s *Scanner) Scan(ctx context.Context, num int) ([]Device, error)
func (s *Scanner) ScanById(ctx context.Context, ids ...string) ([]Device, error)

