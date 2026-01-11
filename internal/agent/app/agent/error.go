package agent

import (
	"errors"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
)

/*
   @Author: orbit-w
   @File: error
   @2023 12月 周六 22:27
*/

func AgentDecodePatternErr(p int8) error {
	return errors.New("agent decode pattern error: " + utils.FormatInteger(p))
}
