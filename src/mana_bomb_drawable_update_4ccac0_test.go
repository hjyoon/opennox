package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

func TestUpdateManaBombCharge4CCAC0HighAddress(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(100, 200), Field_80: 547}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", source)
	}
	const frame = uint32(549)
	sparks, orbs := 0, 0
	hooks := manaBombDrawableHooks4CCAC0{
		radius: func() int { return 32 },
		typeID: func(name string) int {
			switch name {
			case "VioletSpark":
				return 1
			case "ManaBombOrb":
				return 2
			default:
				t.Fatalf("unexpected drawable type %q", name)
				return 0
			}
		},
		random: func(min, max int) int {
			switch {
			case min == 0 && max == 32:
				return 16
			case min == 0 && max == 255:
				return 0
			case min == 10 && max == 30:
				return 20
			case min == 2 && max == 8:
				return 4
			case min == 3 && max == 10:
				return 5
			default:
				t.Fatalf("unexpected random range (%d, %d)", min, max)
				return 0
			}
		},
		frame: func() uint32 { return frame },
		spawn: func(typ int, pos image.Point) *client.Drawable {
			return &client.Drawable{TypeIDVal: uint32(typ), PosVec: pos, AnimStart: frame}
		},
		activate: func(dr *client.Drawable) {
			effect := dr.UnionEffect()
			switch dr.TypeIDVal {
			case 1:
				sparks++
				if dr.PosVec != image.Pt(124, 200) || effect.Field_108 != 124<<12 ||
					effect.Field_109 != 200<<12 || effect.Field_110 != 0 ||
					effect.Field_111 != frame || effect.Field_112 != frame+20 ||
					dr.VelZ != 4 || dr.ZVal != 0 {
					t.Fatalf("spark %d not initialized: %+v, %+v", sparks, dr, effect)
				}
			case 2:
				orbs++
				if dr.ClientUpdateFuncPtr != legacy.Get_sub_4CA720() ||
					effect.Field_108 != 100|200<<16 || effect.Field_111 != 5 {
					t.Fatalf("orb %d not initialized: %+v, %+v", orbs, dr, effect)
				}
				if orbs == 1 {
					if dr.Field_127 != frame%51 || byte(effect.Field_110>>24) != 0 {
						t.Fatalf("first orb direction = (%d, %#x)", dr.Field_127, effect.Field_110)
					}
				} else if orbs == 2 && byte(effect.Field_110>>24) != 1 {
					t.Fatalf("reverse orb direction = %#x", effect.Field_110)
				}
			default:
				t.Fatalf("unexpected drawable type %d", dr.TypeIDVal)
			}
		},
	}
	if got := updateManaBombCharge4CCAC0(source, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if sparks != 20 || orbs != 10 {
		t.Fatalf("spark count = %d, orb count = %d", sparks, orbs)
	}
	source.Field_80 = frame - 10
	updateManaBombCharge4CCAC0(source, hooks)
	if orbs != 10 {
		t.Fatalf("old charge spawned %d more orbs", orbs-10)
	}
}

func TestManaBombOrbStep4CA720HighAddress(t *testing.T) {
	orb := &client.Drawable{PosVec: image.Pt(120, 100), AnimStart: 500}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(orb)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", orb)
	}
	effect := orb.UnionEffect()
	effect.Field_108 = 100 | 100<<16
	effect.Field_110 = 30 | 1<<24
	if next, done := manaBombOrbStep4CA720(orb, 500); next != image.Pt(130, 100) || done {
		t.Fatalf("initial step = %v, done %v", next, done)
	}
	if next, done := manaBombOrbStep4CA720(orb, 530); next != image.Pt(100, 115) || done {
		t.Fatalf("turning step = %v, done %v", next, done)
	}
	if _, done := manaBombOrbStep4CA720(orb, 560); !done {
		t.Fatal("orb survived past 60 frames")
	}
	orb.PosVec = image.Pt(105, 100)
	if _, done := manaBombOrbStep4CA720(orb, 501); !done {
		t.Fatal("orb within 10 pixels was not removed")
	}
}
