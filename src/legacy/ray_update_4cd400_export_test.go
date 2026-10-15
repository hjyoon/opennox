package legacy

import (
	"bytes"
	"math"
	"os"
	"os/exec"
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

type rayUpdateServer219 struct {
	Server
	world *server.Server
}

func (s *rayUpdateServer219) S() *server.Server { return s.world }

func rayUpdateEntryNative219(t *testing.T, kind int, fn unsafe.Pointer) {
	t.Helper()
	if os.Getenv("OPENNOX_RAY_UPDATE_219") != t.Name() {
		cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v")
		cmd.Env = append(os.Environ(), "OPENNOX_RAY_UPDATE_219="+t.Name())
		out, err := cmd.CombinedOutput()
		if err != nil {
			if len(out) > 5000 {
				out = append(out[:2500], out[len(out)-2500:]...)
			}
			t.Fatalf("actual registered CGo ray update: %v\n%s", err, out)
		}
		t.Log(string(out))
		return
	}
	vp, freeVP := alloc.New(noxrender.Viewport{})
	dr, freeDR := alloc.New(client.Drawable{})
	defer freeVP()
	defer freeDR()
	for _, p := range []unsafe.Pointer{vp.C(), dr.C()} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native fixture address below4GiB: %p", p)
		}
	}
	// Give the old entry its genuine initialized type cache and predictable
	// cosmetic RNG, so its RED reaches the PE32 drawable read, not GUI setup.
	for _, off := range []uintptr{1522968, 1522972, 1522976} {
		*memmap.PtrUint32(0x5D4594, off) = 71
	}
	seed := 0
	for prand.New(seed).Int(0, 100) >= 50 {
		seed++
	}
	world := new(server.Server)
	world.Rand.Other = prand.New(seed)
	prior := GetServer
	defer func() { GetServer = prior }()
	GetServer = func() Server { return &rayUpdateServer219{world: world} }
	want := math.MinInt32 + 219 + kind
	calls := 0
	hook := func(v *noxrender.Viewport, d *client.Drawable) int {
		if (calls == 0 && (v != vp || d != dr)) || (calls == 1 && (v != nil || d != nil)) {
			t.Fatalf("callback%d changed pointers: %p/%p", calls, v, d)
		}
		calls++
		return want
	}
	switch kind {
	case 0:
		ClientUpdateCharm4CD400 = hook
	case 1:
		ClientUpdateHeal4CD450 = hook
	case 2:
		ClientUpdateDrainMana4CD690 = hook
	default:
		t.Fatal("unknown ray kind")
	}
	beforeVP := bytes.Clone(unsafe.Slice((*byte)(vp.C()), unsafe.Sizeof(*vp)))
	beforeDR := bytes.Clone(unsafe.Slice((*byte)(dr.C()), unsafe.Sizeof(*dr)))
	if got := ccall.CallIntPtr2(fn, vp.C(), dr.C()); got != want || calls != 1 {
		t.Fatalf("return=%d want=%d calls=%d", got, want, calls)
	}
	if got := ccall.CallIntPtr2(fn, nil, nil); got != want || calls != 2 {
		t.Fatalf("nil delegation return=%d calls=%d", got, calls)
	}
	if !bytes.Equal(beforeVP, unsafe.Slice((*byte)(vp.C()), unsafe.Sizeof(*vp))) || !bytes.Equal(beforeDR, unsafe.Slice((*byte)(dr.C()), unsafe.Sizeof(*dr))) {
		t.Fatal("native entry changed input records")
	}
	t.Logf("registered CGo ray kind%d: native viewport/drawable, nil forwarding, signed return, whole records unchanged", kind)
}

func TestRayUpdate4CD400RegisteredCharmNativePointers(t *testing.T) {
	rayUpdateEntryNative219(t, 0, Get_sub_4CD400())
}
func TestRayUpdate4CD450RegisteredHealNativePointers(t *testing.T) {
	rayUpdateEntryNative219(t, 1, Get_sub_4CD450())
}
func TestRayUpdate4CD690RegisteredDrainManaNativePointers(t *testing.T) {
	rayUpdateEntryNative219(t, 2, Get_sub_4CD690())
}
