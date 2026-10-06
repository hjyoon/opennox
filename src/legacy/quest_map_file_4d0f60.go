package legacy

import "unsafe"

// Quest map history is a PE32 scalar record, not a native pointer table.
// Keep its 32-byte stride on every architecture; Name remains C-owned bytes.
type questMapEntry4D0F60 struct {
	Group    uint32
	Name     [20]byte
	Uses     uint32
	LastUsed uint32
}

type questMapFileHooks4D0F60 struct {
	Count     func() uint32
	LastIndex func() uint32
	Clock     func() uint32
	Entry     func(int32) *questMapEntry4D0F60
	Name      func(int32) unsafe.Pointer
	Random    func(int32, int32) int32
}

func questMapEligible4D0F60(h questMapFileHooks4D0F60, index int32, last uint32, threshold int32) bool {
	entry := h.Entry(index)
	if int32(entry.Uses) >= threshold || uint32(index) == last {
		return false
	}
	if entry.Group == h.Entry(int32(last)).Group {
		return false
	}
	// SUB is a DWORD operation, followed by signed JLE in GAME.EXE. Widening
	// before subtraction or comparing the wrapped difference unsigned changes
	// the four-stage cooldown at clock wrap and for a future LastUsed value.
	return int32(h.Clock()-entry.LastUsed) > 4
}

func questMapFile4D0F60(h questMapFileHooks4D0F60) unsafe.Pointer {
	count := int32(h.Count())
	if count == 0 {
		return nil
	}
	if count == 1 {
		return h.Name(0)
	}
	if count <= 0 {
		return h.Name(h.Random(0, count-1))
	}
	var maximum int32
	for index := int32(0); index < count; index++ {
		if uses := int32(h.Entry(index).Uses); uses > maximum {
			maximum = uses
		}
	}
	if maximum == 0 {
		return h.Name(h.Random(0, count-1))
	}
	equal := true
	for index := int32(0); index < count; index++ {
		uses := h.Entry(index).Uses
		for other := int32(1); other < count; other++ {
			if uses != h.Entry(other).Uses {
				equal = false
			}
		}
	}
	threshold := maximum
	if equal {
		threshold++ // Preserve the original DWORD increment, including wrap.
	}
	last := h.LastIndex()
	var eligible int32
	for index := int32(0); index < count; index++ {
		if questMapEligible4D0F60(h, index, last, threshold) {
			eligible++
		}
	}
	choice := h.Random(0, eligible-1)
	// 004D1063/004D1078 reload these two globals after the random callback.
	// The usage threshold is retained. Do not cache an eligible-entry list
	// across this boundary or recompute its threshold from live history.
	count = int32(h.Count())
	if count <= 0 {
		return h.Name(choice)
	}
	last = h.LastIndex()
	var ordinal int32
	for index := int32(0); index < count; index++ {
		if questMapEligible4D0F60(h, index, last, threshold) {
			if ordinal == choice {
				return h.Name(index)
			}
			ordinal++
		}
	}
	return h.Name(choice)
}
