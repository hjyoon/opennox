package opennox

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// GAME.EXE 004AF990 calls 004B01B0, which marks the new slot active at
// 004B0206 before returning it. No explicit export/handle lookup is needed
// to make a fresh particle eligible for the original draw/update sweep.
func assertParticleRegistration4B01B0(t *testing.T, owner *partFXes, p *particleFx) {
	t.Helper()
	if p == nil || p.pfx != owner || p.chnd == nil || owner.asParticlefx(p.chnd) != p {
		t.Fatalf("factory did not register its real particle: particle=%p owner=%p active=%d", p, owner, len(owner.byHandle))
	}
	if !handles.IsValid(uintptr(p.chnd)) || uintptr(p.chnd)%unsafe.Alignof(uintptr(0)) != 0 {
		t.Fatalf("invalid/unaligned opaque particle handle %p", p.chnd)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p.chnd) <= math.MaxUint32 {
		t.Fatalf("native particle handle did not exercise high address: %p", p.chnd)
	}
}

func TestParticleRegistrationOriginal4B01B0FreshAllocators(t *testing.T) {
	handles.Init()
	t.Run("raw-allocator", func(t *testing.T) {
		var owner partFXes
		owner.Init(nil)
		defer owner.Free()
		p := owner.newParticlefx()
		assertParticleRegistration4B01B0(t, &owner, p)
		if len(owner.byHandle) != 1 {
			t.Fatal("fresh raw allocator registered more than one slot")
		}
		p.free()
		if len(owner.byHandle) != 0 {
			t.Fatal("raw particle free did not unregister the slot")
		}
	})
	for _, duration := range []int{-1, 0, 1, 17} {
		t.Run(fmt.Sprintf("factory-duration-%d", duration), func(t *testing.T) {
			var owner partFXes
			owner.Init(nil)
			defer owner.Free()
			p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(300, 400), 7, duration, color.White)
			assertParticleRegistration4B01B0(t, &owner, p)
			if len(owner.byHandle) != 1 || p.pos != image.Pt(300, 400) || p.z != 7 || p.ticksLeft != duration || p.ticksTotal != duration || p.x16 != 300<<16 || p.y16 != 400<<16 || p.z16 != 7<<16 || p.drawFunc == nil || p.updateFunc == nil {
				t.Fatal("real factory changed initial fields or callback ownership")
			}
			if owner.rand.u16 != 0x1352 || owner.rand.u32 != 0x13527438 {
				t.Fatal("registration consumed either original RNG stream")
			}
			handle := p.chnd
			p.free()
			if owner.asParticlefx(handle) != nil || len(owner.byHandle) != 0 {
				t.Fatal("real factory particle free retained an active handle")
			}
		})
	}
}

func TestParticleRegistrationOriginal4B01B0SimultaneousOwners(t *testing.T) {
	handles.Init()
	var owners [2]partFXes
	var particles [2][32]*particleFx
	seen := make(map[unsafe.Pointer]bool)
	for i := range owners {
		owners[i].Init(nil)
		defer owners[i].Free()
		for j := range particles[i] {
			p := owners[i].nox_xxx_particleFxNew_4AF990(image.Pt(j, i), 0, 2, color.White)
			assertParticleRegistration4B01B0(t, &owners[i], p)
			if seen[p.chnd] || len(owners[i].byHandle) != j+1 {
				t.Fatal("simultaneous factories collided or lost active slots")
			}
			seen[p.chnd] = true
			particles[i][j] = p
		}
	}
	for i := range particles {
		for j, p := range particles[i] {
			if owners[1-i].asParticlefx(p.chnd) != nil {
				t.Fatal("native handle was incorrectly registered to a different owner")
			}
			handle := p.chnd
			p.free()
			if owners[i].asParticlefx(handle) != nil || len(owners[i].byHandle) != 31-j {
				t.Fatal("one particle free damaged neighboring active slots")
			}
		}
	}
}

func TestParticleRegistrationOriginal4B01B0RealPoolReuse(t *testing.T) {
	handles.Init()
	var owner partFXes
	owner.Init(nil)
	defer owner.Free()
	seen := make(map[*particleFx]unsafe.Pointer)
	reused := 0
	// Exercise actual sync.Pool Put/Get, including race builds which may
	// discard cached objects. No allocator hook or GC/pool policy override.
	for cycle := 0; cycle < 128; cycle++ {
		p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(cycle, -cycle), cycle, cycle%7, color.White)
		assertParticleRegistration4B01B0(t, &owner, p)
		if previous, ok := seen[p]; ok {
			reused++
			if previous != p.chnd {
				t.Fatal("a reused real allocation changed its stable opaque handle")
			}
		}
		seen[p] = p.chnd
		if len(owner.byHandle) != 1 || p.cnt40 != 0 || p.flags != 0 || p.remains44 != 0 || p.partfxTyp != nil {
			t.Fatal("reused factory did not reset original fields or register exactly one slot")
		}
		handle := p.chnd
		p.free()
		if owner.asParticlefx(handle) != nil || len(owner.byHandle) != 0 {
			t.Fatal("pool return retained an active slot")
		}
	}
	if reused == 0 {
		t.Fatal("actual pool did not exercise any cached allocation")
	}
	t.Logf("PARTICLE REGISTRATION: 128 real factory/free cycles, cached returns=%d, allocator/result hooks=0, forced C() registration=0", reused)
}
