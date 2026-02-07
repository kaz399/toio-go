package toio

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/kaz399/toio-go/internal/ble"
)

type Device struct {
	Name    string
	Address string
	RSSI    int

	BleAddr ble.Address
}

type Scanner struct {
	Timeout time.Duration
}

func (s Scanner) Scan(ctx context.Context, num int) ([]Device, error) {
	if num <= 0 {
		return nil, nil
	}

	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var mu sync.Mutex
	found := make([]Device, 0, num)
	seen := map[string]bool{}

	errCh := make(chan error, 1)

	go func() {
		errCh <- ble.Scan(func(r ble.ScanResult) {
			name := r.LocalName()
			if !strings.HasPrefix(name, "toio-") {
				return
			}

			key := r.Address.String()

			mu.Lock()
			defer mu.Unlock()

			if seen[key] {
				return
			}
			seen[key] = true

			found = append(found, Device{
				Name:    name,
				Address: key,
				RSSI:    int(r.RSSI),
				BleAddr: r.Address,
			})

			if len(found) >= num {
				_ = ble.StopScan()
				cancel()
			}
		})
	}()

	<-ctx.Done()
	_ = ble.StopScan()

	err := <-errCh
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		println("scan timeout")
	}

	mu.Lock()
	defer mu.Unlock()

	// ポインタを共有させないためにコピーして返す
	result := make([]Device, len(found))
	copy(result, found)
	return result, err
}
