package server

import "math"

// x87TruncSignedQwordLow566DCC models GAME.EXE helper 00566DCC. It truncates
// ST0 toward zero into a signed qword and exposes EAX, the low dword. An
// invalid x87 conversion stores integer-indefinite 0x8000000000000000, whose
// low dword is zero.
func x87TruncSignedQwordLow566DCC(value float64) int32 {
	if math.IsNaN(value) || value >= 0x1p63 || value < -0x1p63 {
		return 0
	}
	return int32(int64(math.Trunc(value)))
}
