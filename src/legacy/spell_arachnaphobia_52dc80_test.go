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

func TestArachnaphobiaCastGoEntryForwardsNativePointers52DC80(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	t.Cleanup(freeSecond)
	owner, freeOwner := alloc.New(server.Object{})
	t.Cleanup(freeOwner)
	caster, freeCaster := alloc.New(server.Object{})
	t.Cleanup(freeCaster)
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	t.Cleanup(freeArg)
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(12, 34)}
	beforeSecond, beforeOwner, beforeCaster, beforeArg := *second, *owner, *caster, *arg
	previous := arachnaphobiaCastCall52DC80
	t.Cleanup(func() { arachnaphobiaCastCall52DC80 = previous })
	for _, level := range []int{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			calls := 0
			arachnaphobiaCastCall52DC80 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, gotLevel int) int {
				calls++
				if id != spell.SPELL_ARACHNAPHOBIA || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || gotLevel != level {
					t.Fatalf("native dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, gotLevel)
				}
				return -17
			}
			if got := Nox_xxx_spellArachna_52DC80(spell.SPELL_ARACHNAPHOBIA, second, owner, caster, arg, level); got != -17 || calls != 1 ||
				*second != beforeSecond || *owner != beforeOwner || *caster != beforeCaster || *arg != beforeArg {
				t.Fatalf("result=%d calls=%d argument=%+v", got, calls, *arg)
			}
		})
	}
}

// A real empty wall map and the production type allocator reject the absent
// type. The original returns success without placement, sound or a duration.
func TestArachnaphobiaCastNativeOpenTraceMissingType52DC80(t *testing.T) {
	for _, level := range []int{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			srv, caster, update := monsterLookAtFixture5125A0(t)
			srv.Map.Init()
			t.Cleanup(srv.Map.Free)
			if srv.Walls.Init() == 0 {
				t.Fatal("cannot initialize native wall map")
			}
			t.Cleanup(srv.Walls.Free)
			caster.PosVec = types.Ptf(300, 300)
			owner, freeOwner := alloc.New(server.Object{})
			t.Cleanup(freeOwner)
			arg, freeArg := alloc.New(server.SpellAcceptArg{})
			t.Cleanup(freeArg)
			*arg = server.SpellAcceptArg{Obj: nil, Pos: types.Ptf(330, 300)}
			cache := memmap.PtrUint32(0x5D4594, 2487812)
			previousCache, previousServer := *cache, GetServer
			t.Cleanup(func() { *cache, GetServer = previousCache, previousServer })
			*cache = 0xffffffff
			GetServer = func() Server { return &arachnaphobiaCastLegacyServer52DC80{srv: srv} }
			beforeOwner, beforeCaster, beforeUpdate, beforeArg := *owner, *caster, *update, *arg
			if got := Nox_xxx_spellArachna_52DC80(spell.SPELL_ARACHNAPHOBIA, nil, owner, caster, arg, level); got != 1 ||
				*cache != 0xffffffff || *owner != beforeOwner || *caster != beforeCaster || *update != beforeUpdate || *arg != beforeArg {
				t.Fatalf("native open trace result=%d cache=%08x", got, *cache)
			}
		})
	}
}
