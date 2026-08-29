package server

import "unsafe"

// SentryUpdateData is GAME.EXE's fixed-width 12-byte SentryGlobeUpdate
// record. The transfer routine treats all three fields as raw dwords so their
// exact floating-point bit patterns survive every supported architecture.
type SentryUpdateData struct {
	Field0 uint32
	Field4 uint32
	Field8 uint32
}

var (
	_ = [1]struct{}{}[12-unsafe.Sizeof(SentryUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SentryUpdateData{}.Field0)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(SentryUpdateData{}.Field4)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(SentryUpdateData{}.Field8)]
)

func (obj *Object) UpdateDataSentry() *SentryUpdateData {
	// Preserve the raw entry pointer and the original nil-fault boundary.
	return (*SentryUpdateData)(obj.UpdateData)
}
