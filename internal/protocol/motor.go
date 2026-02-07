// **************************************************************
//
//	motor.go
//
//	Copyright 2026 Yabe.Kazuhiro
//
// **************************************************************
package protocol

// dir: 0x01 forward, 0x02 backward
func motorDir(speed int) (dir byte, abs byte) {
	if speed >= 0 {
		if speed > 255 {
			speed = 255
		}
		return 0x01, byte(speed)
	}
	if speed < -255 {
		speed = -255
	}
	return 0x02, byte(-speed)
}

// MotorControl: control type 0x01
// [0]=0x01, [1]=left id(0x01), [2]=left dir, [3]=left speed,
// [4]=right id(0x02), [5]=right dir, [6]=right speed
func MotorControl(left int, right int) []byte {
	ld, ls := motorDir(left)
	rd, rs := motorDir(right)
	return []byte{
		0x01,
		0x01, ld, ls,
		0x02, rd, rs,
	}
}
