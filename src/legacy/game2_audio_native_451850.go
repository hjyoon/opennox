package legacy

import (
	"github.com/opennox/opennox/v1/legacy/timer"
	"unsafe"
)

// Only C-owned pointers are stored in these C-compatible structs. Definitions
// keep their 200-byte scalar representation; native linkage is separate.
type game2AudioEvent struct {
	Next, Prev, RankNext, RankPrev *game2AudioEvent
	Definition                     unsafe.Pointer
	Flags, State, SavedState       uint32
	Samples                        [32]unsafe.Pointer
	SampleCount                    uint32
	Cursor                         int32
	Voice                          *game2AudioVoice
	Group                          timer.TimerGroup
	AudioSample                    unsafe.Pointer
	Serial, Delay                  uint32
	Due                            uint64
	Buffer                         unsafe.Pointer
	Category                       int32
	Order                          [32]uint32
	Remaining, Loops               uint32
	Shuffle                        [32]uint32
	ShuffleRemaining               uint32
}
type game2AudioVoice struct {
	Owner            *game2AudioEvent
	Group            timer.TimerGroup
	Mixed            timer.TimerGroup
	Master           *timer.TimerGroup
	Priority, Repeat int32
	Flags            uint32
	Buffer           unsafe.Pointer
}
type game2AudioHandle struct {
	Event      *game2AudioEvent
	Serial     uint32
	Definition unsafe.Pointer
}
type game2AudioContext struct {
	BaseGroup   timer.TimerGroup
	EventMaster *timer.TimerGroup
	Driver      uintptr
}
type game2AudioDefinition struct {
	Pointer                      unsafe.Pointer
	Name                         unsafe.Pointer
	Ranked                       []*game2AudioEvent
	BucketCategory, BucketVolume int32
	BucketLinked                 bool
}

func (d *game2AudioDefinition) word(i int) uint32   { return *(*uint32)(unsafe.Add(d.Pointer, 4*i)) }
func (d *game2AudioDefinition) set(i int, v uint32) { *(*uint32)(unsafe.Add(d.Pointer, 4*i)) = v }
func (d *game2AudioDefinition) volume() *timer.Timer {
	return (*timer.Timer)(unsafe.Add(d.Pointer, 16))
}
func (d *game2AudioDefinition) sample(i int) int16 { return *(*int16)(unsafe.Add(d.Pointer, 128+2*i)) }
func (d *game2AudioDefinition) setSample(i int, v int16) {
	*(*int16)(unsafe.Add(d.Pointer, 128+2*i)) = v
}

type game2AudioState struct {
	Initialized, Busy  bool
	Definitions        [1023]game2AudioDefinition
	DefinitionMap      map[unsafe.Pointer]*game2AudioDefinition
	Pool               []*game2AudioEvent
	Free               []*game2AudioEvent
	Live               map[*game2AudioEvent]bool
	First, Last        *game2AudioEvent
	Buckets            [6][10][]*game2AudioDefinition
	Frame, BucketFrame uint32
	Master             *timer.TimerGroup
	Bank               *game2AudioBank
	OwnBank            bool
	Context            *game2AudioContext
	OwnContext         bool
}

var game2Audio game2AudioState

func game2AudioDef(p unsafe.Pointer) *game2AudioDefinition { return game2Audio.DefinitionMap[p] }
func game2AudioEnabled() bool                              { return game2AudioFXFlag() != 0 }
func game2AudioInit451850(context unsafe.Pointer, bank unsafe.Pointer) int32 {
	game2AudioShutdown451970()
	s := &game2Audio
	s.DefinitionMap = make(map[unsafe.Pointer]*game2AudioDefinition, 1023)
	for i := range s.Definitions {
		d := &s.Definitions[i]
		// Actual existing C-owned scalar owner, preserving record stride.
		d.Pointer = game2AudioDefinitionPointer(i)
		d.Name = game2AudioSoundName(i)
		s.DefinitionMap[d.Pointer] = d
		game2AudioReset451920(d.Pointer)
	}
	b := game2AudioBankFor(bank)
	ctx := game2AudioContexts[context]
	if ctx == nil || ctx.Driver == 0 || b == nil || b.Driver != ctx.Driver {
		return 0
	}
	s.Bank = b
	s.Context = ctx
	s.Live = make(map[*game2AudioEvent]bool)
	for i := 0; i < 200; i++ {
		p := game2AudioAlloc(unsafe.Sizeof(game2AudioEvent{}))
		if p == nil {
			game2AudioShutdown451970()
			return 0
		}
		e := (*game2AudioEvent)(p)
		s.Pool = append(s.Pool, e)
		s.Free = append(s.Free, e)
	}
	s.Master = (*timer.TimerGroup)(game2AudioAlloc(unsafe.Sizeof(timer.TimerGroup{})))
	if s.Master == nil {
		game2AudioShutdown451970()
		return 0
	}
	s.Master.Init()
	ctx.EventMaster = s.Master
	s.Initialized = true
	game2AudioSetInitialized(true)
	return 1
}
func game2AudioReset451920(p unsafe.Pointer) int32 {
	if p == nil {
		return 0
	}
	for _, i := range []int{0, 1, 2, 14, 15, 19, 20, 48, 18, 17, 25, 26} {
		*(*uint32)(unsafe.Add(p, 4*i)) = 0
	}
	*(*uint32)(unsafe.Add(p, 48)) = 1
	*(*uint32)(unsafe.Add(p, 64)) = 600
	return game2AudioBool((*timer.Timer)(unsafe.Add(p, 16)).Init(0x4000))
}
func game2AudioShutdown451970() {
	s := &game2Audio
	if s.Initialized {
		game2AudioClear4521F0()
		game2AudioPurge452230()
	}
	for _, e := range s.Pool {
		game2AudioFree(unsafe.Pointer(e))
	}
	b, owned := s.Bank, s.OwnBank
	ctx, ownedContext := s.Context, s.OwnContext
	if ctx != nil {
		ctx.EventMaster = nil
	}
	game2AudioFree(unsafe.Pointer(s.Master))
	// Banks and driver voices are owned by Game2AudioCloseBank, not by an
	// arbitrary borrowed initialization argument; all retained events ended.
	*s = game2AudioState{}
	game2AudioSetInitialized(false)
	if owned && b != nil {
		game2AudioCloseBank(b.Token)
	}
	if ownedContext && ctx != nil {
		game2AudioCloseContext(unsafe.Pointer(ctx))
	}
}
func game2AudioAppend(e *game2AudioEvent) {
	s := &game2Audio
	e.Prev = s.Last
	e.Next = nil
	if s.Last != nil {
		s.Last.Next = e
	} else {
		s.First = e
	}
	s.Last = e
}
func game2AudioRelease451FE0(e *game2AudioEvent) unsafe.Pointer {
	s := &game2Audio
	if !s.Live[e] {
		return nil
	}
	if e.Prev != nil {
		e.Prev.Next = e.Next
	} else {
		s.First = e.Next
	}
	if e.Next != nil {
		e.Next.Prev = e.Prev
	} else {
		s.Last = e.Prev
	}
	if d := game2AudioDef(e.Definition); d != nil {
		game2AudioUnrank(d, e)
	}
	e.Serial = 0
	e.Next = nil
	e.Prev = nil
	delete(s.Live, e)
	s.Free = append(s.Free, e)
	// Native opaque freed-slot token. The original returned its private
	// PE32 allocator-header address (event-4), not a scalar success value.
	return unsafe.Pointer(e)
}
func game2AudioUnrank(d *game2AudioDefinition, e *game2AudioEvent) {
	for i, v := range d.Ranked {
		if v == e {
			d.Ranked = append(d.Ranked[:i], d.Ranked[i+1:]...)
			break
		}
	}
	e.RankNext = nil
	e.RankPrev = nil
}
func game2AudioAllocate452300(p unsafe.Pointer) *game2AudioEvent {
	s := &game2Audio
	d := game2AudioDef(p)
	if !s.Initialized || !game2AudioEnabled() || d == nil || d.word(0) == 0 {
		return nil
	}
	if len(s.Free) == 0 {
		game2AudioPurge452230()
	}
	if len(s.Free) == 0 {
		return nil
	}
	i := len(s.Free) - 1
	e := s.Free[i]
	s.Free = s.Free[:i]
	*e = game2AudioEvent{Definition: p}
	e.Group.Init()
	e.Serial = game2AudioNextSerial()
	s.Live[e] = true
	game2AudioAppend(e)
	return e
}
func game2AudioLookup452270(i int32) unsafe.Pointer {
	if !game2Audio.Initialized || i < 0 || i >= 1023 {
		return nil
	}
	return game2Audio.Definitions[i].Pointer
}
func game2AudioDuration4522A0(p unsafe.Pointer) int32 {
	if !game2Audio.Initialized || p == nil {
		return 0
	}
	return int32(*(*uint32)(unsafe.Add(p, 64)))
}
func game2AudioClear4521F0() uintptr {
	if !game2Audio.Initialized {
		return 0
	}
	result := uintptr(1)
	for e := game2Audio.First; e != nil; {
		next := e.Next
		game2AudioStop4523D0(e)
		result = uintptr(game2AudioRelease451FE0(e))
		e = next
	}
	return result // scalar flag or opaque native slot bits; never a Go pointer.
}
func game2AudioPurge452230() unsafe.Pointer {
	// The original initialized-flag-as-pointer return was an accidental PE32
	// scalar alias. Native policy returns the actual terminal list pointer
	// (nil for the native empty sentinel), never unsafe.Pointer(uintptr(1)).
	if !game2Audio.Initialized {
		return nil
	}
	for e := game2Audio.First; e != nil; {
		next := e.Next
		if e.Flags&1 != 0 {
			game2AudioRelease451FE0(e)
		}
		e = next
	}
	return nil
}
func game2AudioBucketReset452010() int32 {
	s := &game2Audio
	for i := range s.Buckets {
		for j := range s.Buckets[i] {
			s.Buckets[i][j] = nil
		}
	}
	s.BucketFrame++
	return int32(s.BucketFrame)
}
func game2AudioBucketRemove452190(d *game2AudioDefinition) {
	if d == nil || !d.BucketLinked {
		return
	}
	a, b := d.BucketCategory, d.BucketVolume
	if a >= 0 && a < 6 && b >= 0 && b < 10 {
		v := game2Audio.Buckets[a][b]
		for i, x := range v {
			if x == d {
				game2Audio.Buckets[a][b] = append(v[:i], v[i+1:]...)
				break
			}
		}
	}
	d.BucketLinked = false
}
func game2AudioBucket452050(e *game2AudioEvent) {
	d := game2AudioDef(e.Definition)
	if d == nil {
		return
	}
	cat := int32(d.word(12)) + e.Category
	vol := int32((e.Group.Timers[0].Current >> 16) / 0x666)
	if d.word(26) == game2Audio.BucketFrame {
		if cat < d.BucketCategory || (cat == d.BucketCategory && vol <= d.BucketVolume) {
			return
		}
		game2AudioBucketRemove452190(d)
	}
	d.set(26, game2Audio.BucketFrame)
	d.set(27, uint32(cat))
	d.set(31, uint32(vol))
	d.BucketCategory, d.BucketVolume = cat, vol
	// Values outside the six-by-ten table were old out-of-bounds C accesses.
	// They are explicitly rejected rather than relocated to another owner.
	if cat < 0 || cat >= 6 || vol < 0 || vol >= 10 {
		return
	}
	game2Audio.Buckets[cat][vol] = append(game2Audio.Buckets[cat][vol], d)
	d.BucketLinked = true
}
func game2AudioFindBucket4521A0(limit int32) *game2AudioDefinition {
	if limit > 6 {
		limit = 6
	}
	for i := int32(0); i < limit; i++ {
		for j := 0; j < 10; j++ {
			if a := game2Audio.Buckets[i][j]; len(a) > 0 {
				return a[0]
			}
		}
	}
	return nil
}
func game2AudioSteal452120(e *game2AudioEvent) int32 {
	d := game2AudioDef(e.Definition)
	if d == nil {
		return 0
	}
	victim := game2AudioFindBucket4521A0(int32(d.word(12)) + e.Category)
	if victim == nil {
		return 0
	}
	game2AudioBucketRemove452190(victim)
	var found int32
	for v := game2Audio.First; v != nil; {
		next := v.Next
		if v.Definition == victim.Pointer {
			game2AudioStop4523D0(v)
			game2AudioRelease451FE0(v)
			found = 1
		}
		v = next
	}
	return found
}
func game2AudioRank451BE0(e *game2AudioEvent) int32 {
	d := game2AudioDef(e.Definition)
	if d == nil {
		return 0
	}
	level := e.Group.Timers[0].Current >> 16
	game2AudioUnrank(d, e)
	i := 0
	for i < len(d.Ranked) {
		other := d.Ranked[i]
		ov := other.Group.Timers[0].Current >> 16
		delta := int32(ov - level)
		if delta < 0 {
			delta = int32(level - ov)
		}
		if uint32(delta) >= (d.word(5)>>16)/10 {
			if ov < level {
				break
			}
		} else {
			if d.word(1)&0x10 != 0 {
				if other.State != 0 {
					break
				}
			} else if other.State == 0 {
				break
			}
		}
		i++
	}
	d.Ranked = append(d.Ranked, nil)
	copy(d.Ranked[i+1:], d.Ranked[i:])
	d.Ranked[i] = e
	d.set(13, d.word(13)+1)
	ret := int32(d.word(14))
	if ret != 0 && int32(d.word(13)) > ret {
		last := d.Ranked[len(d.Ranked)-1]
		game2AudioUnrank(d, last)
		game2AudioStop4523D0(last)
		d.set(13, d.word(13)-1)
		ret = int32(d.word(13))
	}
	return ret
}
func game2AudioStop4523D0(e *game2AudioEvent) int32 {
	if e == nil || e.Flags&1 != 0 {
		return 0
	}
	game2AudioDetach452410(e)
	game2AudioReleaseSamples451F90(e)
	e.State = 4
	e.Serial = 0
	e.Flags |= 1
	return int32(e.Flags)
}
func game2AudioDetach452410(e *game2AudioEvent) unsafe.Pointer {
	if e == nil {
		return nil
	}
	v := e.Voice
	if v != nil && v.Owner == e {
		if e.Flags&2 != 0 {
			game2AudioVoiceStop(v)
		}
		v.Flags &^= 0x10
		v.Owner = nil
		v.Buffer = nil
		v.Master = nil
		e.Voice = nil
	}
	return unsafe.Pointer(v)
}
func game2AudioPlay452490(e *game2AudioEvent) int32 {
	v := e.Voice
	if v == nil || v.Owner != e {
		return 0
	}
	saved := e.Buffer
	v.Buffer = e.Buffer
	e.State = 3
	e.Flags |= 2
	e.Buffer = nil
	if game2AudioVoicePlay(v) == 0 {
		return 1
	}
	e.State = 1
	e.Buffer = saved
	e.Flags &^= 2
	return 0
}
func game2AudioDelay452690(e *game2AudioEvent, delta int64, next int32) int64 {
	e.SavedState = uint32(next)
	e.Due = uint64(delta) + PlatformTicks()
	e.State = 2
	return int64(e.Due)
}
func game2AudioAdvance452510(e *game2AudioEvent) {
	if !game2AudioEnabled() {
		e.State = 4
	}
	for {
		switch e.State {
		case 0:
			if game2AudioSetup452580(e) == 0 {
				game2AudioStop4523D0(e)
			}
			return
		case 2:
			if PlatformTicks() <= e.Due {
				return
			}
			e.State = e.SavedState
		case 4:
			game2AudioStop4523D0(e)
			return
		default:
			return
		}
	}
}
func game2AudioEnd4526D0(v *game2AudioVoice) int32 {
	if v != nil && v.Owner != nil {
		v.Owner.State = 4
	}
	return 0
}
func game2AudioDone4526F0(v *game2AudioVoice) int32 {
	if v == nil || v.Owner == nil {
		return 0
	}
	e := v.Owner
	e.Flags &^= 2
	if e.State != 4 {
		next := int32(4)
		if e.Buffer != nil || e.ShuffleRemaining != 0 {
			next = 1
		} else {
			e.Delay = 0
		}
		if e.Delay != 0 {
			game2AudioDelay452690(e, int64(e.Delay), next)
			e.Delay = 0
			return 0
		}
		e.State = uint32(next)
	}
	return 0
}
func game2AudioVolume452F10(e *game2AudioEvent, volume int32) uint32 {
	if volume < 0 {
		volume = 0
	} else if volume > 100 {
		volume = 100
	}
	d := game2AudioDef(e.Definition)
	if d == nil {
		return 0
	}
	return (uint32(163) * uint32(volume) * (d.word(5) >> 16)) >> 14
}
func game2AudioPan452FA0(pan int32) int32 {
	if pan < -50 {
		pan = -50
	} else if pan > 50 {
		pan = 50
	}
	return pan*8192/50 + 8192
}
func game2AudioVolume452EE0(e *game2AudioEvent, v int32) int32 {
	e.Group.Timers[0].SetRaw(game2AudioVolume452F10(e, v))
	return game2AudioBool(e.Group.Timers[0].Update())
}
func game2AudioVolume452F50(e *game2AudioEvent, v int32) int32 {
	e.Group.Timers[0].SetInterp(game2AudioVolume452F10(e, v))
	return 0
}
func game2AudioPan452F80(e *game2AudioEvent, v int32) unsafe.Pointer {
	e.Group.Timers[2].SetRaw(uint32(game2AudioPan452FA0(v)))
	return nil
}
func game2AudioPan452FE0(e *game2AudioEvent, v int32) int32 {
	e.Group.Timers[2].SetInterp(uint32(game2AudioPan452FA0(v)))
	return 0
}
func game2AudioSetHandle452E90(h *game2AudioHandle, e *game2AudioEvent) unsafe.Pointer {
	h.Event = e
	if e == nil {
		return nil
	}
	h.Serial = e.Serial
	h.Definition = e.Definition
	return e.Definition
}
func game2AudioGetHandle452EB0(h *game2AudioHandle) *game2AudioEvent {
	e := h.Event
	// Stop keeps the slot on the owned list until Release, but invalidates its
	// handle. A zero generation can otherwise collide with Stop's zero serial
	// before blob initialization or after uint32 wrap; keep allocation order.
	if e != nil && (!game2Audio.Live[e] || e.Flags&1 != 0 || h.Definition != e.Definition || h.Serial != e.Serial) {
		h.Event = nil
		return nil
	}
	return e
}
func game2AudioPlayAt452DC0(id, volume, pan, category int32) {
	e := game2AudioAllocate452300(game2AudioLookup452270(id))
	if e == nil {
		return
	}
	game2AudioVolume452EE0(e, volume)
	game2AudioPan452F80(e, pan)
	e.Category = category
	game2AudioAdvance452510(e)
}
func game2AudioFrame4519C0() {
	s := &game2Audio
	if !s.Initialized || s.Busy {
		return
	}
	s.Busy = true
	defer func() { s.Busy = false }()
	game2AudioDrainCompletions()
	if fx := game2AudioFXMaster(); fx != nil {
		fx.Update()
	}
	s.Frame++
	for e := s.First; e != nil; e = e.Next {
		d := game2AudioDef(e.Definition)
		if d.word(25) != s.Frame {
			d.Ranked = nil
			d.set(13, 0)
			d.set(25, s.Frame)
		}
		e.Group.Update()
		if e.State != 4 {
			game2AudioRank451BE0(e)
		}
	}
	for e := s.First; e != nil; e = e.Next {
		game2AudioAdvance452510(e)
	}
	total := int32(0)
	game2AudioBucketReset452010()
	for e := s.First; e != nil; {
		next := e.Next
		if e.Voice == nil || e.Voice.Owner != e {
			game2AudioStop4523D0(e)
		}
		if e.Flags&1 != 0 {
			game2AudioRelease451FE0(e)
		} else {
			total += int32(uint32(33) * (game2AudioDef(e.Definition).word(5) >> 16) >> 14)
			game2AudioBucket452050(e)
		}
		e = next
	}
	target := uint32(0x4000)
	if total > 100 {
		target = 0x190000 / uint32(total)
	}
	s.Master.Timers[0].SetInterp(target)
	s.Master.Update()
	for e := s.First; e != nil; {
		next := e.Next
		if e.State == 1 {
			game2AudioLoad451DC0(e)
			e.Buffer = game2AudioStart451CA0(e)
			for e.Buffer == nil {
				if game2AudioSteal452120(e) == 0 {
					break
				}
				next = e.Next
				game2AudioLoad451DC0(e)
				e.Buffer = game2AudioStart451CA0(e)
			}
			// Preserve the original third selection, including its cursor advance.
			e.Buffer = game2AudioStart451CA0(e)
			if e.Buffer == nil || game2AudioPlay452490(e) == 0 {
				game2AudioStop4523D0(e)
				game2AudioRelease451FE0(e)
			}
		}
		e = next
	}
	if s.Bank != nil {
		for _, v := range s.Bank.Voices {
			game2AudioVoiceRefresh(v)
		}
	}
}
func game2AudioBool(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
