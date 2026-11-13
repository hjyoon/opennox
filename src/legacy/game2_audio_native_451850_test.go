package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/timer"
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func game2AudioTestSetup(t *testing.T) *game2AudioDefinition {
	t.Helper()
	t.Cleanup(game2AudioHighFixtureMemory())
	oldTicks, oldTimer := PlatformTicks, timer.PlatformTicks
	PlatformTicks = func() uint64 { return 0x100000005 }
	timer.PlatformTicks = PlatformTicks
	oldFlag := game2AudioFixtureEnabled(-1)
	b := &game2AudioBank{Driver: 1, Names: map[string]int16{"sample0": 0, "sample1": 1}, Generations: make(map[*game2AudioVoice]uint64)}
	b.Token = game2AudioAlloc(1)
	if b.Token == nil {
		t.Fatal("bank allocation")
	}
	game2AudioBanks[b.Token] = b
	for i := 0; i < 2; i++ {
		s := &game2AudioSample{Token: game2AudioAlloc(1), Data: game2AudioAlloc(32), Size: 32, Name: fmt.Sprintf("sample%d", i)}
		if s.Token == nil || s.Data == nil {
			t.Fatal("sample allocation")
		}
		b.Samples = append(b.Samples, s)
		game2AudioSamples[s.Token] = s
	}
	ctx := Game2AudioOpenContext(1)
	if ctx == nil || game2AudioFixtureInit(ctx, b.Token) != 1 {
		t.Fatal("public initializer with distinct typed device context")
	}
	game2AudioFixtureEnabled(1)
	t.Cleanup(func() {
		game2AudioFixtureShutdown()
		Game2AudioCloseBank(b.Token)
		Game2AudioCloseContext(ctx)
		game2AudioFixtureEnabled(oldFlag)
		PlatformTicks = oldTicks
		timer.PlatformTicks = oldTimer
	})
	return &game2Audio.Definitions[1]
}
func game2AudioTestFile(t *testing.T, data []byte) *binfile.MemFile {
	t.Helper()
	p, _ := alloc.MallocHigh(uintptr(len(data) + 32))
	if p == nil {
		t.Fatal("input allocation")
	}
	copy(unsafe.Slice((*byte)(p), len(data)), data)
	f := binfile.NewMemFile(p, len(data))
	t.Cleanup(f.Free)
	return f
}
func game2AudioTestString(s string) []byte {
	if len(s) > 255 {
		panic("string8")
	}
	return append([]byte{byte(len(s))}, []byte(s)...)
}
func game2AudioTestWord(v uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return b[:]
}
func game2AudioTestShort(v int16) []byte {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], uint16(v))
	return b[:]
}

func TestGame2AudioNativeScalarCReferences451850(t *testing.T) {
	game2AudioTestSetup(t)
	for _, v := range []int32{math.MinInt32, -100, -51, -50, -49, -1, 0, 1, 49, 50, 51, 100, math.MaxInt32} {
		t.Run(fmt.Sprint(v), func(t *testing.T) {
			goValue := game2AudioPan452FA0(v)
			oldC := game2AudioFixturePan(true, v)
			publicC := game2AudioFixturePan(false, v)
			if goValue != oldC || publicC != oldC {
				t.Fatalf("signed C division: Go %d C %d old %d", goValue, publicC, oldC)
			}
		})
	}
}

func TestGame2AudioNativeLayoutsAndLifetime451850(t *testing.T) {
	d := game2AudioTestSetup(t)
	d.set(0, 1)
	c := game2AudioFixtureLayout()
	e := game2AudioEvent{}
	v := game2AudioVoice{}
	h := game2AudioHandle{}
	ctx := game2AudioContext{}
	want := []uintptr{unsafe.Sizeof(e), unsafe.Sizeof(v), unsafe.Sizeof(h), unsafe.Offsetof(e.Definition), unsafe.Offsetof(e.SampleCount), unsafe.Offsetof(e.Voice), unsafe.Offsetof(e.Group), unsafe.Offsetof(e.AudioSample), unsafe.Offsetof(e.Serial), unsafe.Offsetof(e.Due), unsafe.Offsetof(e.Buffer), unsafe.Offsetof(e.Order), unsafe.Offsetof(e.Remaining), unsafe.Offsetof(e.ShuffleRemaining), unsafe.Sizeof(uintptr(0)), unsafe.Offsetof(h.Definition), unsafe.Sizeof(ctx), unsafe.Offsetof(ctx.Driver), unsafe.Offsetof(ctx.EventMaster), unsafe.Offsetof(ctx.BaseGroup), unsafe.Offsetof(v.Mixed)}
	for i := range want {
		if c[i] != want[i] {
			t.Fatalf("C/Go native layout %d=%d want%d", i, c[i], want[i])
		}
	}
	a, b := game2AudioFixtureAllocate(d.Pointer), game2AudioAllocate452300(d.Pointer)
	if a == nil || b == nil || a.Next != b || b.Prev != a {
		t.Fatal("actual full-pointer C-owned linked list")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (game2AudioFixturePointer(unsafe.Pointer(a)) <= math.MaxUint32 || game2AudioFixturePointer(d.Pointer) <= math.MaxUint32) {
		t.Fatal("native pointer witness did not exceed 32bits")
	}
	hp := (*game2AudioHandle)(game2AudioAlloc(unsafe.Sizeof(h)))
	defer game2AudioFree(unsafe.Pointer(hp))
	if game2AudioFixtureSetHandle(hp, a) != d.Pointer || game2AudioFixtureGetHandle(hp) != a {
		t.Fatal("typed public handle ABI")
	}
	for i := 0; i < 4; i++ {
		runtime.GC()
	}
	if hp.Event != a || hp.Definition != d.Pointer || a.Next != b || b.Prev != a {
		t.Fatal("full native pointers after GC")
	}
	a.Flags = 0x12345600
	if game2AudioFixtureStop(a) != 0x12345601 || a.State != 4 || a.Serial != 0 {
		t.Fatal("stop preserves high flag bytes and invalidates serial")
	}
	if game2AudioFixtureGetHandle(hp) != nil || hp.Event != nil {
		t.Fatal("stale handle")
	}
	if game2AudioFixtureRelease(a) != unsafe.Pointer(a) || b.Prev != nil || game2Audio.First != b {
		t.Fatal("cached successor and opaque native freed slot")
	}
	// No dereference of returned freed slot; a reused slot has fresh serial.
	reused := game2AudioFixtureAllocate(d.Pointer)
	if reused != a || reused.Serial == hp.Serial {
		t.Fatal("pool reuse generation")
	}
	var pinned game2AudioEvent
	pinned.Definition = d.Pointer
	pinned.Serial = 0x76543210
	game2Audio.Live[&pinned] = true
	var pin runtime.Pinner
	pin.Pin(&pinned)
	func() {
		defer pin.Unpin()
		defer delete(game2Audio.Live, &pinned)
		defer func() { *hp = game2AudioHandle{} }()
		game2AudioFixtureSetHandle(hp, &pinned)
		for i := 0; i < 4; i++ {
			runtime.GC()
		}
		if game2AudioFixtureGetHandle(hp) != &pinned {
			t.Fatal("individually pinned Go owner public C handle")
		}
	}()
}

func TestGame2AudioNativeSchedulingAndSelection451850(t *testing.T) {
	d := game2AudioTestSetup(t)
	d.set(0, 1)
	d.set(48, 2)
	d.setSample(0, 0)
	d.setSample(1, 1)
	e := game2AudioAllocate452300(d.Pointer)
	if game2AudioLoadSample451F30(e, 0) != 1 || game2AudioLoadSample451F30(e, 1) != 2 {
		t.Fatal("native sample handle count")
	}
	if game2AudioStart451CA0(e) != game2Audio.Bank.Samples[0].Token || game2AudioNext451CF0(e) != game2Audio.Bank.Samples[1].Token || game2AudioNext451CF0(e) != nil {
		t.Fatal("ordered real pointer selection")
	}
	d.set(1, 1)
	d.set(15, 2)
	if game2AudioStart451CA0(e) == nil || game2AudioNext451CF0(e) == nil || game2AudioNext451CF0(e) == nil || e.Loops != 1 {
		t.Fatal("loop counter at list exhaustion")
	}
	if game2AudioFixtureDelay(e, 0x100000007, 3) != 0x20000000c || e.Due != 0x20000000c || e.State != 2 || e.SavedState != 3 {
		t.Fatal("full64 clock and delay")
	}
	PlatformTicks = func() uint64 { return e.Due }
	game2AudioAdvance452510(e)
	if e.State != 2 {
		t.Fatal("inclusive due boundary")
	}
	PlatformTicks = func() uint64 { return e.Due + 1 }
	game2AudioAdvance452510(e)
	if e.State != 3 {
		t.Fatal("current clock after due")
	}
	d.volume().Current = 0xffffffff
	for _, v := range []int32{-1, 0, 1, 100, 101, math.MaxInt32} {
		x := v
		if x < 0 {
			x = 0
		} else if x > 100 {
			x = 100
		}
		want := uint32(163) * uint32(x) * (d.word(5) >> 16) >> 14
		if game2AudioVolume452F10(e, v) != want {
			t.Fatal("uint32 wrapping volume")
		}
	}
	game2AudioReleaseSamples451F90(e)
	for _, s := range game2Audio.Bank.Samples {
		if s.References != 0 {
			t.Fatal("sample retain/release balance")
		}
	}
}

func TestGame2AudioNativeRankingBucketsAndCallbacks451850(t *testing.T) {
	d := game2AudioTestSetup(t)
	d.set(0, 1)
	d.set(14, 1)
	a, b := game2AudioAllocate452300(d.Pointer), game2AudioAllocate452300(d.Pointer)
	a.Group.Timers[0].Current = 0x10000000
	b.Group.Timers[0].Current = 0x20000000
	game2AudioRank451BE0(a)
	game2AudioRank451BE0(b)
	if a.State != 4 || a.Flags&1 == 0 || len(d.Ranked) != 1 || d.Ranked[0] != b || d.word(13) != 1 {
		t.Fatal("volume sort, capped tail removal and native pointers")
	}
	game2AudioBucketReset452010()
	game2AudioBucket452050(b)
	if game2AudioFindBucket4521A0(2) != d {
		t.Fatal("first eligible category/bucket")
	}
	game2AudioBucketRemove452190(d)
	if game2AudioFindBucket4521A0(2) != nil {
		t.Fatal("unlink definition")
	}
	vp := game2AudioAlloc(unsafe.Sizeof(game2AudioVoice{}))
	if vp == nil {
		t.Fatal("voice allocation")
	}
	defer game2AudioFree(vp)
	v := (*game2AudioVoice)(vp)
	v.Group.Init()
	v.Owner = b
	b.Voice = v
	b.State = 3
	b.Flags = 2
	b.Buffer = game2Audio.Bank.Samples[0].Token
	b.Delay = 77
	game2AudioDone4526F0(v)
	if b.State != 2 || b.SavedState != 1 || b.Due != 0x100000052 || b.Delay != 0 || b.Flags&2 != 0 {
		t.Fatal("actual voice owner callback full64 scheduling")
	}
	game2AudioEnd4526D0(v)
	if b.State != 4 {
		t.Fatal("end callback actual owner")
	}
	b.Flags = 0
	b.State = 3
	v.Flags = 0x10
	if game2AudioDetach452410(b) != vp || b.Voice != nil || v.Owner != nil || v.Master != nil || v.Flags&0x10 != 0 {
		t.Fatal("owner identity and detach clearing")
	}
	game2AudioSetFXFlag(0)
	game2AudioAdvance452510(b)
	if b.State != 4 || b.Flags&1 == 0 {
		t.Fatal("live FX disable")
	}
}

func TestGame2AudioNativeParserCReferences452890(t *testing.T) {
	d := game2AudioTestSetup(t)
	name := sound.ID(1).String()
	avnt := game2AudioTestString(name)
	avnt = append(avnt, 1, 2, 2, 0x1f, 3, 99, 4, 3, 5, 2, 6, 0x80, 0x7f, 7)
	avnt = append(avnt, game2AudioTestString("sample0")...)
	avnt = append(avnt, game2AudioTestString("missing")...)
	avnt = append(avnt, game2AudioTestString("sample1")...)
	avnt = append(avnt, 0, 8)
	avnt = append(avnt, game2AudioTestWord(0x80000000)...)
	avnt = append(avnt, game2AudioTestWord(0xffffffff)...)
	avnt = append(avnt, 9)
	avnt = append(avnt, game2AudioTestShort(123)...)
	avnt = append(avnt, 10)
	avnt = append(avnt, game2AudioTestShort(-32768)...)
	avnt = append(avnt, 0)
	aud := game2AudioTestString(name)
	aud = append(aud, game2AudioTestShort(-1234)...)
	aud = append(aud, 100)
	aud = append(aud, game2AudioTestShort(25)...)
	aud = append(aud, 0x80, 0x81, 0x7f, 2)
	aud = append(aud, game2AudioTestString("sample0.wav")...)
	aud = append(aud, game2AudioTestString("missing.wav")...)
	aud = append(aud, game2AudioTestString("sample1.wav")...)
	aud = append(aud, 0)
	cases := []struct {
		name string
		kind int
		data []byte
	}{{"all-avnt-properties", 0, avnt}, {"aud-signed-fields", 1, aud}, {"skip-avnt", 2, avnt}, {"skip-inner", 3, avnt[1+len(name):]}, {"unknown-property", 0, append(game2AudioTestString(name), 0x7f)}, {"missing-sound", 0, append(game2AudioTestString("not-a-real-sound"), 1, 99, 0)}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fOld := game2AudioTestFile(t, tc.data)
			fNew := game2AudioTestFile(t, tc.data)
			scratchOld := game2AudioAlloc(288)
			scratchNew := game2AudioAlloc(288)
			defer game2AudioFree(scratchOld)
			defer game2AudioFree(scratchNew)
			for i := range unsafe.Slice((*byte)(d.Pointer), 200) {
				unsafe.Slice((*byte)(d.Pointer), 200)[i] = 0xa7
			}
			game2AudioReset451920(d.Pointer)
			game2AudioFixtureReferenceReset(name)
			a := game2AudioFixtureParse(tc.kind, true, fOld, scratchOld)
			b := game2AudioFixtureParse(tc.kind, false, fNew, scratchNew)
			if a != b || len(fOld.Data()) != len(fNew.Data()) {
				t.Fatalf("return/cursor old%d new%d remaining%d/%d", a, b, len(fOld.Data()), len(fNew.Data()))
			}
			if !bytes.Equal(game2AudioFixtureReferenceWords(), unsafe.Slice((*byte)(d.Pointer), 200)) {
				t.Fatal("all200 scalar bytes versus original named C parser")
			}
			if tc.kind < 2 && !bytes.Equal(unsafe.Slice((*byte)(scratchOld), 288), unsafe.Slice((*byte)(scratchNew), 288)) {
				t.Fatal("all scratch bytes including NUL and canaries")
			}
		})
	}
}

func TestGame2AudioNativeMalformedPolicy452890(t *testing.T) {
	d := game2AudioTestSetup(t)
	name := sound.ID(1).String()
	data := append(game2AudioTestString(name), 8, 1, 2, 3, 4, 5, 6, 7, 8, 0)
	for n := 0; n < len(data); n++ {
		t.Run(fmt.Sprintf("short-%d", n), func(t *testing.T) {
			d.set(0, 0)
			f := game2AudioTestFile(t, data[:n])
			if game2AudioFixtureParse(0, false, f, nil) != 0 || d.word(0) != 0 || len(f.Data()) != 0 {
				t.Fatal("short read rejected after consuming only available bytes")
			}
		})
	}
	aud := game2AudioTestString(name)
	aud = append(aud, 0, 0, 100, 1, 0, 0, 0, 0, 1)
	for i := 0; i < 40; i++ {
		aud = append(aud, game2AudioTestString("sample0")...)
	}
	aud = append(aud, 0)
	f := game2AudioTestFile(t, aud)
	canary := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Add(d.Pointer, 200)), 200)...)
	if game2AudioFixtureParse(1, false, f, nil) != 1 || d.word(48) != 32 || len(f.Data()) != 0 {
		t.Fatal("bounded32 sample slots and complete cursor")
	}
	if !bytes.Equal(canary, unsafe.Slice((*byte)(unsafe.Add(d.Pointer, 200)), 200)) {
		t.Fatal("next definition changed by oversized sample list")
	}
}

func TestGame2AudioNativeContextAndConditionalMirrors451850(t *testing.T) {
	d := game2AudioTestSetup(t)
	ctx := game2Audio.Context
	bank := game2Audio.Bank
	if ctx == nil || ctx.EventMaster != game2Audio.Master || ctx.Driver != 1 || game2AudioFixturePointer(unsafe.Pointer(ctx)) == ctx.Driver {
		t.Fatal("full context address is distinct from numeric AIL handle")
	}
	name := sound.ID(1).String()
	input := append(game2AudioTestString(name), 7)
	input = append(input, game2AudioTestString("sample0")...)
	input = append(input, 0, 0)
	before := append([]byte(nil), input...)
	if n, err := Game2AudioReadAVNTData(input); err != nil || n != len(input) || d.word(0) != 1 || d.word(48) != 1 || d.sample(0) != 0 {
		t.Fatalf("conditional live parser mirror: n%d err%v", n, err)
	}
	if !bytes.Equal(before, input) {
		t.Fatal("mirror modified caller input")
	}
	Game2AudioResetDefinitions()
	if d.word(0) != 0 || d.word(48) != 0 || game2Audio.Bank != bank || game2Audio.Context != ctx || ctx.EventMaster != game2Audio.Master {
		t.Fatal("definition reset changed native resource owners")
	}
	Game2AudioShutdownNative()
	if ctx.EventMaster != nil || game2Audio.Initialized {
		t.Fatal("borrowed context master reset on shutdown")
	}
	if n, err := Game2AudioReadAVNTData(input); n != 0 || err != nil {
		t.Fatal("inactive legacy mirror must not activate or duplicate production FX")
	}
	unregistered := game2AudioAlloc(16)
	defer game2AudioFree(unregistered)
	if game2AudioFixtureInit(unregistered, bank.Token) != 0 || game2Audio.Initialized {
		t.Fatal("unregistered old context not reinterpreted as driver handle")
	}
}

func game2AudioTestGroup(g *timer.TimerGroup, values [3]uint32) {
	g.Init()
	for i, v := range values {
		g.Timers[i].SetRaw(v)
		g.Timers[i].Update()
	}
}

func TestGame2AudioNativeFiveStageMix4BD840(t *testing.T) {
	d := game2AudioTestSetup(t)
	d.set(0, 1)
	e := game2AudioAllocate452300(d.Pointer)
	p := game2AudioAlloc(unsafe.Sizeof(game2AudioVoice{}))
	if p == nil {
		t.Fatal("native voice")
	}
	defer game2AudioFree(p)
	v := (*game2AudioVoice)(p)
	v.Owner = e
	fxp := game2AudioAlloc(unsafe.Sizeof(timer.TimerGroup{}))
	if fxp == nil {
		t.Fatal("FX group")
	}
	defer game2AudioFree(fxp)
	fx := (*timer.TimerGroup)(fxp)
	oldFX := game2AudioFixtureFXOwner(fxp, true)
	defer game2AudioFixtureFXOwner(oldFX, true)
	cases := []struct {
		name         string
		values       [5][3]uint32
		fx, adaptive bool
	}{
		{"all-non-neutral", [5][3]uint32{{0x3000, 125, 0x1000}, {0x2000, 80, 0x4000}, {0x3800, 150, 0}, {0x1000, 75, 0x4000}, {0x2800, 110, 0x1000}}, true, true},
		{"pan-clips-between-stages", [5][3]uint32{{0x4000, 100, 0}, {0x4000, 150, 0}, {0x2000, 80, 0x4000}, {0x3000, 75, 0x4000}, {0x2000, 100, 0}}, true, true},
		{"above-unity-no-volume-clamp", [5][3]uint32{{0x6000, 100, 0x2000}, {0x6000, 100, 0x2000}, {0x4000, 100, 0x2000}, {0x4000, 100, 0x2000}, {0x4000, 100, 0x2000}}, true, true},
		{"no-FX-adaptive-once", [5][3]uint32{{0x4000, 100, 0x2000}, {0x4000, 100, 0x2000}, {0x1000, 200, 0}, {0x4000, 100, 0x2000}, {0x2000, 125, 0x3000}}, false, true},
		{"no-adaptive-pointer", [5][3]uint32{{0x3000, 125, 0x1000}, {0x2800, 80, 0x4000}, {0x2000, 110, 0}, {0x3800, 150, 0x3000}, {0x1000, 200, 0}}, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var groups [5]timer.TimerGroup
			for i := range groups {
				game2AudioTestGroup(&groups[i], tc.values[i])
			}
			v.Group = groups[0]
			e.Group = groups[1]
			*fx = groups[2]
			game2Audio.Context.BaseGroup = groups[3]
			*game2Audio.Master = groups[4]
			v.Master = fx
			fxRef := &groups[2]
			if !tc.fx {
				game2AudioFixtureFXOwner(nil, true)
				v.Master = game2AudioFXMaster()
				fxRef = nil
				if v.Master != nil {
					t.Fatal("nil actual FX must not alias adaptive master")
				}
			} else {
				game2AudioFixtureFXOwner(fxp, true)
			}
			game2Audio.Context.EventMaster = game2Audio.Master
			adaptiveRef := &groups[4]
			if !tc.adaptive {
				game2Audio.Context.EventMaster = nil
				adaptiveRef = nil
			}
			// Identical literal output-buffer seed, independent of the previous
			// candidate composition (Init preserves some unchanged timer flags).
			v.Mixed = timer.TimerGroup{}
			var expected timer.TimerGroup
			game2AudioFixtureMix(&expected, &groups[0], &groups[1], fxRef, &groups[3], adaptiveRef)
			game2AudioUpdateMixSources(v)
			game2AudioMixVoice4BD840(v)
			if v.Mixed != expected || v.Group != groups[0] || e.Group != groups[1] {
				t.Fatal("exact Voice/Owner/FX/Base/Adaptive composition, order, or updated-bit clearing")
			}
			for _, rate := range []uint32{0, 44100, 0x80000001, 0xffffffff} {
				want := game2AudioFixtureMixedParameters(&expected, rate)
				gain, pan, pitch := game2AudioMixedSampleParameters(v, rate)
				if [3]int32{int32(gain), int32(pan), int32(pitch)} != want {
					t.Fatal("actual43F060 composed gain/pan and wrapping signed pitch consumers")
				}
			}
			// The adaptive group is an actual context pointer owner and affects
			// the next refreshed group without replacing the independent FX group.
			if tc.adaptive {
				game2Audio.Master.Timers[0].SetRaw(0x1000)
				game2Audio.Master.Timers[0].Update()
				game2AudioUpdateMixSources(v)
				game2AudioMixVoice4BD840(v)
				if v.Mixed.Timers[0].Current == expected.Timers[0].Current {
					t.Fatal("adaptive gain change was ignored")
				}
			}
		})
	}
	game2Audio.Context.EventMaster = game2Audio.Master
}

func TestGame2AudioNativeAudibleGainStealing487810(t *testing.T) {
	d := game2AudioTestSetup(t)
	d.set(0, 1)
	a, b := game2AudioAllocate452300(d.Pointer), game2AudioAllocate452300(d.Pointer)
	p := game2AudioAlloc(2 * unsafe.Sizeof(game2AudioVoice{}))
	if p == nil {
		t.Fatal("voice owners")
	}
	defer game2AudioFree(p)
	first := (*game2AudioVoice)(p)
	second := (*game2AudioVoice)(unsafe.Add(p, unsafe.Sizeof(game2AudioVoice{})))
	bank := game2Audio.Bank
	oldVoices := bank.Voices
	bank.Voices = []*game2AudioVoice{first, second}
	defer func() { bank.Voices = oldVoices }()
	game2AudioTestGroup(&first.Group, [3]uint32{0x4000, 100, 0x2000})
	second.Group = first.Group
	game2AudioTestGroup(&a.Group, [3]uint32{0x2000, 100, 0x2000})
	game2AudioTestGroup(&b.Group, [3]uint32{0x1000, 100, 0x2000})
	game2AudioTestGroup(&game2Audio.Context.BaseGroup, [3]uint32{0x4000, 100, 0x2000})
	game2AudioTestGroup(game2Audio.Master, [3]uint32{0x4000, 100, 0x2000})
	first.Owner = a
	second.Owner = b
	first.Priority = 7
	second.Priority = 7
	first.Flags = 1
	second.Flags = 1
	game2AudioMixVoice4BD840(first)
	game2AudioMixVoice4BD840(second)
	if first.Group.Timers[0].Current != second.Group.Timers[0].Current || first.Mixed.Timers[0].Current <= second.Mixed.Timers[0].Current {
		t.Fatal("independent raw/effective gain stimulus")
	}
	if got := game2AudioAcquire452810(7, 0); got != second {
		t.Fatal("same-priority stealing must choose quieter composed voice")
	}
	for _, delta := range []uint32{0x665, 0x666} {
		first.Priority = 7
		second.Priority = 7
		first.Flags = 1
		second.Flags = 1
		first.Owner = a
		second.Owner = b
		first.Mixed.Timers[0].Current = 0x50000000
		second.Mixed.Timers[0].Current = 0x50000000 - delta
		want := first
		if delta == 0x666 {
			want = second
		}
		if got := game2AudioAcquire452810(7, 0); got != want {
			t.Fatal("exact cached mixed-gain0x666 cutoff")
		}
	}
}
