package opennox

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// Literal GF(2) basis columns of original GAME.EXE 0048C5E0's 32 ROL/JAE/XOR
// instructions. The oracle does not call rotl, cfshl, randI32, or a supplied
// random-service callback. The original PE and instruction bytes are audited
// separately; no original Windows executable is run by these native tests.
var particleRNG32Basis48C5E0 = [32]uint32{
	0x00000a0b, 0x00001416, 0x0000282c, 0x00005058,
	0x0000a0b0, 0x00014160, 0x000282c0, 0x00050580,
	0x000a0b00, 0x00141600, 0x00282c00, 0x00505800,
	0x00a0b000, 0x01416000, 0x0282c000, 0x05058000,
	0x0a0b0000, 0x14160000, 0x282c0000, 0x50580000,
	0xa0b00000, 0x41600a0b, 0x82c01416, 0x05802227,
	0x0b00444e, 0x1600889c, 0x2c011138, 0x58022270,
	0xb00444e0, 0x600883cb, 0xc0110796, 0x80220527,
}

func particleRNG32Word48C5E0(seed uint32) uint32 {
	var out uint32
	for bit, column := range particleRNG32Basis48C5E0 {
		if seed&(uint32(1)<<bit) != 0 {
			out ^= column
		}
	}
	return out
}

func particleRNG32Range48C5E0(seed uint32, low, high int) (int, uint32) {
	word := particleRNG32Word48C5E0(seed)
	span := uint32(high) - uint32(low)
	upperProduct := uint32((uint64(span) * uint64(word)) >> 32)
	return int(int32(uint32(low) + upperProduct)), word
}

func particleRNG32Recover48C5E0(t *testing.T, call func()) (ok bool) {
	t.Helper()
	defer func() {
		if value := recover(); value != nil {
			t.Errorf("native particle RNG32 panicked: %v", value)
		}
	}()
	call()
	return true
}

func TestParticleRNG32Original48C5E0BoundaryStreams(t *testing.T) {
	seeds := []uint32{0, 1, 0x80000000, 0x7fffffff, 0xffffffff, 0x13527438, 0x55555555, 0xaaaaaaaa, 0xdeadbeef}
	ranges := []struct{ low, high int }{
		{0, 0}, {1, 1}, {-9, -9}, {0, 1}, {-256, 256}, {-8, -2}, {2, 5},
		{math.MinInt32, math.MaxInt32}, {math.MaxInt32, math.MinInt32},
		{math.MaxInt32, math.MaxInt32 - 1}, {19, -37}, {-1, 0},
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wide := uint64(1) << 40
		ranges = append(ranges, struct{ low, high int }{int(wide) - 256, -int(wide) + 256}, struct{ low, high int }{int(wide) + math.MaxInt32, -int(wide) + math.MinInt32})
	}
	for _, seed := range seeds {
		for _, bounds := range ranges {
			t.Run(fmt.Sprintf("seed-%08x/low-%d/high-%d", seed, bounds.low, bounds.high), func(t *testing.T) {
				pfx := partFXes{}
				pfx.rand.u16, pfx.rand.u32 = 0xbeef, seed
				pfx.types[7].field1, pfx.types[7].field4 = 71, 79
				before := pfx.types
				word := seed
				if !particleRNG32Recover48C5E0(t, func() {
					for i := 0; i < 8; i++ {
						want, next := particleRNG32Range48C5E0(word, bounds.low, bounds.high)
						got := pfx.randI32(bounds.low, bounds.high)
						if got != want || pfx.rand.u32 != next {
							t.Fatalf("step %d result=%d state=%08x, original result=%d state=%08x", i, got, pfx.rand.u32, want, next)
						}
						word = next
					}
				}) {
					return
				}
				if pfx.rand.u16 != 0xbeef || !reflect.DeepEqual(pfx.types, before) || pfx.r != nil || pfx.byHandle != nil {
					t.Fatal("RNG32 changed the 16-bit stream, particle table, renderer, or native-handle map")
				}
			})
		}
	}
}

func TestParticleRNG32Original48C5E0AllBasisAndLowWords(t *testing.T) {
	pfx := partFXes{}
	pfx.rand.u16 = 0x1352
	if !particleRNG32Recover48C5E0(t, func() {
		for i := 0; i < 32; i++ {
			pfx.rand.u32 = uint32(1) << i
			got := pfx.randI32(0, 0)
			if got != 0 || pfx.rand.u32 != particleRNG32Basis48C5E0[i] {
				t.Fatalf("bit %d: result=%d state=%08x, original column=%08x", i, got, pfx.rand.u32, particleRNG32Basis48C5E0[i])
			}
		}
		for seed := uint32(0); seed <= 0xffff; seed++ {
			pfx.rand.u32 = seed
			want, next := particleRNG32Range48C5E0(seed, -256, 256)
			if got := pfx.randI32(-256, 256); got != want || pfx.rand.u32 != next {
				t.Fatalf("low word %04x: result=%d state=%08x, want %d/%08x", seed, got, pfx.rand.u32, want, next)
			}
		}
	}) {
		return
	}
	if pfx.rand.u16 != 0x1352 {
		t.Fatal("exhaustive RNG32 changed RNG16")
	}
}

func TestParticleRNG32Original48C5E0InitialLiteralStream(t *testing.T) {
	pfx := partFXes{}
	pfx.Init(nil)
	if pfx.rand.u16 != 0x1352 || pfx.rand.u32 != 0x13527438 {
		t.Fatal("original particle initial seeds changed")
	}
	want := []uint32{0x982811ac, 0x79b2ee28, 0xb5f2dd97, 0x02f9b741, 0x2e135d6d, 0xbef586c0, 0x73a478a2, 0x62e63cb5}
	particleRNG32Recover48C5E0(t, func() {
		for i, word := range want {
			if got := pfx.randI32(math.MinInt32, math.MinInt32); got != math.MinInt32 || pfx.rand.u32 != word {
				t.Fatalf("initial stream step %d: result=%d state=%08x, original state=%08x", i, got, pfx.rand.u32, word)
			}
		}
	})
}

// Real randPoint -> factory -> opaque native handle -> default update/free
// lifecycle. Trigonometric table words and particle properties are explicit
// inputs; no RNG, creation, lifetime, movement, or lookup result is hooked.
// This isolates arithmetic/lifecycle, not stock sprite or full GUI pixels.
func TestParticleRNG32Original48C5E0PointFactoryLifecycle(t *testing.T) {
	var oldX, oldY [257]int32
	for i := 0; i <= 256; i++ {
		x, y := memmap.PtrInt32(0x5D4594, 1310096+uintptr(i)*4), memmap.PtrInt32(0x5D4594, 1309840+uintptr(i)*4)
		oldX[i], oldY[i] = *x, *y
	}
	t.Cleanup(func() {
		for i := 0; i <= 256; i++ {
			*memmap.PtrInt32(0x5D4594, 1310096+uintptr(i)*4) = oldX[i]
			*memmap.PtrInt32(0x5D4594, 1309840+uintptr(i)*4) = oldY[i]
		}
	})
	// The original X/Y views overlap. Restore and initialize by physical
	// address rather than pretending their backing data are independent.
	for i := 0; i < 321; i++ {
		*memmap.PtrInt32(0x5D4594, 1309840+uintptr(i)*4) = int32((i - 160) * 1024)
	}
	handles.Init()
	pfx := partFXes{}
	pfx.Init(nil)
	defer pfx.Free()
	particleRNG32Recover48C5E0(t, func() {
		word := uint32(0x13527438)
		for i := 0; i < 8; i++ {
			index, next := particleRNG32Range48C5E0(word, 0, 256)
			word = next
			wantPos := image.Pt((120*(index+64-160)*1024)>>16, (120*(index-160)*1024)>>16).Add(image.Pt(300, 400))
			position := pfx.randPoint(0, 256, 120, image.Pt(300, 400))
			if position != wantPos || pfx.rand.u32 != word {
				t.Fatalf("real randPoint step %d=%v, original %v", i, position, wantPos)
			}
			duration, next := particleRNG32Range48C5E0(word, 1, 17)
			word = next
			p := pfx.nox_xxx_particleFxNew_4AF990(position, 1, pfx.randI32(1, 17), color.White)
			if p == nil || p.pos != wantPos || p.ticksLeft != duration || pfx.rand.u32 != word {
				t.Fatal("real particle factory/RNG32 state differs")
			}
			handle := p.C()
			if handle == nil || pfx.asParticlefx(handle) != p || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(handle) <= math.MaxUint32) {
				t.Fatal("real particle handle lost native width or ownership")
			}
			p.drawFunc = nil // No renderer assets are part of this fixture.
			p.xrel16, p.yrel16, p.zrel16 = 0x4000, -0x2000, 0
			for tick := 1; tick <= duration; tick++ {
				pfx.Draw()
				if pfx.asParticlefx(handle) != p || p.ticksLeft != duration-tick || p.cnt40 != tick || p.x16 != (wantPos.X<<16)+tick*0x4000 || p.y16 != (wantPos.Y<<16)-tick*0x2000 || p.z16 != 0x10000 {
					t.Fatal("actual particle update position/lifetime differs")
				}
			}
			pfx.Draw()
			if pfx.asParticlefx(handle) != nil || len(pfx.byHandle) != 0 {
				t.Fatal("expired real particle was not released")
			}
		}
	})
}
