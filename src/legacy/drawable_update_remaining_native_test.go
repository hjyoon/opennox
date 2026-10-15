package legacy

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type remainingDrawableHook228 func(*noxrender.Viewport, *client.Drawable) int
type remainingDrawableServer228 struct {
	Server
	world *server.Server
}

func (s *remainingDrawableServer228) S() *server.Server { return s.world }

func TestRemainingDrawableUpdateEntriesNative228(t *testing.T) {
	entries := []struct {
		name string
		fn   unsafe.Pointer
		void bool
		set  func(remainingDrawableHook228) func()
	}{
		{name: "MagicMissile", fn: Get_nox_xxx_updDrawMagicMissile_4CD9E0(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateMagicMissile4CD9E0
			ClientUpdateMagicMissile4CD9E0 = h
			return func() { ClientUpdateMagicMissile4CD9E0 = old }
		}},
		{name: "TeleportWake", fn: Get_nox_xxx_updDrawTeleportWake_4CD8D0(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateTeleportWake4CD8D0
			ClientUpdateTeleportWake4CD8D0 = h
			return func() { ClientUpdateTeleportWake4CD8D0 = old }
		}},
		{name: "SparkleTrail", fn: Get_nox_xxx_updDrawSparkleTrail_4CDBF0(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateSparkleTrail4CDBF0
			ClientUpdateSparkleTrail4CDBF0 = h
			return func() { ClientUpdateSparkleTrail4CDBF0 = old }
		}},
		{name: "VortexSource", fn: Get_nox_xxx_updDrawVortexSource_4CC950(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateVortexSource4CC950
			ClientUpdateVortexSource4CC950 = h
			return func() { ClientUpdateVortexSource4CC950 = old }
		}},
		{name: "ManaBombCharge", fn: Get_nox_xxx_updDrawManabombCharge_4CCAC0(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateManaBombCharge4CCAC0
			ClientUpdateManaBombCharge4CCAC0 = h
			return func() { ClientUpdateManaBombCharge4CCAC0 = old }
		}},
		{name: "ManaBombOrb", fn: Get_sub_4CA720(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateManaBombOrb4CA720
			ClientUpdateManaBombOrb4CA720 = h
			return func() { ClientUpdateManaBombOrb4CA720 = old }
		}},
		{name: "Cloud", fn: Get_nox_xxx_updDrawCloud_4CE1D0(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateCloud4CE1D0
			ClientUpdateCloud4CE1D0 = nil
			if h != nil {
				ClientUpdateCloud4CE1D0 = func(v *noxrender.Viewport, d *client.Drawable, radius int) int {
					if radius != 75 {
						panic("registered cloud radius changed")
					}
					return h(v, d)
				}
			}
			return func() { ClientUpdateCloud4CE1D0 = old }
		}},
		{name: "SmallCloud", fn: Get_sub_4CE360(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateCloud4CE1D0
			ClientUpdateCloud4CE1D0 = nil
			if h != nil {
				ClientUpdateCloud4CE1D0 = func(v *noxrender.Viewport, d *client.Drawable, radius int) int {
					if radius != 35 {
						panic("registered cloud radius changed")
					}
					return h(v, d)
				}
			}
			return func() { ClientUpdateCloud4CE1D0 = old }
		}},
		{name: "CloudRise", fn: Get_sub_4CE340(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateCloudRise4CE340
			ClientUpdateCloudRise4CE340 = h
			return func() { ClientUpdateCloudRise4CE340 = old }
		}},
		{name: "Meteor", fn: Get_sub_4CCD00(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateMeteor4CCD00
			ClientUpdateMeteor4CCD00 = h
			return func() { ClientUpdateMeteor4CCD00 = old }
		}},
		{name: "Fist", fn: Get_nox_xxx_updDrawFist_4CCDB0(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateFist4CCDB0
			ClientUpdateFist4CCDB0 = h
			return func() { ClientUpdateFist4CCDB0 = old }
		}},
		{name: "PredictLinear", fn: Get_nox_xxx_sprite_4CA540(), void: false, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdatePredictLinear4CA540
			ClientUpdatePredictLinear4CA540 = h
			return func() { ClientUpdatePredictLinear4CA540 = old }
		}},
		{name: "Magic", fn: Get_nox_xxx_updDrawMagic_4CDD80(), void: true, set: func(h remainingDrawableHook228) func() {
			old := ClientUpdateMagic4CDD80
			ClientUpdateMagic4CDD80 = nil
			if h != nil {
				ClientUpdateMagic4CDD80 = func(v *noxrender.Viewport, d *client.Drawable) { h(v, d) }
			}
			return func() { ClientUpdateMagic4CDD80 = old }
		}},
	}
	for index, entry := range entries {
		t.Run(entry.name, func(t *testing.T) {
			if os.Getenv("OPENNOX_DRAWABLE_UPDATE_228") != t.Name() {
				cmd := exec.Command(os.Args[0], "-test.run=^"+regexp.QuoteMeta(t.Name())+"$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_DRAWABLE_UPDATE_228="+t.Name())
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("actual registered CGo callback: %v\n%s", err, out)
				}
				t.Log(string(out))
				return
			}
			vp, freeVP := alloc.New(noxrender.Viewport{})
			defer freeVP()
			dr, freeDR := alloc.New(client.Drawable{})
			defer freeDR()
			for _, p := range []unsafe.Pointer{vp.C(), dr.C()} {
				if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
					t.Fatalf("native fixture below 4GiB: %p", p)
				}
			}
			dr.PosVec = image.Pt(123, -456)
			// Only bypass old renderer type-cache setup, not a game condition. The
			// subject is the registered ABI, with opaque records and a recording hook.
			for off := uintptr(1522944); off <= 1523020; off += 4 {
				*memmap.PtrUint32(0x5D4594, off) = 71
			}
			*memmap.PtrInt32(0x587000, 190108) = 0
			world := new(server.Server)
			world.Rand.Other = prand.New(0)
			priorServer := GetServer
			defer func() { GetServer = priorServer }()
			GetServer = func() Server { return &remainingDrawableServer228{world: world} }
			want := math.MinInt32 + 228 + index
			pairs := [][2]unsafe.Pointer{{vp.C(), dr.C()}, {vp.C(), nil}, {nil, dr.C()}, {nil, nil}}
			calls := 0
			restore := entry.set(func(v *noxrender.Viewport, d *client.Drawable) int {
				if calls >= len(pairs) || unsafe.Pointer(v) != pairs[calls][0] || unsafe.Pointer(d) != pairs[calls][1] {
					t.Fatalf("callback%d pointer identity changed: %p/%p", calls, v, d)
				}
				calls++
				return want
			})
			defer restore()
			beforeVP := bytes.Clone(unsafe.Slice((*byte)(vp.C()), unsafe.Sizeof(*vp)))
			beforeDR := bytes.Clone(unsafe.Slice((*byte)(dr.C()), unsafe.Sizeof(*dr)))
			t.Logf("registered native input: name=%s viewport=%p drawable=%p void=%v", entry.name, vp, dr, entry.void)
			for i, pair := range pairs {
				if entry.void {
					ccall.CallVoidPtr2(entry.fn, pair[0], pair[1])
				} else if got := ccall.CallIntPtr2(entry.fn, pair[0], pair[1]); got != want {
					t.Fatalf("call%d return=%d want=%d callbacks=%d", i, got, want, calls)
				}
				if calls != i+1 {
					t.Fatalf("call%d delegated %d times", i, calls)
				}
			}
			if !bytes.Equal(beforeVP, unsafe.Slice((*byte)(vp.C()), unsafe.Sizeof(*vp))) || !bytes.Equal(beforeDR, unsafe.Slice((*byte)(dr.C()), unsafe.Sizeof(*dr))) {
				t.Fatal("ABI entry modified opaque input records")
			}
			_ = entry.set(nil)
			if entry.void {
				ccall.CallVoidPtr2(entry.fn, nil, nil)
			} else if got := ccall.CallIntPtr2(entry.fn, nil, nil); got != 1 {
				t.Fatalf("renderer-free return=%d want=1", got)
			}
			t.Log(fmt.Sprintf("native %s: four pointer/nil pairs, signed return or original void ABI, complete input bytes unchanged, renderer-free entry", entry.name))
		})
	}
}
