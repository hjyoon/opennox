package opennox

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// Original PE32 calls KERNEL32.dll!MulDiv, not the Win64 implementation.
// Contract: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-muldiv
// The Win32 signed half-denominator bug is documented by Microsoft:
// https://devblogs.microsoft.com/oldnewthing/20120514-00/?p=7633
// This independent exact-integer reference uses quotient/remainder, rather
// than the production rounding-addition implementation or a supplied result.
func particleMulDivReference4B0020(a, b, c int32) int32 {
	if c == 0 {
		return -1
	}
	x, y, d := big.NewInt(int64(a)), big.NewInt(int64(b)), big.NewInt(int64(c))
	x.Abs(x)
	y.Abs(y)
	d.Abs(d)
	product := new(big.Int).Mul(x, y)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(product, d, remainder)
	if c == -0x80000000 {
		// Original SAR of absolute denominator 0x80000000 produces the
		// unsigned DWORD 0xc0000000, instead of mathematical half 0x40000000.
		// Its extra whole divisor is explicit here, not copied from muldiv.
		quotient.Add(quotient, big.NewInt(1))
	}
	if new(big.Int).Lsh(remainder, 1).Cmp(d) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if quotient.Cmp(big.NewInt(0x7fffffff)) > 0 {
		return -1
	}
	if a^b^c < 0 {
		quotient.Neg(quotient)
	}
	return int32(quotient.Int64())
}

func particleMulDivActual4B0020(t *testing.T, a, b, c int, want int32) {
	t.Helper()
	defer func() {
		if value := recover(); value != nil {
			t.Fatalf("actual native MulDiv(%d,%d,%d) panicked: %v", a, b, c, value)
		}
	}()
	if got := muldiv(a, b, c); got != int(want) {
		t.Fatalf("actual MulDiv(%d,%d,%d)=%d, original signed DWORD=%d", a, b, c, got, want)
	}
}

func TestParticleMulDivOriginal4B0020LiteralBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c int64
		want    int32
	}{
		{"zero", 0, 17, 3, 0},
		{"zero-multiplier", -17, 0, -3, 0},
		{"positive-half", 1, 1, 2, 1},
		{"negative-half", -1, 1, 2, -1},
		{"negative-multiplier", 1, -1, 2, -1},
		{"negative-divisor", 1, 1, -2, -1},
		{"double-negative", -1, -1, 2, 1},
		{"negative-numerator-divisor", 1, -1, -2, 1},
		{"all-negative", -1, -1, -2, -1},
		{"below-half", 1, 1, 3, 0},
		{"negative-below-half", -1, 1, 3, 0},
		{"above-half", 2, 1, 3, 1},
		{"negative-above-half", -2, 1, 3, -1},
		{"positive-two-half", 5, 1, 2, 3},
		{"negative-two-half", -5, 1, 2, -3},
		{"positive-limit", 0x7fffffff, 1, 1, 0x7fffffff},
		{"negative-limit", -0x7fffffff, 1, 1, -0x7fffffff},
		{"unsigned-quotient-sign-bit", -0x80000000, 1, 1, -1},
		{"product-overflow-positive", 0x7fffffff, 2, 1, -1},
		{"product-overflow-negative", -0x7fffffff, 2, 1, -1},
		{"wide-product-in-range", 0x7fffffff, 0x7fffffff, 0x7fffffff, 0x7fffffff},
		{"absolute-minimum-operand", -0x80000000, 1, -2, 0x40000000},
		{"zero-divisor", 7, 9, 0, -1},
		{"zero-divisor-zero-product", 0, 0, 0, -1},
		{"original-bounce-positive", 65537, 45875, 65536, 45876},
		{"original-bounce-negative", -32769, 45875, 65536, -22938},
		{"original-damping-half", 3, 32768, 65536, 2},
		{"original-damping-negative-half", -3, 32768, 65536, -2},
		{"Win32-min-divisor-quirk", 1, -0x80000000, -0x80000000, 2},
		{"Win32-min-divisor-zero-positive", 0, -1, -0x80000000, 1},
		{"Win32-min-divisor-zero-negative", 0, 1, -0x80000000, -1},
		{"Win32-min-divisor-half", 0x40000000, 1, -0x80000000, -2},
		{"Win32-min-divisor-product", 0x40000000, 2, -0x80000000, -2},
		{"low-DWORD-wide-number", 0x100000003, 32768, 65536, 2},
		{"low-DWORD-wide-multiplier", 3, 0x100008000, 65536, 2},
		{"low-DWORD-wide-denominator", 3, 32768, 0x100010000, 2},
		{"low-DWORD-zero-denominator", 3, 32768, 0x100000000, -1},
		{"low-DWORD-signed-number", 0x1ffffffff, 1, 2, -1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if reference := particleMulDivReference4B0020(int32(tc.a), int32(tc.b), int32(tc.c)); reference != tc.want {
				t.Fatalf("literal reference is inconsistent: %d != %d", reference, tc.want)
			}
			particleMulDivActual4B0020(t, int(tc.a), int(tc.b), int(tc.c), tc.want)
		})
	}
}

func TestParticleMulDivOriginal4B0020ExactIntegerMatrix(t *testing.T) {
	values := []int32{-0x80000000, -0x7fffffff, -65537, -65536, -32769, -3, -1, 0, 1, 3, 32769, 65536, 65537, 0x7fffffff}
	calls := 0
	for _, a := range values {
		for _, b := range values {
			for _, c := range values {
				particleMulDivActual4B0020(t, int(a), int(b), int(c), particleMulDivReference4B0020(a, b, c))
				calls++
			}
		}
	}
	if calls != 2744 {
		t.Fatal("missing immutable exact-integer matrix inputs")
	}
	t.Log("PARTICLE MULDIV: 2744 exact-integer matrix calls; Win32 min-divisor SAR behavior, signed WORD/DWORD aliases and overflow/zero-divisor results retained")
}

type particleMulDivState4B0020 struct {
	x, y, z, vx, vy, vz, ax, ay, az, gravity, bounce int32
	left, delay, ticks                               int32
	bouncing, dead                                   bool
}

func particleMulDivInput4B0020(mask int, bounce bool) particleMulDivState4B0020 {
	r := particleMulDivState4B0020{x: 135 << 16, y: 131 << 16, z: 1 << 16, vx: 65537, vy: -32769, vz: -73728, gravity: -32768, bounce: 45875, left: 4, delay: 2, bouncing: bounce}
	if mask&1 != 0 {
		r.ax = 32768
	}
	if mask&2 != 0 {
		r.ay = 98304
	}
	if mask&4 != 0 {
		r.az = 49152
	}
	return r
}

func (r *particleMulDivState4B0020) step() {
	if r.dead {
		return
	}
	if r.left < 0 {
		r.dead = true
		return
	}
	r.ticks++
	if r.delay != 0 {
		r.delay--
		return
	}
	x, y, z := r.x+r.vx, r.y+r.vy, r.z+r.vz
	vx, vy, vz := r.vx, r.vy, r.vz
	if z < 0 {
		if r.bouncing {
			z = r.z
			vx = particleMulDivReference4B0020(vx, r.bounce, 65536)
			vy = particleMulDivReference4B0020(vy, r.bounce, 65536)
			vz = particleMulDivReference4B0020(-vz, r.bounce, 65536)
		} else {
			z, vz = 0, 0
		}
	}
	if r.ax != 0 {
		vx = particleMulDivReference4B0020(vx, r.ax, 65536)
	}
	if r.ay != 0 {
		vy = particleMulDivReference4B0020(vy, r.ay, 65536)
	}
	if r.az != 0 {
		vz = particleMulDivReference4B0020(vz, r.az, 65536)
	}
	r.left--
	if r.left < 0 {
		r.dead = true
		return
	}
	r.x, r.y, r.z, r.vx, r.vy, r.vz = x, y, z, vx, vy, vz+r.gravity
}

func setParticleMulDivInput4B0020(p *particleFx, r particleMulDivState4B0020) {
	p.xrel16, p.yrel16, p.zrel16 = int(r.vx), int(r.vy), int(r.vz)
	p.xrel16md, p.yrel16md, p.zrel16md = int(r.ax), int(r.ay), int(r.az)
	p.zrel16off, p.remains44 = int(r.gravity), int(r.delay)
	p.setMulDiv(int(r.bounce), r.bouncing)
}

func assertParticleMulDivState4B0020(t *testing.T, owner *partFXes, p *particleFx, r particleMulDivState4B0020, tick int) {
	t.Helper()
	if p.x16 != int(r.x) || p.y16 != int(r.y) || p.z16 != int(r.z) || p.xrel16 != int(r.vx) || p.yrel16 != int(r.vy) || p.zrel16 != int(r.vz) || p.ticksLeft != int(r.left) || p.remains44 != int(r.delay) || p.cnt40 != int(r.ticks) || (owner.asParticlefx(p.chnd) == p) == r.dead {
		t.Fatalf("real default update/bounce/damping/free differs at tick %d: p=(%d,%d,%d,%d,%d,%d,%d,%d,%d) original=%+v", tick, p.x16, p.y16, p.z16, p.xrel16, p.yrel16, p.zrel16, p.ticksLeft, p.remains44, p.cnt40, r)
	}
}

func TestParticleMulDivOriginal4B0020RealDefaultUpdates(t *testing.T) {
	handles.Init()
	for mask := 0; mask < 8; mask++ {
		for _, bounce := range []bool{false, true} {
			t.Run(fmt.Sprintf("mask-%d/bounce-%t", mask, bounce), func(t *testing.T) {
				defer func() {
					if value := recover(); value != nil {
						t.Fatalf("real particle update panicked: %v", value)
					}
				}()
				var owner partFXes
				owner.Init(nil)
				defer owner.Free()
				r := particleMulDivInput4B0020(mask, bounce)
				p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(135, 131), 1, 4, color.White)
				assertParticleRegistration4B01B0(t, &owner, p)
				setParticleMulDivInput4B0020(p, r)
				p.flags |= 8 // actual hidden-input path; no draw/update result hook
				dr, free := alloc.New(client.Drawable{})
				defer free()
				dr.PosVec = image.Pt(135, 131)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= 0xffffffff {
					t.Fatal("source drawable not native C-owned high-address memory")
				}
				p.setDrawable8(dr)
				// Point8 equals current position so attachment does not add motion.
				dr.Field_8, dr.Field_9 = 135, 131
				before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr))))
				for tick := 0; tick < 8; tick++ {
					r.step()
					owner.Draw()
					assertParticleMulDivState4B0020(t, &owner, p, r, tick)
					if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr)))) {
						t.Fatal("default update changed source drawable bytes")
					}
				}
				if !r.dead || len(owner.byHandle) != 0 {
					t.Fatal("real particle did not expire at the original boundary")
				}
			})
		}
	}
}
