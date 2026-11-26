package legacy

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/internal/porttest"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type radialDamageOriginalServer4E25B0 struct {
	Server
	scene *porttest.RadialScene4E25B0
}

func (s *radialDamageOriginalServer4E25B0) MapDamageUnitsAroundNative4E25B0(p *types.Pointf, a, b float32, d int32, typ object.DamageType, source, excluded *server.Object) {
	if p != s.scene.Center(s.scene.Input.Center) {
		panic("C center pointer identity")
	}
	server.MapRadialDamageRaw4E25B0(p, a, b, d, typ, source, excluded, s.scene.Runtime)
}

func TestMapRadialDamageOriginalNativeC4E25B0(t *testing.T) {
	porttest.RunRadialOriginal4E25B0(t, "../server/testdata/map_radial_damage_original_4e25b0.json", true, func(s *porttest.RadialScene4E25B0) {
		old := GetServer
		defer func() { GetServer = old }()
		GetServer = func() Server { return &radialDamageOriginalServer4E25B0{scene: s} }
		i := s.Input
		MapRadialDamageCProbe4E25B0(s.Center(i.Center), math.Float32frombits(i.Outer), math.Float32frombits(i.Inner), i.Damage, object.DamageType(i.Typ), s.Object(i.Source), s.Object(i.Excluded))
	})
}

type radialDamageArgumentCapture4E25B0 struct {
	Server
	center           *types.Pointf
	source, excluded *server.Object
	outer, inner     uint32
	damage           int32
	typ              object.DamageType
}

func (s *radialDamageArgumentCapture4E25B0) MapDamageUnitsAroundNative4E25B0(center *types.Pointf, outer, inner float32, damage int32, typ object.DamageType, source, excluded *server.Object) {
	s.center, s.source, s.excluded = center, source, excluded
	s.outer, s.inner = math.Float32bits(outer), math.Float32bits(inner)
	s.damage, s.typ = damage, typ
}

func TestMapRadialDamageCEntryRetainsFloatWords4E25B0(t *testing.T) {
	center, freeCenter := alloc.NewHigh(types.Pointf{})
	defer freeCenter()
	source, freeSource := alloc.NewHigh(server.Object{})
	defer freeSource()
	excluded, freeExcluded := alloc.NewHigh(server.Object{})
	defer freeExcluded()
	var capture radialDamageArgumentCapture4E25B0
	old := GetServer
	t.Cleanup(func() { GetServer = old })
	GetServer = func() Server { return &capture }
	for _, radii := range [][2]uint32{
		{0x7f812345, 0xff812346},
		{0xffc23456, 0x7fc34567},
		{0x80000000, 0x00000001},
	} {
		const damage, typ = -19, object.DamageType(-7)
		MapRadialDamageCProbe4E25B0(center, math.Float32frombits(radii[0]), math.Float32frombits(radii[1]), damage, typ, source, excluded)
		if capture.outer != radii[0] || capture.inner != radii[1] {
			t.Fatalf("radii = %08x %08x, want %08x %08x", capture.outer, capture.inner, radii[0], radii[1])
		}
		if capture.center != center || capture.source != source || capture.excluded != excluded ||
			capture.damage != damage || capture.typ != typ {
			t.Fatalf("other arguments changed: %+v", capture)
		}
	}
}
