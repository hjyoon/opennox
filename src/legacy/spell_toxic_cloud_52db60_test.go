package legacy

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestToxicCloudCastGoEntryForwardsNativePointers52DB60(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(12, 34)}
	before := *arg
	previous := toxicCloudCastCall52DB60
	t.Cleanup(func() { toxicCloudCastCall52DB60 = previous })
	for _, level := range []int{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		calls := 0
		toxicCloudCastCall52DB60 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, gotLevel int) int {
			calls++
			if id != spell.SPELL_TOXIC_CLOUD || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || gotLevel != level {
				t.Fatalf("native dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, gotLevel)
			}
			return -17
		}
		if got := Nox_xxx_castToxicCloud_52DB60(spell.SPELL_TOXIC_CLOUD, second, owner, caster, arg, level); got != -17 || calls != 1 || *arg != before {
			t.Fatalf("result=%d calls=%d argument=%+v", got, calls, *arg)
		}
	}
}

// Exercise the real public entry, empty native wall map, type lookup and
// audio service. A missing type must return success, not touch balance or
// call placement; neither a cast result nor trace outcome is supplied.
func TestToxicCloudCastNativeOpenTraceMissingType52DB60(t *testing.T) {
	for _, level := range []int{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			srv, caster, update := monsterLookAtFixture5125A0(t)
			srv.Map.Init()
			t.Cleanup(srv.Map.Free)
			if srv.Walls.Init() == 0 {
				t.Fatal("cannot initialize empty native wall map")
			}
			t.Cleanup(srv.Walls.Free)
			caster.PosVec = types.Ptf(300, 300)
			owner, freeOwner := alloc.New(server.Object{})
			t.Cleanup(freeOwner)
			arg, freeArg := alloc.New(server.SpellAcceptArg{})
			t.Cleanup(freeArg)
			*arg = server.SpellAcceptArg{Obj: nil, Pos: types.Ptf(330, 300)}
			cache := memmap.PtrUint32(0x5D4594, 2487808)
			previousCache, previousServer := *cache, GetServer
			t.Cleanup(func() { *cache, GetServer = previousCache, previousServer })
			*cache = 0xffffffff
			GetServer = func() Server { return &toxicCloudCastLegacyServer52DB60{srv: srv} }
			beforeOwner, beforeCaster, beforeUpdate, beforeArg := *owner, *caster, *update, *arg
			if got := Nox_xxx_castToxicCloud_52DB60(spell.SPELL_TOXIC_CLOUD, nil, owner, caster, arg, level); got != 1 ||
				*cache != 0xffffffff || *owner != beforeOwner || *caster != beforeCaster || *update != beforeUpdate || *arg != beforeArg {
				t.Fatalf("native open trace result=%d cache=%08x", got, *cache)
			}
		})
	}
}
