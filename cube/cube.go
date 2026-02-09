/**************************************************************/
/*
   cube.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package cube

import (
	"errors"

	"github.com/kaz399/toio-go/internal/ble"
	"github.com/kaz399/toio-go/internal/protocol"
	"github.com/kaz399/toio-go/toio"
	"tinygo.org/x/bluetooth"
)

type Cube struct {
	dev toio.Device

	connected bool
	device    bluetooth.Device
	motor     bluetooth.DeviceCharacteristic
}

func New(d toio.Device) *Cube { return &Cube{dev: d} }

func (c *Cube) Connect() error {
	device, err := ble.Connect(c.dev.BleAddr)
	if err != nil {
		return err
	}
	c.device = device

	services, err := device.DiscoverServices([]bluetooth.UUID{protocol.ServiceToio})
	if err != nil {
		_ = device.Disconnect()
		return err
	}
	if len(services) == 0 {
		_ = device.Disconnect()
		return err
	}

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{protocol.CharMotor})
	if err != nil {
		_ = device.Disconnect()
		return err
	}
	if len(chars) == 0 {
		_ = device.Disconnect()
		return errors.New("motor characteristic not found")
	}
	c.motor = chars[0]
	c.connected = true
	return nil
}

func (c *Cube) Disconnect() error {
	if !c.connected {
		return nil
	}
	return c.device.Disconnect()
}
