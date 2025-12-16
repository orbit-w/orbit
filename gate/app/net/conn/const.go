package conn

import "time"

/*
   @Author: orbit-w
   @File: const
   @2024 4月 周六 10:25
*/

const (
	StatusNormal = iota
	StatusClosed

	WriteTimeout = time.Second * 5
)
