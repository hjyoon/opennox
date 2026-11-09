package legacy

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"runtime"
	"unsafe"
)

const gameSettingsRecordOffset4161E0 = uintptr(371380)
const gameSettingsRecordBytes4161E0 = uintptr(58)
const (
	gameSettingsPlayerCount4161E0 = 1 + iota
	gameSettingsPlayerLimit4161E0
	gameSettingsHeadless4161E0
	gameSettingsVideo4161E0
	gameSettingsFlags4161E0
	gameSettingsGameData4161E0
	gameSettingsModeByte4161E0
	gameSettingsCheckFlag4161E0
	gameSettingsEquipmentMask4161E0
	gameSettingsSpellMask4161E0
	gameSettingsItemMask4161E0
	gameSettingsServerName4161E0
	gameSettingsMapName4161E0
)

// Region addresses the actual byte at offset371380; Dirty addresses the real
// C owner by default. Overrides contain real allocated native storage.
type gameSettingsNativeObserver4161E0 struct {
	Region, Dirty unsafe.Pointer
	Word          func(service int, argument int32) (uint32, bool)
	Fill          func(service int, output []byte) bool
	String        func(service int) (unsafe.Pointer, bool)
}
type gameSettingsObserverFrame4161E0 struct {
	value *gameSettingsNativeObserver4161E0
	prior *gameSettingsObserverFrame4161E0
}

var gameSettingsObserver4161E0 *gameSettingsObserverFrame4161E0

func pushGameSettingsNativeObserver4161E0(value *gameSettingsNativeObserver4161E0) func() {
	frame := &gameSettingsObserverFrame4161E0{value: value, prior: gameSettingsObserver4161E0}
	gameSettingsObserver4161E0 = frame
	return func() {
		if gameSettingsObserver4161E0 != frame {
			panic("game settings observer LIFO")
		}
		gameSettingsObserver4161E0 = frame.prior
	}
}
func gameSettingsCurrent4161E0() *gameSettingsNativeObserver4161E0 {
	if f := gameSettingsObserver4161E0; f != nil {
		return f.value
	}
	return nil
}
func gameSettingsAddress4161E0(off uintptr) unsafe.Pointer {
	if o := gameSettingsCurrent4161E0(); o != nil && o.Region != nil {
		return unsafe.Add(o.Region, int64(off)-int64(gameSettingsRecordOffset4161E0))
	}
	return memmap.PtrOff(0x5D4594, off)
}
func gameSettingsDirtyCell4161E0() unsafe.Pointer {
	if o := gameSettingsCurrent4161E0(); o != nil && o.Dirty != nil {
		return o.Dirty
	}
	return gameSettingsActualDirtyCell4161E0()
}
func gameSettingsDirty4161E0() int32                  { return int32(*(*uint32)(gameSettingsDirtyCell4161E0())) }
func gameSettingsMark4161E0()                         { *(*uint32)(gameSettingsDirtyCell4161E0()) = 1 }
func gameSettingsByte4161E0(off uintptr) byte         { return *(*byte)(gameSettingsAddress4161E0(off)) }
func gameSettingsWord4161E0(off uintptr) uint16       { return *(*uint16)(gameSettingsAddress4161E0(off)) }
func gameSettingsDWORD4161E0(off uintptr) uint32      { return *(*uint32)(gameSettingsAddress4161E0(off)) }
func gameSettingsPutByte4161E0(off uintptr, v byte)   { *(*byte)(gameSettingsAddress4161E0(off)) = v }
func gameSettingsPutWord4161E0(off uintptr, v uint16) { *(*uint16)(gameSettingsAddress4161E0(off)) = v }
func gameSettingsPutDWORD4161E0(off uintptr, v uint32) {
	*(*uint32)(gameSettingsAddress4161E0(off)) = v
}
func gameSettingsServiceWord4161E0(service int, arg int32) uint32 {
	if o := gameSettingsCurrent4161E0(); o != nil && o.Word != nil {
		if v, ok := o.Word(service, arg); ok {
			return v
		}
	}
	return gameSettingsDefaultWord4161E0(service, arg)
}
func gameSettingsServiceFill4161E0(service int, out []byte) {
	if o := gameSettingsCurrent4161E0(); o != nil && o.Fill != nil && o.Fill(service, out) {
		return
	}
	gameSettingsDefaultFill4161E0(service, out)
	runtime.KeepAlive(out)
}
func gameSettingsServiceString4161E0(service int) unsafe.Pointer {
	if o := gameSettingsCurrent4161E0(); o != nil && o.String != nil {
		if v, ok := o.String(service); ok {
			return v
		}
	}
	return gameSettingsDefaultString4161E0(service)
}
func gameSettingsChar4161E0(v byte) int32 {
	if gameSettingsPlainCharSigned4161E0 {
		return int32(int8(v))
	}
	return int32(v)
}
func gameSettingsStrncpy4161E0(dst, src unsafe.Pointer, n int) {
	ended := false
	for i := 0; i < n; i++ {
		v := byte(0)
		if !ended {
			v = *(*byte)(unsafe.Add(src, i))
			ended = v == 0
		}
		*(*byte)(unsafe.Add(dst, i)) = v
	}
	runtime.KeepAlive(src)
	runtime.KeepAlive(dst)
}
func gameSettingsStrncmp4161E0(a, b unsafe.Pointer, n int) bool {
	for i := 0; i < n; i++ {
		x, y := *(*byte)(unsafe.Add(a, i)), *(*byte)(unsafe.Add(b, i))
		if x != y {
			return false
		}
		if x == 0 {
			return true
		}
	}
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	return true
}

func sub4161E0Native() int32 {
	var spell [5]uint32
	var item [4]byte
	gameSettingsServiceFill4161E0(gameSettingsSpellMask4161E0, unsafe.Slice((*byte)(unsafe.Pointer(&spell[0])), 20))
	gameSettingsServiceFill4161E0(gameSettingsItemMask4161E0, item[:])
	equipment := gameSettingsServiceWord4161E0(gameSettingsEquipmentMask4161E0, 0)
	count := int32(gameSettingsServiceWord4161E0(gameSettingsPlayerCount4161E0, 0))
	limit := int32(gameSettingsServiceWord4161E0(gameSettingsPlayerLimit4161E0, 0))
	if gameSettingsServiceWord4161E0(gameSettingsHeadless4161E0, 1<<18) != 0 {
		count--
		limit--
	}
	if int32(gameSettingsByte4161E0(371619)) != count {
		gameSettingsPutByte4161E0(371619, byte(count))
		gameSettingsMark4161E0()
	}
	if int32(gameSettingsByte4161E0(371620)) != limit {
		gameSettingsPutByte4161E0(371620, byte(limit))
		gameSettingsMark4161E0()
	}
	if int32(gameSettingsByte4161E0(371618)&0xEF) != int32(gameSettingsServiceWord4161E0(gameSettingsVideo4161E0, 0)) {
		video := byte(gameSettingsServiceWord4161E0(gameSettingsVideo4161E0, 0))
		gameSettingsMark4161E0()
		gameSettingsPutByte4161E0(371618, gameSettingsByte4161E0(371618)&0x80|video)
	}
	flags := uint16(gameSettingsServiceWord4161E0(gameSettingsFlags4161E0, 0))
	cached := int16(gameSettingsWord4161E0(371432))
	if (gameSettingsWord4161E0(371432)^flags)&0xFFF0 != 0 {
		gameSettingsPutWord4161E0(371432, uint16(gameSettingsServiceWord4161E0(gameSettingsFlags4161E0, 0)))
		cached = int16(gameSettingsWord4161E0(371432))
		gameSettingsMark4161E0()
	}
	if int32(gameSettingsWord4161E0(371434)) != int32(int16(gameSettingsServiceWord4161E0(gameSettingsGameData4161E0, int32(cached)))) {
		gameSettingsPutWord4161E0(371434, uint16(gameSettingsServiceWord4161E0(gameSettingsGameData4161E0, int32(int16(gameSettingsWord4161E0(371432))))))
		gameSettingsMark4161E0()
	}
	if gameSettingsByte4161E0(371436) != byte(gameSettingsServiceWord4161E0(gameSettingsModeByte4161E0, int32(int16(gameSettingsWord4161E0(371432))))) {
		gameSettingsPutByte4161E0(371436, byte(gameSettingsServiceWord4161E0(gameSettingsModeByte4161E0, int32(int16(gameSettingsWord4161E0(371432))))))
		gameSettingsMark4161E0()
	}
	var server [16]byte
	gameSettingsStrncpy4161E0(unsafe.Pointer(&server[0]), gameSettingsAddress4161E0(371389), 15)
	server[15] = 0
	if !gameSettingsStrncmp4161E0(unsafe.Pointer(&server[0]), gameSettingsServiceString4161E0(gameSettingsServerName4161E0), 15) {
		p := gameSettingsServiceString4161E0(gameSettingsServerName4161E0)
		gameSettingsStrncpy4161E0(gameSettingsAddress4161E0(371389), p, 15)
		gameSettingsMark4161E0()
	}
	var mapname [12]byte
	gameSettingsStrncpy4161E0(unsafe.Pointer(&mapname[0]), gameSettingsAddress4161E0(371380), 12)
	if !gameSettingsStrncmp4161E0(unsafe.Pointer(&mapname[0]), gameSettingsServiceString4161E0(gameSettingsMapName4161E0), 12) {
		p := gameSettingsServiceString4161E0(gameSettingsMapName4161E0)
		gameSettingsStrncpy4161E0(gameSettingsAddress4161E0(371380), p, 12)
		gameSettingsMark4161E0()
	}
	same := true
	for i := 0; i < 5; i++ {
		if gameSettingsDWORD4161E0(371404+4*uintptr(i)) != spell[i] {
			same = false
			break
		}
	}
	if !same {
		gameSettingsMark4161E0()
		gameSettingsCopy4161E0(gameSettingsAddress4161E0(371404), unsafe.Pointer(&spell[0]), 20)
	}
	if gameSettingsServiceWord4161E0(gameSettingsCheckFlag4161E0, 1) == 0 {
		return gameSettingsDirty4161E0()
	}
	if gameSettingsDWORD4161E0(371428) != equipment {
		gameSettingsPutDWORD4161E0(371428, equipment)
		gameSettingsMark4161E0()
	}
	same = true
	for i := 0; i < 4; i++ {
		if int32(gameSettingsByte4161E0(371424+uintptr(i))) != gameSettingsChar4161E0(item[i]) {
			same = false
			break
		}
	}
	if same {
		return gameSettingsDirty4161E0()
	}
	gameSettingsCopy4161E0(gameSettingsAddress4161E0(371424), unsafe.Pointer(&item[0]), 4)
	gameSettingsMark4161E0()
	return 1
}
func sub4164F0Native()       { *(*uint32)(gameSettingsDirtyCell4161E0()) = 0 }
func sub416580Native() int32 { return int32(gameSettingsDWORD4161E0(371688)) }

// Unsigned index math wraps in uint32. Signed overflow has no C-defined
// equivalence claim; deterministic int32 offsets never narrow native pointers.
func sub4165B0Native() unsafe.Pointer {
	return gameSettingsAddress4161E0(uintptr(uint32(371380) + 58*gameSettingsDWORD4161E0(371688)))
}
func gameSettingsSignedRecord4161E0(index int32) unsafe.Pointer {
	off := int32(uint32(371380) + 58*uint32(index))
	return gameSettingsAddress4161E0(uintptr(int64(off)))
}
func sub4165D0Native(index int32) unsafe.Pointer {
	gameSettingsPutDWORD4161E0(371688, uint32(index))
	return gameSettingsSignedRecord4161E0(index)
}
func sub4165F0Native(from, to int32) int32 {
	gameSettingsCopy4161E0(gameSettingsSignedRecord4161E0(to), gameSettingsSignedRecord4161E0(from), 58)
	return to
}
func sub416630Native() unsafe.Pointer { return gameSettingsAddress4161E0(371616) }
func sub416640Native() unsafe.Pointer { return gameSettingsAddress4161E0(371516) }
func sub416650Native() int32          { return int32(gameSettingsDWORD4161E0(371700)) }

// Current C's if(0) remains a no-op policy, not original broadcast credit.
func sub416690Native()       {}
func sub4169C0Native() int32 { return int32(gameSettingsDWORD4161E0(371704)) }
func sub4169F0Native() unsafe.Pointer {
	p := sub416640Native()
	q := (*byte)(unsafe.Add(p, 100))
	*q &= 0xEF
	return p
}
