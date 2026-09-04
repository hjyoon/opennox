package opennox

import (
	"encoding/binary"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestDecodeTrySpellPacket(t *testing.T) {
	wantSpells := [5]int32{1, 0, 0x12345678, -1, 136}
	data := make([]byte, trySpellPacketSize+3)
	data[0] = byte(netmsg.MSG_TRY_SPELL)
	for i, sp := range wantSpells {
		binary.LittleEndian.PutUint32(data[1+4*i:], uint32(sp))
	}
	data[21] = 0xa5
	copy(data[22:], []byte{0xde, 0xad, 0xbe})

	spells, targetMode, ok := decodeTrySpellPacket(data)
	if !ok {
		t.Fatal("decodeTrySpellPacket rejected a complete packet")
	}
	if spells != wantSpells || targetMode != 0xa5 {
		t.Fatalf("decoded packet = (%v, %#x), want (%v, %#x)", spells, targetMode, wantSpells, byte(0xa5))
	}
}

func TestDecodeTrySpellPacketRejectsIncompleteOrWrongOpcode(t *testing.T) {
	data := make([]byte, trySpellPacketSize)
	data[0] = byte(netmsg.MSG_TRY_SPELL)
	for n := 0; n < trySpellPacketSize; n++ {
		if _, _, ok := decodeTrySpellPacket(data[:n]); ok {
			t.Fatalf("decodeTrySpellPacket accepted %d-byte packet", n)
		}
	}
	data[0] = byte(netmsg.MSG_TRY_ABILITY)
	if _, _, ok := decodeTrySpellPacket(data); ok {
		t.Fatal("decodeTrySpellPacket accepted the wrong opcode")
	}
}

func TestMagicEntityNextSpellBounds(t *testing.T) {
	it := &server.MagicEntityClass{Spells8: [5]int32{10, 20, 30, 40, 50}}
	for i, want := range []int32{20, 30, 40, 50, 0} {
		it.SpellInd28 = uint8(i)
		if got := magicEntityNextSpell(it); got != want {
			t.Fatalf("next spell at %d = %d, want %d", i, got, want)
		}
	}
	it.SpellInd28 = 0xff
	if got := magicEntityNextSpell(it); got != 0 {
		t.Fatalf("next spell for invalid index = %d, want 0", got)
	}
}

func TestMagicEntityUnlinkPreservesNativeLinks(t *testing.T) {
	if magicEntityHead != nil || magicEntityAlloc.Class != nil {
		t.Fatal("magic-entity queue unexpectedly initialized before test")
	}
	magicEntityAlloc = alloc.NewClassT("magicEntityClassTest", server.MagicEntityClass{}, 3)
	t.Cleanup(func() {
		magicEntityHead = nil
		magicEntityAlloc.Free()
		magicEntityAlloc = alloc.ClassT[server.MagicEntityClass]{}
	})

	first := magicEntityAlloc.NewObject()
	middle := magicEntityAlloc.NewObject()
	last := magicEntityAlloc.NewObject()
	first.Next52 = middle
	middle.Prev56 = first
	middle.Next52 = last
	last.Prev56 = middle
	magicEntityHead = first

	if got := magicEntityUnlink(middle); got != last {
		t.Fatalf("unlink middle returned %p, want %p", got, last)
	}
	if magicEntityHead != first || first.Next52 != last || last.Prev56 != first {
		t.Fatalf("middle unlink links = head %p, first.next %p, last.prev %p", magicEntityHead, first.Next52, last.Prev56)
	}

	if got := magicEntityUnlink(first); got != last {
		t.Fatalf("unlink head returned %p, want %p", got, last)
	}
	if magicEntityHead != last || last.Prev56 != nil {
		t.Fatalf("head unlink links = head %p, last.prev %p", magicEntityHead, last.Prev56)
	}

	if got := magicEntityUnlink(last); got != nil || magicEntityHead != nil {
		t.Fatalf("unlink tail = next %p, head %p; want nil, nil", got, magicEntityHead)
	}
}
