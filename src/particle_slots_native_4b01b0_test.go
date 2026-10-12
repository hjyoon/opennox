package opennox

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// GAME.EXE 004B01BF searches the active bitmap for its FIRST zero byte.
// A free record is chosen by original slot index, not last free or a runtime
// cache's arbitrary order. Native opaque handles identify logical records;
// these tests do not assert the physical addresses of original PE realloc.
func TestParticleSlotsOriginal4B01B0LowestInactive(t *testing.T) {
	handles.Init()
	tests := []struct {
		name        string
		freed, want []int
	}{
		{"ascending-two", []int{0, 1}, []int{0, 1}},
		{"descending-two", []int{1, 0}, []int{0, 1}},
		{"three-middle-first", []int{1, 3, 0}, []int{0, 1, 3}},
		{"three-high-first", []int{3, 1, 0}, []int{0, 1, 3}},
		{"three-even-reverse", []int{4, 2, 0}, []int{0, 2, 4}},
		{"three-even-mixed", []int{0, 4, 2}, []int{0, 2, 4}},
		{"three-odd-reverse", []int{5, 3, 1}, []int{1, 3, 5}},
		{"three-odd-mixed", []int{1, 5, 3}, []int{1, 3, 5}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var owner partFXes
			owner.Init(nil)
			defer owner.Free()
			var slots [6]*particleFx
			var ids [6]unsafe.Pointer
			for i := range slots {
				slots[i] = owner.nox_xxx_particleFxNew_4AF990(image.Pt(100+i, 200+i), 1, 7, color.White)
				assertParticleRegistration4B01B0(t, &owner, slots[i])
				ids[i] = slots[i].chnd
			}
			for _, index := range tc.freed {
				slots[index].free()
			}
			if len(owner.byHandle) != 6-len(tc.freed) {
				t.Fatal("real free retained an active slot")
			}
			for i, index := range tc.want {
				p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(300+i, 400+i), 2, 9, color.White)
				assertParticleRegistration4B01B0(t, &owner, p)
				if p != slots[index] || p.chnd != ids[index] {
					t.Fatalf("allocation %d did not reuse original lowest inactive slot %d: got=%p/%p expected=%p/%p", i, index, p, p.chnd, slots[index], ids[index])
				}
				if p.pos != image.Pt(300+i, 400+i) || p.z != 2 || p.ticksLeft != 9 || p.cnt40 != 0 || p.flags != 0 {
					t.Fatal("canonical reuse changed factory reset semantics")
				}
				if len(owner.byHandle) != 6-len(tc.freed)+i+1 {
					t.Fatal("reuse lost or duplicated an active handle")
				}
			}
			for i, p := range slots {
				if owner.asParticlefx(ids[i]) != p {
					t.Fatal("reuse damaged a neighboring live slot")
				}
			}
			if owner.rand.u16 != 0x1352 || owner.rand.u32 != 0x13527438 {
				t.Fatal("slot allocation consumed original RNG")
			}
		})
	}
}

func TestParticleSlotsOriginal4B01B0GrowthAndLowHoleReuse(t *testing.T) {
	handles.Init()
	for _, count := range []int{7, 512, 513, 1025} {
		t.Run(fmt.Sprintf("slots-%d", count), func(t *testing.T) {
			var owner partFXes
			owner.Init(nil)
			defer owner.Free()
			slots := make([]*particleFx, count)
			ids := make([]unsafe.Pointer, count)
			for i := range slots {
				slots[i] = owner.nox_xxx_particleFxNew_4AF990(image.Pt(i, 0), 0, 11, color.White)
				assertParticleRegistration4B01B0(t, &owner, slots[i])
				ids[i] = slots[i].chnd
				if i > 0 && uintptr(ids[i]) <= uintptr(ids[i-1]) {
					t.Fatal("first-created logical slots did not retain increasing native handles")
				}
			}
			for i, p := range slots {
				if owner.asParticlefx(ids[i]) != p {
					t.Fatal("real growth invalidated a pre-existing active record")
				}
			}
			middle := count / 2
			for _, index := range []int{count - 1, 0, middle} {
				slots[index].free()
			}
			for _, index := range []int{0, middle, count - 1} {
				p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(index, 99), 0, 13, color.White)
				assertParticleRegistration4B01B0(t, &owner, p)
				if p != slots[index] || p.chnd != ids[index] {
					t.Fatalf("growth=%d did not refill original low hole %d", count, index)
				}
			}
			if len(owner.byHandle) != count {
				t.Fatal("actual 512-boundary growth/reuse lost active records")
			}
			t.Logf("PARTICLE SLOTS: %d simultaneously active logical records; low holes and native-width lookups preserved across actual factory growth; allocator/result hooks=0", count)
		})
	}
}

func TestParticleSlotsOriginal4B01B0SurviveNativeGC(t *testing.T) {
	handles.Init()
	var owner partFXes
	owner.Init(nil)
	defer owner.Free()
	var slots [32]*particleFx
	var ids [32]unsafe.Pointer
	for i := range slots {
		slots[i] = owner.nox_xxx_particleFxNew_4AF990(image.Pt(i, 0), 0, 1, color.White)
		assertParticleRegistration4B01B0(t, &owner, slots[i])
		ids[i] = slots[i].chnd
	}
	for _, index := range []int{31, 17, 3, 0} {
		slots[index].free()
	}
	// Real garbage collection is an explicit native-runtime stimulus, not a
	// replacement allocator, a GC policy override, or a substituted result.
	runtime.GC()
	for _, index := range []int{0, 3, 17, 31} {
		p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(index, 3), 0, 1, color.White)
		assertParticleRegistration4B01B0(t, &owner, p)
		if p != slots[index] || p.chnd != ids[index] {
			t.Fatalf("runtime GC changed original inactive slot identity/order at %d", index)
		}
	}
	runtime.KeepAlive(&owner)
	if len(owner.byHandle) != 32 {
		t.Fatal("native GC/reuse did not preserve every active record")
	}
}

func TestParticleSlotsOriginal4B01B0IsolatedOwnersAndReset(t *testing.T) {
	handles.Init()
	var owners [2]partFXes
	var slots [2][4]*particleFx
	for i := range owners {
		owners[i].Init(nil)
		defer owners[i].Free()
		for j := range slots[i] {
			slots[i][j] = owners[i].nox_xxx_particleFxNew_4AF990(image.Pt(j, i), 0, 2, color.White)
			assertParticleRegistration4B01B0(t, &owners[i], slots[i][j])
		}
	}
	for i := range owners {
		slots[i][0].free()
		p := owners[i].nox_xxx_particleFxNew_4AF990(image.Pt(9, i), 0, 2, color.White)
		if p != slots[i][0] || p.pfx != &owners[i] || owners[1-i].asParticlefx(p.chnd) != nil {
			t.Fatal("allocator crossed owner boundaries or ignored its lowest inactive slot")
		}
	}
	owners[0].Free()
	p := owners[0].nox_xxx_particleFxNew_4AF990(image.Pt(7, 7), 0, 2, color.White)
	assertParticleRegistration4B01B0(t, &owners[0], p)
	if p != slots[0][0] || len(owners[0].byHandle) != 1 || len(owners[1].byHandle) != 4 {
		t.Fatal("owner reset did not restart at logical slot zero or altered another owner")
	}
}
