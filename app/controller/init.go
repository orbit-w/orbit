package controller

import (
	"gitee.com/orbit-w/orbit/app/core/dispatch"
)

/*
@Author: orbit-w
@File: init
@2025 9月 周一 14:45
*/
func init() {
	if err := dispatch.RegisterController(GExampleController); err != nil {
		panic(err)
	}
}
