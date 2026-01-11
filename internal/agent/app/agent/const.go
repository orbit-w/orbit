package agent

/*
   @Author: orbit-w
   @File: const
   @2024 4月 周二 22:57
*/

const (
	StatusNormal = iota
	StatusClosed

	headLen         = 4
	MaxInPacketSize = 1048576 //1MB
)

const (
	PatternNone = iota
	PatternKick
)
