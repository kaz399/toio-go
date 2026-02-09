/**************************************************************/
/*
   main.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"context"
	"log"
	"time"

	"github.com/kaz399/toio-go/cube"
	"github.com/kaz399/toio-go/internal/ble"
	"github.com/kaz399/toio-go/toio"
)

func main() {
	if err := ble.Enable(); err != nil {
		log.Fatal(err)
	}

	sc := toio.Scanner{Timeout: 10 * time.Second}
	ctx := context.Background()
	devs, err := sc.Scan(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}
	if len(devs) == 0 {
		log.Fatal("no toio found")
	}

	c := cube.New(devs[0])
	if err := c.Connect(); err != nil {
		log.Fatal(err)
	}
	defer c.Disconnect()

	if err := c.MotorControl(60, -60); err != nil {
		log.Fatal(err)
	}
	time.Sleep(2 * time.Second)

	_ = c.MotorControl(0, 0)
}
