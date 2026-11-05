package legacy

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestGeneratorCreditActualC4D61B0(t *testing.T) {
	if specialDeathOriginalABI54E460(3) != 64 {
		t.Fatal("typed native object argument required")
	}
	for _, name := range []string{"live", "wrap", "destroyed", "nonplayer-class", "nil"} {
		t.Run(name, func(t *testing.T) {
			unit := specialDeathActualFixture54E460(t)[3]
			update := (*server.PlayerUpdateData)(unit.UpdateData)
			player := update.Player
			counter := uint32(0xaabbccdd)
			if name == "wrap" {
				counter = 0xffffffff
			}
			*playerDieQuestTestWord54D2B0(t, player, "field4668") = counter
			if name == "destroyed" {
				unit.ObjFlags |= object.FlagDestroyed
			} else if name == "nonplayer-class" {
				unit.ObjClass = object.ClassMonster
			}
			records := [][]byte{
				unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))),
				unsafe.Slice((*byte)(unit.UpdateData), int(unsafe.Sizeof(*update))),
				unsafe.Slice((*byte)(unsafe.Pointer(player)), int(unsafe.Sizeof(*player))),
			}
			want := make([][]byte, len(records))
			for i, record := range records {
				want[i] = bytes.Clone(record)
			}
			flags := uint32(0x80000000)
			if name != "destroyed" && name != "nil" {
				counter++
				flags |= 8
				off := specialDeathOriginalLayout54E460()
				binary.LittleEndian.PutUint32(want[2][off[17]:], counter)
				binary.LittleEndian.PutUint32(want[2][off[18]:], flags)
			}
			argument := unit
			if name == "nil" {
				argument = nil
			}
			specialDeathOriginalC54E460(3, argument)
			_, words := specialDeathOriginalPlayer54E460(unit)
			if words != [4]uint32{counter, flags, 0x12345678, 0xabcdef01} {
				t.Fatal("actual C entry and independent C player getter", words)
			}
			for i, record := range records {
				if !bytes.Equal(record, want[i]) {
					t.Fatal("actual constructor object/update/player complete memory guard", i)
				}
			}
		})
	}
}
