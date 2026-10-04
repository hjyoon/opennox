package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestInversionSpell52BEB0RealSelectorAndOwnershipService(t *testing.T) {
	for _, mode := range []string{"targeted", "other target", "non magic", "non missile"} {
		t.Run(mode, func(t *testing.T) {
			s := NewServer(nil, nil, strman.New())
			t.Cleanup(s.Close)
			if !s.Objs.Init(3) {
				t.Fatal("cannot initialize native object allocator")
			}
			t.Cleanup(s.Objs.FreeObjects)
			s.Map.Init()
			s.SetFrame(712)
			oldServer, oldGame := noxServer, noxflags.GetGame()
			noxServer = s
			noxflags.ResetGame()
			t.Cleanup(func() { noxServer = oldServer; noxflags.ResetGame(); noxflags.SetGame(oldGame) })
			owner, caster, missile := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
			caster.PosVec = types.Ptf(300, 300)
			missile.PosVec, missile.NewPos = caster.PosVec, caster.PosVec
			missile.ObjClass, missile.ObjSubClass, missile.ObjFlags = object.ClassMissile, object.SubClass(object.MissileMagic), object.FlagActive
			missile.Field32 = 123
			data, freeData := alloc.New(server.MissileUpdateData{})
			t.Cleanup(freeData)
			*data = server.MissileUpdateData{Owner: caster, Target: owner, SpellID: int32(spell.SPELL_MAGIC_MISSILE)}
			missile.UpdateData = unsafe.Pointer(data)
			missile.SetOwner(caster)
			switch mode {
			case "other target":
				data.Target = caster
			case "non magic":
				missile.ObjSubClass = 0
				missile.UpdateData = nil
			case "non missile":
				missile.ObjClass = object.ClassImmobile
				missile.UpdateData = nil
			}
			s.Map.AddObjectToIndex(missile)
			before := *data
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(caster)) <= math.MaxUint32 || uintptr(unsafe.Pointer(missile)) <= math.MaxUint32) {
				t.Fatal("selector fixture is not above 4 GiB")
			}
			if got := legacy.Sub_52BEB0(spell.SPELL_INVERSION, nil, owner, caster, nil, 5); got != 1 {
				t.Fatalf("real selector result=%d", got)
			}
			if mode == "targeted" {
				if data.Target != caster || data.Owner != owner || missile.ObjOwner != owner || missile.Field32 != 712 || caster.Field129 != nil || owner.Field129 != missile {
					t.Fatalf("real inversion ownership: update=%+v owner=%p frame=%d linked=%p/%p", *data, missile.ObjOwner, missile.Field32, caster.Field129, owner.Field129)
				}
			} else if *data != before || missile.ObjOwner != caster || missile.Field32 != 123 {
				t.Fatal("inert missile changed native ownership")
			}
		})
	}
}
