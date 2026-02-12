/**************************************************************/
/*
   events.go

   Copyright 2026 Yabe.Kazuhiro
/**************************************************************/

package cube

type Event interface{ isEvent() }

type PositionID struct {
	CubeX, CubeY     uint16
	CubeAngle        uint16
	SensorX, SensorY uint16
	SensorAngle      uint16
}

func (PositionID) isEvent() {}

type StandardID struct {
	Value uint32
	Angle uint16
}

func (StandardID) isEvent() {}

type PositionIDMissed struct{}

func (PositionIDMissed) isEvent() {}

type StandardIDMissed struct{}

func (StandardIDMissed) isEvent() {}
