package legacy

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/internal/porttest"
	"github.com/opennox/opennox/v1/server"
	"math"
	"testing"
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
