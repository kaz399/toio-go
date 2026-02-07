// **************************************************************
//
//	tinygo.go
//
//	Copyright 2026 Yabe.Kazuhiro
//
// **************************************************************
package ble

import (
	"tinygo.org/x/bluetooth"
)

var Adapter = bluetooth.DefaultAdapter

type ScanResult = bluetooth.ScanResult
type Device = bluetooth.ScanResult
type Address = bluetooth.Address

func Enable() error { return Adapter.Enable() }

func Scan(cb func(ScanResult)) error {
	return Adapter.Scan(func(_ *bluetooth.Adapter, r bluetooth.ScanResult) { cb(r) })
}

func StopScan() error { return Adapter.StopScan() }

func Connect(addr bluetooth.Address) (bluetooth.Device, error) {
	// Initially, there are no parameters.
	return Adapter.Connect(addr, bluetooth.ConnectionParams{})
}
