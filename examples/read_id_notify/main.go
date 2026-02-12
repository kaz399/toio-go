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
	ctx := context.Background()

	if err := ble.Enable(); err != nil {
		log.Fatal(err)
	}

	sc := toio.Scanner{Timeout: 10 * time.Second}
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

	if err := c.StartIDNotify(); err != nil {
		log.Fatal(err)
	}

	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case ev := <-c.Events():
			log.Printf("events %#v\n", ev)
		case <-timeout.C:
			return
		}
	}
}
