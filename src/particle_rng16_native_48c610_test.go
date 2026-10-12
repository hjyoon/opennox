package opennox

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Independent literal columns of original 0048C61E's 16 ROL-WORD/carry/XOR
// operations. No rotl16/cfshl/randU16 result or callback supplies the oracle.
var particleRNG16Basis48C610 = [16]uint16{
	0x000b, 0x0016, 0x002c, 0x0058, 0x00b0, 0x0160, 0x02c0, 0x0580,
	0x0b00, 0x1600, 0x2c00, 0x5800, 0xb000, 0x600b, 0xc016, 0x8027,
}

func particleRNG16Word48C610(seed uint16) uint16 {
	var word uint16
	for bit, column := range particleRNG16Basis48C610 {
		if seed&(uint16(1)<<bit) != 0 {
			word ^= column
		}
	}
	return word
}

func particleRNG16Recover48C610(t *testing.T, call func()) (ok bool) {
	t.Helper()
	defer func() {
		if value := recover(); value != nil {
			t.Errorf("native particle RNG16 panicked: %v", value)
		}
	}()
	call()
	return true
}

func TestParticleRNG16Original48C610BoundaryStreams(t *testing.T) {
	for _, seed := range []uint16{0, 1, 0x7fff, 0x8000, 0xffff, 0x1352, 0x5555, 0xaaaa, 0xdead, 0x000a} {
		t.Run(fmt.Sprintf("seed-%04x", seed), func(t *testing.T) {
			pfx := partFXes{}
			pfx.rand.u16, pfx.rand.u32 = seed, 0x89abcdef
			pfx.types[5].field1, pfx.types[5].field4 = 57, 59
			before := pfx.types
			particleRNG16Recover48C610(t, func() {
				word := seed
				for i := 0; i < 64; i++ {
					word = particleRNG16Word48C610(word)
					if got := pfx.randU16(); got != word || pfx.rand.u16 != word {
						t.Fatalf("step %d: returned=%04x state=%04x, original WORD=%04x", i, got, pfx.rand.u16, word)
					}
				}
			})
			if pfx.rand.u32 != 0x89abcdef || !reflect.DeepEqual(before, pfx.types) || pfx.r != nil || pfx.byHandle != nil {
				t.Fatal("RNG16 changed RNG32, particle table, renderer, or handle map")
			}
		})
	}
}

func TestParticleRNG16Original48C610ExhaustiveWords(t *testing.T) {
	pfx := partFXes{}
	pfx.rand.u32 = 0x13527438
	particleRNG16Recover48C610(t, func() {
		for seed := uint32(0); seed <= 0xffff; seed++ {
			pfx.rand.u16 = uint16(seed)
			want := particleRNG16Word48C610(uint16(seed))
			if got := pfx.randU16(); got != want || pfx.rand.u16 != want {
				t.Fatalf("input %04x: returned=%04x state=%04x, original WORD=%04x", seed, got, pfx.rand.u16, want)
			}
		}
	})
	if pfx.rand.u32 != 0x13527438 {
		t.Fatal("exhaustive WORD sequence changed the DWORD stream")
	}
}

func TestParticleRNG16Original48C610InitialLiteralStream(t *testing.T) {
	pfx := partFXes{}
	pfx.Init(nil)
	if pfx.rand.u16 != 0x1352 || pfx.rand.u32 != 0x13527438 {
		t.Fatal("original particle initial seeds changed")
	}
	want := []uint16{0xaf66, 0x8ab6, 0xca4d, 0x0c8e, 0x71e2, 0x1d2b, 0xce25, 0x2376, 0x7e21, 0x7376, 0x0e37, 0x63e1, 0xbb36, 0x14c6, 0x9b7a, 0x7679}
	particleRNG16Recover48C610(t, func() {
		for i, word := range want {
			if got := pfx.randU16(); got != word || pfx.rand.u16 != word {
				t.Fatalf("initial stream %d: returned=%04x state=%04x, original WORD=%04x", i, got, pfx.rand.u16, word)
			}
		}
	})
	if pfx.rand.u32 != 0x13527438 {
		t.Fatal("WORD initialization stream changed DWORD state")
	}
}

func TestParticleRNG16Original48C610InterleavedStreams(t *testing.T) {
	for _, order := range []string{"16-first", "32-first", "alternating"} {
		t.Run(order, func(t *testing.T) {
			pfx := partFXes{}
			pfx.Init(nil)
			word, dword := uint16(0x1352), uint32(0x13527438)
			particleRNG16Recover48C610(t, func() {
				for i := 0; i < 128; i++ {
					call16 := func() {
						word = particleRNG16Word48C610(word)
						if got := pfx.randU16(); got != word || pfx.rand.u16 != word || pfx.rand.u32 != dword {
							t.Fatal("WORD call changed the original/interleaved streams")
						}
					}
					call32 := func() {
						want, next := particleRNG32Range48C5E0(dword, -256, 256)
						dword = next
						if got := pfx.randI32(-256, 256); got != want || pfx.rand.u32 != dword || pfx.rand.u16 != word {
							t.Fatal("DWORD call changed the original/interleaved streams")
						}
					}
					if order == "32-first" || (order == "alternating" && i%2 != 0) {
						call32()
						call16()
					} else {
						call16()
						call32()
					}
				}
			})
		})
	}
}

// Invoke the real type-1 emitter with a C-owned high-address drawable. Observe
// both actual RNG streams and complete drawable bytes without replacing its
// allocator, creation, or random callbacks. Active-map registration/rendering
// is an independently missing service and is not fabricated by this fixture.
func TestParticleRNG16Original48C610RealType1Consumption(t *testing.T) {
	dr, free := alloc.New(client.Drawable{})
	defer free()
	dr.PosVec = image.Pt(300, 400)
	dr.Field_8, dr.Field_9 = 299, 399
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
		t.Fatal("real type-1 drawable not above 4 GiB")
	}
	before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr))))
	for _, seed := range []uint16{0x1352, 0x8000} {
		for _, count := range []int{-1, 0, 1, 3, 7} {
			t.Run(fmt.Sprintf("seed-%04x/count-%d", seed, count), func(t *testing.T) {
				pfx := partFXes{}
				pfx.Init(nil)
				defer pfx.Free()
				pfx.rand.u16 = seed
				word, dword := seed, uint32(0x13527438)
				for i := 0; i < count; i++ {
					word = particleRNG16Word48C610(word)
					dword = particleRNG32Word48C5E0(particleRNG32Word48C5E0(dword))
				}
				if !particleRNG16Recover48C610(t, func() { pfx.newPartfxT1(&packetParticleFx{dr: dr, cnt: count, max: 128}) }) {
					return
				}
				if pfx.rand.u16 != word || pfx.rand.u32 != dword {
					t.Fatalf("actual type-1 RNG consumption=%04x/%08x, original %04x/%08x", pfx.rand.u16, pfx.rand.u32, word, dword)
				}
				if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr)))) {
					t.Fatal("type-1 emitter altered source drawable bytes")
				}
				t.Logf("actual type-1 emission completed: count=%d RNG16=%04x RNG32=%08x active_handles=%d (registration/render lifecycle not certified)", count, word, dword, len(pfx.byHandle))
			})
		}
	}
}
