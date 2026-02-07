/**************************************************************/
/*
   motor.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/
package cube

import "github.com/kaz399/toio-go/internal/protocol"

func (c *Cube) MotorControl(left, right int) error {
	cmd := protocol.MotorControl(left, right)
	_, err := c.motor.WriteWithoutResponse(cmd)
	return err
}
