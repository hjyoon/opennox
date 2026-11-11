package legacy

/*
#include <stdlib.h>
#include "game2_audio_native.h"
extern uint32_t dword_587000_126996;
extern uint32_t dword_5d4594_1045432;
extern uint32_t dword_5d4594_1045460;
extern void *dword_587000_127004;
*/
import "C"
import (
	"encoding/binary"
	"fmt"
	"github.com/opennox/libs/ifs"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/timer"
	"io"
	"strings"
	"sync"
	"unsafe"
)

type game2AudioSample struct {
	Token, Data            unsafe.Pointer
	Size                   int
	Name                   string
	Rate, Flags, BlockSize uint32
	References             uint32
}
type game2AudioBank struct {
	Token       unsafe.Pointer
	Driver      uintptr // AIL native opaque handle, not a pointer reconstructed from DWORD.
	Samples     []*game2AudioSample
	Names       map[string]int16
	Voices      []*game2AudioVoice
	Handles     map[*game2AudioVoice]ail.Sample
	Generations map[*game2AudioVoice]uint64
}

var game2AudioMu sync.Mutex
var game2AudioPendingMu sync.Mutex

type game2AudioCompletion struct {
	voice      *game2AudioVoice
	bank       *game2AudioBank
	generation uint64
}

var game2AudioPending []game2AudioCompletion
var game2AudioBanks = make(map[unsafe.Pointer]*game2AudioBank)
var game2AudioContexts = make(map[unsafe.Pointer]*game2AudioContext)
var game2AudioSamples = make(map[unsafe.Pointer]*game2AudioSample)

// Synchronous fixture boundaries. Nil keeps the product's calloc/free and
// existing 200-byte definition owner. Native-pointer comparisons can request
// foreign high allocations without changing the ordinary runtime allocator.
var game2AudioAllocationService func(uintptr) unsafe.Pointer
var game2AudioReleaseService func(unsafe.Pointer)
var game2AudioDefinitionService func(int) unsafe.Pointer

func game2AudioAlloc(n uintptr) unsafe.Pointer {
	if game2AudioAllocationService != nil {
		return game2AudioAllocationService(n)
	}
	return C.calloc(1, C.size_t(n))
}
func game2AudioFree(p unsafe.Pointer) {
	if game2AudioReleaseService != nil {
		game2AudioReleaseService(p)
	} else {
		C.free(p)
	}
}
func game2AudioDefinitionPointer(i int) unsafe.Pointer {
	if game2AudioDefinitionService != nil {
		return game2AudioDefinitionService(i)
	}
	return memmap.PtrOff(0x5d4594, uintptr(840628+200*i))
}
func game2AudioBankFor(p unsafe.Pointer) *game2AudioBank { return game2AudioBanks[p] }

// A device context address is NOT an AIL numeric driver handle. Only a context
// created by this typed adapter can initialize the native public entry.
func Game2AudioOpenContext(driver ail.Driver) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioOpenContext(driver)
}
func game2AudioOpenContext(driver ail.Driver) unsafe.Pointer {
	if driver == 0 {
		return nil
	}
	p := game2AudioAlloc(unsafe.Sizeof(game2AudioContext{}))
	if p == nil {
		return nil
	}
	ctx := (*game2AudioContext)(p)
	ctx.BaseGroup.Init()
	ctx.Driver = uintptr(driver)
	game2AudioContexts[p] = ctx
	return p
}
func Game2AudioCloseContext(p unsafe.Pointer) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioCloseContext(p)
}
func game2AudioCloseContext(p unsafe.Pointer) {
	ctx := game2AudioContexts[p]
	if ctx == nil {
		return
	}
	if game2Audio.Context == ctx {
		game2Audio.OwnContext = false
		game2AudioShutdown451970()
	}
	delete(game2AudioContexts, p)
	game2AudioFree(p)
}
func game2AudioSoundName(i int) unsafe.Pointer {
	return unsafe.Pointer(internCStr(sound.ID(i).String()))
}
func game2AudioFXFlag() uint32        { return uint32(C.dword_587000_126996) }
func game2AudioSetFXFlag(v uint32)    { C.dword_587000_126996 = C.uint32_t(v) }
func game2AudioSetInitialized(v bool) { C.dword_5d4594_1045432 = C.uint32_t(game2AudioBool(v)) }
func game2AudioNextSerial() uint32 {
	p := memmap.PtrUint32(0x587000, 127000)
	v := *p
	*p = v + 1
	return v
}
func game2AudioFXMaster() *timer.TimerGroup {
	if p := unsafe.Pointer(C.dword_587000_127004); p != nil {
		return (*timer.TimerGroup)(p)
	}
	return nil // absent FX is skipped; adaptive EventMaster is a separate stage.
}
func game2AudioReadFile(path string) ([]byte, error) {
	f, e := ifs.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	return io.ReadAll(f)
}

// Game2AudioOpenBank uses the same GABA v1/v2 sound index and native AIL
// driver used by audio_effects.go. C pool/driver structs are never interpreted.
// The caller owns this bank until Game2AudioCloseBank; initialization borrows it.
func Game2AudioOpenBank(driver ail.Driver, base string) (unsafe.Pointer, error) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioOpenBank(driver, base)
}
func game2AudioOpenBank(driver ail.Driver, base string) (unsafe.Pointer, error) {
	idx, e := game2AudioReadFile(strings.TrimSuffix(base, ".bag") + ".idx")
	if e != nil {
		return nil, e
	}
	data, e := game2AudioReadFile(strings.TrimSuffix(base, ".bag") + ".bag")
	if e != nil {
		return nil, e
	}
	return game2AudioOpenBankData(driver, idx, data)
}
func game2AudioOpenBankData(driver ail.Driver, idx, data []byte) (unsafe.Pointer, error) {
	if driver == 0 {
		return nil, fmt.Errorf("audio driver is zero")
	}
	if len(idx) < 12 || string(idx[:4]) != "GABA" {
		return nil, fmt.Errorf("invalid audio index header")
	}
	v := binary.LittleEndian.Uint32(idx[4:])
	stride := uint64(0)
	if v == 1 {
		stride = 32
	} else if v == 2 {
		stride = 36
	} else {
		return nil, fmt.Errorf("unsupported audio index version %d", v)
	}
	count := uint64(binary.LittleEndian.Uint32(idx[8:]))
	if count > 32768 || 12+count*stride > uint64(len(idx)) {
		return nil, fmt.Errorf("truncated or oversized audio index")
	}
	b := &game2AudioBank{Driver: uintptr(driver), Names: make(map[string]int16), Handles: make(map[*game2AudioVoice]ail.Sample), Generations: make(map[*game2AudioVoice]uint64)}
	b.Token = game2AudioAlloc(1)
	if b.Token == nil {
		return nil, fmt.Errorf("audio bank allocation")
	}
	game2AudioBanks[b.Token] = b
	fail := func(e error) (unsafe.Pointer, error) { game2AudioCloseBank(b.Token); return nil, e }
	for i := uint64(0); i < count; i++ {
		r := idx[12+i*stride : 12+(i+1)*stride]
		name := r[:16]
		for j, c := range name {
			if c == 0 {
				name = name[:j]
				break
			}
		}
		off, n := uint64(binary.LittleEndian.Uint32(r[16:])), uint64(binary.LittleEndian.Uint32(r[20:]))
		if len(name) == 0 || off+n > uint64(len(data)) {
			return fail(fmt.Errorf("audio index sample %d range", i))
		}
		s := &game2AudioSample{Name: string(name), Size: int(n), Rate: binary.LittleEndian.Uint32(r[24:]), Flags: binary.LittleEndian.Uint32(r[28:])}
		if stride == 36 {
			s.BlockSize = binary.LittleEndian.Uint32(r[32:])
		}
		if s.Flags&8 != 0 && (s.BlockSize == 0 || n%uint64(s.BlockSize) != 0) {
			return fail(fmt.Errorf("invalid ADPCM sample %d", i))
		}
		s.Token = game2AudioAlloc(1)
		s.Data = game2AudioAlloc(uintptr(n) + 1)
		if s.Token == nil || s.Data == nil {
			game2AudioFree(s.Token)
			game2AudioFree(s.Data)
			return fail(fmt.Errorf("audio sample allocation"))
		}
		copy(unsafe.Slice((*byte)(s.Data), int(n)), data[off:off+n])
		b.Names[strings.ToLower(s.Name)] = int16(i)
		b.Samples = append(b.Samples, s)
		game2AudioSamples[s.Token] = s
	}
	// The original event pool holds 200 events; the existing platform backend
	// supplies sixteen native voices. Each voice is C-owned and independently
	// holds a real AIL sample handle in this Go sidecar.
	for i := 0; i < 16; i++ {
		h := driver.AllocateSample()
		if h == 0 {
			return fail(fmt.Errorf("allocate audio voice %d", i))
		}
		p := game2AudioAlloc(unsafe.Sizeof(game2AudioVoice{}))
		if p == nil {
			h.Release()
			return fail(fmt.Errorf("audio voice allocation"))
		}
		voice := (*game2AudioVoice)(p)
		voice.Group.Init()
		voice.Mixed.Init()
		b.Voices = append(b.Voices, voice)
		b.Handles[voice] = h
	}
	return b.Token, nil
}
func Game2AudioCloseBank(p unsafe.Pointer) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioCloseBank(p)
}
func game2AudioCloseBank(p unsafe.Pointer) {
	b := game2AudioBanks[p]
	if b == nil {
		return
	}
	if game2Audio.Bank == b {
		game2Audio.OwnBank = false
		game2AudioShutdown451970()
	}
	for _, v := range b.Voices {
		h := b.Handles[v]
		h.RegisterEOSCallback(nil)
		h.End()
		h.Release()
		game2AudioFree(unsafe.Pointer(v))
	}
	for _, s := range b.Samples {
		delete(game2AudioSamples, s.Token)
		game2AudioFree(s.Data)
		game2AudioFree(s.Token)
	}
	delete(game2AudioBanks, p)
	game2AudioFree(b.Token)
}

// Game2AudioInitialize is the native Go startup adapter. It accepts the actual
// AIL handle, never the old packed service object. Root calls it after opening
// the native audio device; errors leave the subsystem disabled.
func Game2AudioInitialize(driver ail.Driver, base string) error {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	old := game2Audio.Bank
	p, e := game2AudioOpenBank(driver, base)
	if e != nil {
		return e
	}
	ctx := game2AudioOpenContext(driver)
	if ctx == nil {
		game2AudioCloseBank(p)
		return fmt.Errorf("audio device context allocation")
	}
	if game2AudioInit451850(ctx, p) == 0 {
		game2AudioCloseBank(p)
		game2AudioCloseContext(ctx)
		return fmt.Errorf("initialize audio event service")
	}
	game2Audio.OwnBank = true
	game2Audio.OwnContext = true
	if old != nil {
		game2AudioCloseBank(old.Token)
	}
	return nil
}
func game2AudioSampleLookup(name string) int16 {
	if b := game2Audio.Bank; b != nil {
		if i, ok := b.Names[strings.ToLower(name)]; ok {
			return i
		}
	}
	return -1
}
func game2AudioSampleAt(i int16) unsafe.Pointer {
	b := game2Audio.Bank
	if b == nil || i < 0 || int(i) >= len(b.Samples) {
		return nil
	}
	return b.Samples[i].Token
}
func game2AudioRetain(p unsafe.Pointer) {
	if s := game2AudioSamples[p]; s != nil {
		s.References++
	}
}
func game2AudioRelease(p unsafe.Pointer) {
	if s := game2AudioSamples[p]; s != nil && s.References != 0 {
		s.References--
	}
}
func game2AudioVoiceStop(v *game2AudioVoice) {
	if b := game2Audio.Bank; b != nil {
		b.Generations[v]++
		if h := b.Handles[v]; h != 0 {
			h.End()
		}
	}
	game2AudioEnd4526D0(v)
	v.Buffer = nil
	v.Flags &^= 5
}
func game2AudioVoicePlay(v *game2AudioVoice) int32 {
	if v.Flags&5 != 0 {
		return -2146500608
	}
	s := game2AudioSamples[v.Buffer]
	b := game2Audio.Bank
	if s == nil || b == nil {
		return -2147024896
	}
	h := b.Handles[v]
	if h == 0 {
		return -2147221504
	}
	b.Generations[v]++
	generation := b.Generations[v]
	h.RegisterEOSCallback(func() {
		game2AudioPendingMu.Lock()
		game2AudioPending = append(game2AudioPending, game2AudioCompletion{v, b, generation})
		game2AudioPendingMu.Unlock()
	})
	h.Init()
	typ := int32(0)
	if s.Flags&1 != 0 {
		typ |= 2
	}
	if s.Flags&4 != 0 {
		typ |= 1
	}
	if s.Flags&8 != 0 {
		typ = 5
		if s.Flags&1 != 0 {
			typ = 7
		}
	}
	h.SetType(typ, 0)
	if s.Flags&8 != 0 {
		h.SetADPCMBlockSize(s.BlockSize)
	}
	game2AudioUpdateMixSources(v)
	game2AudioMixVoice4BD840(v)
	gain, pan, rate := game2AudioMixedSampleParameters(v, s.Rate)
	h.SetVolume(gain)
	h.SetPan(pan)
	h.SetPlaybackRate(rate)
	ready := h.BufferReady()
	if ready < 0 {
		return -2147221504
	}
	h.LoadBuffer(uint32(ready), unsafe.Slice((*byte)(s.Data), s.Size))
	v.Flags |= 1
	return 0
}
func game2AudioVoiceRefresh(v *game2AudioVoice) {
	b := game2Audio.Bank
	if b == nil || v.Owner == nil || v.Flags&1 == 0 {
		return
	}
	h := b.Handles[v]
	if h == 0 {
		return
	}
	game2AudioUpdateMixSources(v)
	game2AudioMixVoice4BD840(v)
	if s := game2AudioSamples[v.Buffer]; s != nil {
		gain, pan, rate := game2AudioMixedSampleParameters(v, s.Rate)
		h.SetVolume(gain)
		h.SetPan(pan)
		h.SetPlaybackRate(rate)
	}
}
func game2AudioUpdateMixSources(v *game2AudioVoice) {
	v.Group.Update()
	if v.Owner != nil {
		v.Owner.Group.Update()
	}
	if v.Master != nil {
		v.Master.Update()
	}
	if ctx := game2Audio.Context; ctx != nil {
		ctx.BaseGroup.Update()
		if ctx.EventMaster != nil {
			ctx.EventMaster.Update()
		}
	}
}

// Exact composition order of GAME3_1.c:4BD840. TimerGroup.Mix already preserves
// the wrapping gain/pitch multiplications and centered, clipped pan operation.
// Only voice and owner updated bits are cleared here, as in that function.
func game2AudioMixVoice4BD840(v *game2AudioVoice) {
	v.Mixed.Init()
	v.Mixed.Mix(&v.Group)
	v.Group.ClearUpdated()
	if v.Owner != nil {
		v.Mixed.Mix(&v.Owner.Group)
		v.Owner.Group.ClearUpdated()
	}
	if v.Master != nil {
		v.Mixed.Mix(v.Master)
	}
	if ctx := game2Audio.Context; ctx != nil {
		v.Mixed.Mix(&ctx.BaseGroup)
		if ctx.EventMaster != nil {
			v.Mixed.Mix(ctx.EventMaster)
		}
	}
}
func game2AudioMixedSampleParameters(v *game2AudioVoice, rate uint32) (int, int, int) {
	// Actual 43F060 consumers read the composed group, not individual inputs.
	return int(uint32(127) * (v.Mixed.Timers[0].Current >> 16) >> 14), int(uint32(127) * (v.Mixed.Timers[2].Current >> 16) >> 14), int(timer.ScalePercentageNative486640(&v.Mixed, int32(rate)))
}
func game2AudioAcquire452810(priority int32, flags int8) *game2AudioVoice {
	b := game2Audio.Bank
	if b == nil {
		return nil
	}
	var chosen, pending *game2AudioVoice
	activePriority, pendingPriority := int32(127), int32(127)
	level := uint32(0xffffffff)
	for _, v := range b.Voices {
		if v.Flags&0x15 == 0 {
			chosen = v
			break
		}
		if v.Flags&1 != 0 {
			if v.Priority < activePriority || (v.Priority == activePriority && v.Mixed.Timers[0].Current < level && level-v.Mixed.Timers[0].Current >= 0x666) {
				chosen = v
				activePriority = v.Priority
				level = v.Mixed.Timers[0].Current
			}
		}
		if v.Flags&1 == 0 && v.Priority < pendingPriority {
			pending = v
			pendingPriority = v.Priority
		}
	}
	if pending != nil && (chosen == nil || pendingPriority <= activePriority) {
		chosen = pending
	}
	if chosen == nil || chosen.Flags&0x15 != 0 && chosen.Priority > priority {
		return nil
	}
	game2AudioVoiceStop(chosen)
	chosen.Master = game2AudioFXMaster()
	chosen.Priority = priority
	chosen.Repeat = 0
	if flags&1 != 0 {
		chosen.Repeat = -1
	}
	chosen.Group.Timers[0].SetRaw(0x4000)
	return chosen
}
func game2AudioRandom(lo, hi int32, line int) int32 { return int32(game2AudioRandomC(lo, hi, line)) }
func game2AudioMask453080(bit int8) int32 {
	mask := uint32(1) << (uint8(bit) & 31)
	return game2AudioMaskC(mask, C.dword_5d4594_1045460 != 0)
}

// No raw offset272 getter is used by the new path: these functions expose the
// native event's actual C-owned sample slot for any future typed integration.
func Game2AudioEventSample(p unsafe.Pointer) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	if !game2Audio.Live[(*game2AudioEvent)(p)] {
		return nil
	}
	return (*game2AudioEvent)(p).AudioSample
}
func Game2AudioEventSetSample(p, value unsafe.Pointer) bool {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	if !game2Audio.Live[(*game2AudioEvent)(p)] {
		return false
	}
	(*game2AudioEvent)(p).AudioSample = value
	return true
}
func game2AudioDrainCompletions() {
	game2AudioPendingMu.Lock()
	pending := game2AudioPending
	game2AudioPending = nil
	game2AudioPendingMu.Unlock()
	for _, p := range pending {
		b := game2Audio.Bank
		if b != p.bank || b == nil || b.Generations[p.voice] != p.generation {
			continue
		}
		v := p.voice
		if v.Owner == nil {
			continue
		}
		v.Flags &^= 5
		game2AudioNeedBuffer452770(v)
		if b.Generations[v] == p.generation {
			game2AudioDone4526F0(v)
		}
	}
}
