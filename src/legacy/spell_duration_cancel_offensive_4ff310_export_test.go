package legacy

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellDurationCancelOffensiveLegacyServer4FF310 struct {
	Server
	srv *server.Server
}

func (s *spellDurationCancelOffensiveLegacyServer4FF310) S() *server.Server {
	return s.srv
}

func setSpellDurationCancelOffensiveFlags4FF310(
	t *testing.T,
	srv *server.Server,
	defs map[spell.ID]*server.SpellDef,
) {
	t.Helper()
	field := reflect.ValueOf(&srv.Spells).Elem().FieldByName("byID")
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(
		reflect.ValueOf(defs),
	)
}

func TestSpellDurationCancelOffensiveExport4FF310PreservesNativePointer(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	setSpellDurationCancelOffensiveFlags4FF310(t, srv, map[spell.ID]*server.SpellDef{
		31: {Def: things.Spell{Flags: things.SpellOffensive}},
	})

	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationCancelOffensiveLegacyServer4FF310{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	caster, freeCaster := alloc.New(server.Object{})
	other, freeOther := alloc.New(server.Object{})
	record, freeRecord := alloc.New(server.DurSpell{})
	t.Cleanup(freeCaster)
	t.Cleanup(freeOther)
	t.Cleanup(freeRecord)

	record.Caster16 = caster
	record.Spell = 31
	record.Flags88 = 0xa1b2c300
	srv.Spells.Dur.List = record

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"caster": unsafe.Pointer(caster),
			"other":  unsafe.Pointer(other),
			"record": unsafe.Pointer(record),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	spellDurationCancelOffensiveExportCall4FF310(caster)
	if record.Flags88 != 0xa1b2c301 {
		t.Fatalf("matching CGo call flags = %#x, want 0xa1b2c301", record.Flags88)
	}

	record.Flags88 = 0xa1b2c300
	spellDurationCancelOffensiveExportCall4FF310(other)
	if record.Flags88 != 0xa1b2c300 {
		t.Fatalf("nonmatching CGo call flags = %#x, want unchanged", record.Flags88)
	}

	record.Caster16 = nil
	spellDurationCancelOffensiveExportCall4FF310(nil)
	if record.Flags88 != 0xa1b2c301 {
		t.Fatalf("nil-identity CGo call flags = %#x, want 0xa1b2c301", record.Flags88)
	}

	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
	runtime.KeepAlive(record)
}
