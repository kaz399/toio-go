/**************************************************************/
/*
   id_notify.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package cube

import (
	"encoding/binary"
	"fmt"
)

func (c *Cube) StartIDNotify() error {
	return c.idChar.EnableNotifications(func(buf []byte) {
		ev, err := parseID(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		select {
		case c.events <- ev:
		default:
		}
	})
}

func parseID(b []byte) (Event, error) {
	if len(b) < 1 {
		return nil, fmt.Errorf("id notify: empty")
	}
	switch b[0] {
	case 0x01: // Position ID
		// +0: type
		// +1: cube x (u16)
		// +3: cube y (u16)
		// +5: cube angle (u16)
		// +7: sensor x (u16)
		// +9: sensor y (u16)
		// +11: sensor angle (u16)
		if len(b) < 13 {
			return nil, fmt.Errorf("position id: invalid data length: %d (expected 13 bytes)", len(b))
		}
		return PositionID{
			CubeX:       binary.LittleEndian.Uint16(b[1:3]),
			CubeY:       binary.LittleEndian.Uint16(b[3:5]),
			CubeAngle:   binary.LittleEndian.Uint16(b[5:7]),
			SensorX:     binary.LittleEndian.Uint16(b[7:9]),
			SensorY:     binary.LittleEndian.Uint16(b[9:11]),
			SensorAngle: binary.LittleEndian.Uint16(b[11:13]),
		}, nil

	case 0x02: //Standard ID
		if len(b) < 7 {
			return nil, fmt.Errorf("standard id: invalid data length: %d (expected 7 bytes)", len(b))
		}
		return StandardID{
			Value: binary.LittleEndian.Uint32(b[1:5]),
			Angle: binary.LittleEndian.Uint16(b[5:7]),
		}, nil

	case 0x03:
		return PositionIDMissed{}, nil

	case 0x04:
		return StandardIDMissed{}, nil

	default:
		return nil, fmt.Errorf("unknown id type: 0x%02x", b[0])
	}
}
