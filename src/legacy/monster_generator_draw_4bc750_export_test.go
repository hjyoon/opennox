package legacy

import (
	"bytes"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

func preserveMonsterGeneratorDraw4BC750(t *testing.T) {
	t.Helper()
	old := monsterGeneratorDrawCall4BC750
	t.Cleanup(func() { monsterGeneratorDrawCall4BC750 = old })
}

func TestMonsterGeneratorDraw4BC750NativeLayout(t *testing.T) {
	var data monsterGeneratorDrawDataNative4BC750
	got := [6]uintptr{unsafe.Sizeof(data), unsafe.Offsetof(data.size), unsafe.Offsetof(data.images),
		unsafe.Offsetof(data.count), unsafe.Offsetof(data.delay), unsafe.Offsetof(data.kind)}
	if want := monsterGeneratorDrawCLayout4BC750(); got != want {
		t.Fatalf("Go layout=%v C layout=%v", got, want)
	}
	want := [6]uintptr{56, 0, 4, 24, 29, 36}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = [6]uintptr{80, 0, 8, 48, 53, 60}
	}
	if got != want {
		t.Fatalf("native conditional animation layout=%v want=%v", got, want)
	}
}

func TestMonsterGeneratorDraw4BC750CEntryFullPointersAndReturn(t *testing.T) {
	preserveMonsterGeneratorDraw4BC750(t)
	vp, freeVP := alloc.New(noxrender.Viewport{})
	dr, freeDR := alloc.New(client.Drawable{})
	t.Cleanup(freeVP)
	t.Cleanup(freeDR)
	for _, ptr := range []unsafe.Pointer{vp.C(), dr.C()} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("pointer=%p, want above 4 GiB", ptr)
		}
	}
	var gotVP *noxrender.Viewport
	var gotDR *client.Drawable
	monsterGeneratorDrawCall4BC750 = func(v *noxrender.Viewport, d *client.Drawable) int32 {
		gotVP, gotDR = v, d
		return math.MinInt32 + 1
	}
	for _, call := range []func(*noxrender.Viewport, *client.Drawable) int32{
		monsterGeneratorDrawCEntry4BC750,
		func(v *noxrender.Viewport, d *client.Drawable) int32 {
			return int32(ccall.CallIntPtr2(Get_nox_thing_monster_gen_draw(), v.C(), d.C()))
		},
		DrawMonsterGenerator4BC750,
	} {
		if got := call(vp, dr); got != math.MinInt32+1 || gotVP != vp || gotDR != dr {
			t.Fatalf("return=%d viewport=%p/%p drawable=%p/%p", got, gotVP, vp, gotDR, dr)
		}
		if got := call(nil, nil); got != math.MinInt32+1 || gotVP != nil || gotDR != nil {
			t.Fatalf("nil delegation: return=%d viewport=%p drawable=%p", got, gotVP, gotDR)
		}
	}
}

type nativeGeneratorFixture4BC750 struct {
	vp     *noxrender.Viewport
	dr     *client.Drawable
	data   *monsterGeneratorDrawDataNative4BC750
	images [5][]noxrender.ImageHandle
}

func newNativeGeneratorFixture4BC750(t *testing.T) nativeGeneratorFixture4BC750 {
	t.Helper()
	vp, freeVP := alloc.New(noxrender.Viewport{})
	dr, freeDR := alloc.New(client.Drawable{})
	data, freeData := alloc.New(monsterGeneratorDrawDataNative4BC750{})
	for _, free := range []func(){freeVP, freeDR, freeData} {
		t.Cleanup(free)
	}
	f := nativeGeneratorFixture4BC750{vp: vp, dr: dr, data: data}
	data.size = uint32(unsafe.Sizeof(*data))
	for state := range data.images {
		// The extra entries are deliberate: verify original inclusive random
		// and unbounded slave indexing without accessing unallocated memory.
		table, free := alloc.Make([]noxrender.ImageHandle(nil), 6)
		t.Cleanup(free)
		f.images[state], data.images[state] = table, &table[0]
		data.count[state], data.delay[state], data.kind[state] = 4, 1, 2
		for i := range table {
			p, free := alloc.Malloc(8)
			t.Cleanup(free)
			table[i] = noxrender.ImageHandle(p)
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
				t.Fatalf("image=%p, want above 4 GiB", p)
			}
		}
	}
	data.count[4], data.delay[4] = 3, 0
	dr.DrawData, dr.NetCode32 = unsafe.Pointer(data), 1
	dr.ObjClass, dr.ObjFlags = object.Class(0x123fffff), object.Flags(0xabcdef00)
	for _, p := range []unsafe.Pointer{vp.C(), dr.C(), unsafe.Pointer(data), unsafe.Pointer(data.images[0])} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native field=%p, want above 4 GiB", p)
		}
	}
	return f
}

func TestMonsterGeneratorDraw4BC750NativeCEntryCachedDataAndLiveOverlay(t *testing.T) {
	preserveMonsterGeneratorDraw4BC750(t)
	f := newNativeGeneratorFixture4BC750(t)
	other, freeOther := alloc.New(monsterGeneratorDrawDataNative4BC750{})
	t.Cleanup(freeOther)
	other.kind[0] = 1
	changed, freeChanged := alloc.Make([]noxrender.ImageHandle(nil), 6)
	t.Cleanup(freeChanged)
	copy(changed, f.images[1])
	buf := unsafe.Slice((*byte)(f.dr.C()), int(unsafe.Sizeof(*f.dr)))
	before := bytes.Clone(buf)
	h := monsterGeneratorDrawNativeHooks4BC750(f.vp, f.dr)
	ticks := 0
	h.frame = func() uint32 {
		ticks++
		if ticks == 1 {
			return 5
		}
		return 0
	}
	var draws []noxrender.ImageHandle
	h.draw = func(img noxrender.ImageHandle) {
		draws = append(draws, img)
		if len(draws) == 1 {
			f.dr.DrawData, f.dr.NetCode32 = unsafe.Pointer(other), 9
			f.data.images[4], f.data.count[4], f.data.delay[4] = &changed[0], 3, 1
		} else {
			f.dr.Flags70Val, f.dr.ObjFlags = 0x800, object.Flags(0x12345600)
		}
	}
	monsterGeneratorDrawCall4BC750 = func(vp *noxrender.Viewport, dr *client.Drawable) int32 {
		if vp != f.vp || dr != f.dr {
			t.Fatalf("native callback viewport=%p drawable=%p", vp, dr)
		}
		return monsterGeneratorDraw4BC750(h)
	}
	if got := monsterGeneratorDrawCEntry4BC750(f.vp, f.dr); got != 1 {
		t.Fatalf("return=%d", got)
	}
	if want := []noxrender.ImageHandle{f.images[0][3], changed[1]}; !reflect.DeepEqual(draws, want) || ticks != 2 {
		t.Fatalf("draws=%v want=%v frame reads=%d", draws, want, ticks)
	}
	if f.dr.Flags70Val != 0x800 || uint32(f.dr.ObjFlags) != 0x12345601 || f.dr.DrawData != unsafe.Pointer(other) ||
		f.dr.NetCode32 != 9 || uint32(f.dr.ObjClass) != 0x123fffff || f.dr.UnionEffect().Field_108 != 0 {
		t.Fatalf("unexpected drawable writes: %+v", f.dr)
	}
	f.dr.Flags70Val, f.dr.ObjFlags, f.dr.DrawData, f.dr.NetCode32 = 0, object.Flags(0xabcdef00), unsafe.Pointer(f.data), 1
	if !bytes.Equal(buf, before) {
		t.Fatal("unrelated drawable fields were modified")
	}
}

func TestMonsterGeneratorDraw4BC750NativeCEntryKindsTimersAndNilImage(t *testing.T) {
	for _, tc := range []struct {
		name     string
		flags    uint32
		kind     uint32
		timer    uint32
		slave    uint32
		frame    int
		nilImage bool
	}{
		{"animated", 0, 2, 0, 0, 3, false},
		{"random-inclusive", 0, 4, 0, 0, 4, false},
		{"slave-not-bounded", 0, 5, 0, 5, 5, false},
		{"kind-zero-timer-override", 0x400, 0, 2, 0, 3, false},
		{"kind-zero-terminal-override", 0x800, 0, 0, 0, 3, false},
		{"timer-completes", 0x400, 5, 1, 5, 3, false},
		{"nil-image-passed-through", 0, 2, 0, 0, 3, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			preserveMonsterGeneratorDraw4BC750(t)
			f := newNativeGeneratorFixture4BC750(t)
			state := 0
			if tc.flags&0xc00 != 0 {
				state = 3
			}
			f.dr.Flags70Val, f.dr.AnimFrameSlave, f.dr.UnionEffect().Field_108 = tc.flags, tc.slave, tc.timer
			f.data.kind[state] = tc.kind
			if tc.nilImage {
				f.images[state][tc.frame] = nil
			}
			h := monsterGeneratorDrawNativeHooks4BC750(f.vp, f.dr)
			h.frame = func() uint32 { return 5 }
			randoms, draws := 0, 0
			h.random = func(min, max int32, source string, line int32) int32 {
				randoms++
				if min != 0 || max != 4 || source != monsterGeneratorDrawSource4BC750 || line != 86 {
					t.Fatalf("random bounds/source=%d,%d/%q:%d", min, max, source, line)
				}
				return max
			}
			h.draw = func(img noxrender.ImageHandle) {
				draws++
				if img != f.images[state][tc.frame] {
					t.Fatalf("image=%p want=%p", img, f.images[state][tc.frame])
				}
				f.dr.Flags70Val |= 0x400 // prevent a second image in this fixture
			}
			monsterGeneratorDrawCall4BC750 = func(vp *noxrender.Viewport, dr *client.Drawable) int32 {
				return monsterGeneratorDraw4BC750(h)
			}
			if got := monsterGeneratorDrawCEntry4BC750(f.vp, f.dr); got != 1 || draws != 1 {
				t.Fatalf("return=%d draws=%d", got, draws)
			}
			if (tc.kind == 4 && randoms != 1) || (tc.kind != 4 && randoms != 0) {
				t.Fatalf("random calls=%d", randoms)
			}
			if tc.timer != 0 && f.dr.UnionEffect().Field_108 != tc.timer-1 {
				t.Fatalf("timer=%d want=%d", f.dr.UnionEffect().Field_108, tc.timer-1)
			}
			if tc.flags&0x800 != 0 && (uint32(f.dr.ObjClass)&0x80000 != 0 || uint32(f.dr.ObjFlags)&0x20000000 != 0) {
				t.Fatalf("terminal class=%x flags=%x", f.dr.ObjClass, f.dr.ObjFlags)
			}
			if tc.flags&0x800 != 0 || tc.timer == 1 {
				if uint32(f.dr.ObjFlags)&1 == 0 || f.dr.Flags70Val&0x800 == 0 {
					t.Fatalf("completion state=%x flags=%x", f.dr.Flags70Val, f.dr.ObjFlags)
				}
			}
		})
	}
}
