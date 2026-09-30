package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type serverSpellMaskLegacyServer454040 struct {
	Server
	srv *server.Server
}

func (s *serverSpellMaskLegacyServer454040) S() *server.Server { return s.srv }

func TestServerSpellMask454040NativeStackAndPackedWords(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldGetServer := GetServer
	GetServer = func() Server { return &serverSpellMaskLegacyServer454040{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	defs := make(map[spell.ID]*server.SpellDef)
	for id := spell.ID(1); id < server.SpellsMax; id++ {
		defs[id] = &server.SpellDef{
			Valid: id%7 != 0, Enabled: id%3 == 0,
			Def: things.Spell{Flags: things.SpellFlags(1 << (24 + uint32(id%4)))},
		}
	}
	for _, id := range []spell.ID{1, 31, 32, 63, 64, 95, 96, 127, 128, 136} {
		defs[id] = &server.SpellDef{Valid: true, Def: things.Spell{Flags: things.SpellClassAny}}
	}
	defs[137] = &server.SpellDef{Valid: true, Def: things.Spell{Flags: things.SpellClassAny}}
	field := reflect.ValueOf(&srv.Spells).Elem().FieldByName("byID")
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(defs))

	want := [7]uint32{0xa1b2c3d4, math.MaxUint32, math.MaxUint32, math.MaxUint32, math.MaxUint32, math.MaxUint32, 0x5e6f7081}
	for id, def := range defs {
		if id > 0 && id < server.SpellsMax && def.Valid && def.Def.Flags&0x7000000 != 0 && !def.Enabled {
			want[1+int(id)/32] &^= 1 << (uint(id) % 32)
		}
	}
	got, address := serverSpellMaskStackCall454040()
	if unsafe.Sizeof(uintptr(0)) == 8 && address <= math.MaxUint32 {
		t.Fatalf("C stack mask address = %#x, want above 4 GiB", address)
	}
	if got != want {
		t.Fatalf("C stack snapshot = %#v, want %#v", got, want)
	}

	heap, free := alloc.New([7]uint32{})
	t.Cleanup(free)
	*heap = [7]uint32{want[0], 0, 0, 0, 0, 0, want[6]}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(heap)) <= math.MaxUint32 {
		t.Fatalf("C heap snapshot address = %p, want above 4 GiB", heap)
	}
	serverSpellMaskCall454040(&heap[1])
	if *heap != want {
		t.Fatalf("C heap snapshot = %#v, want %#v", *heap, want)
	}

	srv.Spells.EnableAll()
	for i := 1; i <= 5; i++ {
		want[i] = math.MaxUint32
	}
	serverSpellMaskCall454040(&heap[1])
	if *heap != want {
		t.Fatalf("reenabled snapshot = %#v, want %#v", *heap, want)
	}
	clear(defs)
	got, _ = serverSpellMaskStackCall454040()
	if got != want {
		t.Fatalf("empty spell table snapshot = %#v, want %#v", got, want)
	}
	serverSpellMaskCall454040(nil)
}
