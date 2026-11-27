package legacy

import (
	"bytes"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestSecretCreditNative4D61F0(t *testing.T) {
	player := new(server.Player)
	data := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(data)}
	record := (*nox_playerInfo)(player.C())
	record.field_4672, record.field_4676, record.field_4692 = 0xffffffff, 5, 2
	unit.ObjFlags = object.FlagDestroyed
	if sub_4D61F0(asObjectC(unit)) != nil || sub_4D61F0(nil) != nil || record.field_4672 != 0xffffffff {
		t.Fatal("destroyed or missing unit awarded secret credit")
	}
	unit.ObjFlags = object.FlagDead
	if got := sub_4D61F0(asObjectC(unit)); got != record ||
		record.field_4672 != 0 || record.field_4676 != 6 || record.field_4692 != 0x12 {
		t.Fatalf("credit = %p, counters = %d/%d, flags = %#x", got, record.field_4672, record.field_4676, record.field_4692)
	}
}

func TestQuestClosingSettingsNative4D70B0(t *testing.T) {
	state := memmap.PtrUint32(0x5D4594, 1556152)
	settings := sub416640Native()
	flags := (*byte)(unsafe.Add(settings, 100))
	oldState, oldFlags := *state, *flags
	t.Cleanup(func() { *state, *flags = oldState, oldFlags })
	*state, *flags = 0xf0010001, 0xff
	if got := sub_4D70B0(); got != 0xf0010001 || *flags != 0xff {
		t.Fatalf("closed quest result = %#x, flags = %#x", got, *flags)
	}
	*state = 0
	if got := uintptr(sub_4D70B0()); got != uintptr(settings) || *flags != 0xef {
		t.Fatalf("reopened quest result = %#x, flags = %#x", got, *flags)
	}
}

func TestBroadcastPlayerTextNative4D9FD0(t *testing.T) {
	for _, row := range []struct {
		tag  byte
		text []uint16
		want []byte
	}{
		{0x10, []uint16{'a', 0xe9, 0}, []byte{0xa8, 0, 0, 0x12, 0, 0, 0, 0, 3, 0, 0, 'a', 0xe9, 0}},
		{0x10, []uint16{0x17a, 0}, []byte{0xa8, 0, 0, 0x14, 0, 0, 0, 0, 2, 0, 0, 0x7a, 1, 0, 0}},
		{4, []uint16{'x', 0}, []byte{0xa8, 0, 0, 6, 0, 0, 0, 0, 2, 0, 0, 'x', 0, 0, 0}},
	} {
		var packet [11 + 2*255]byte
		if got := broadcastPlayerTextPacket4D9FD0(&packet, row.tag, &row.text[0]); !bytes.Equal(got, row.want) {
			t.Errorf("chat payload for %x = %x, want %x", row.text, got, row.want)
		}
	}
	var text [256]uint16
	for i := range text[:255] {
		text[i] = 'x'
	}
	var packet [11 + 2*255]byte
	if got := broadcastPlayerTextPacket4D9FD0(&packet, 0, &text[0]); len(got) != 11 || got[8] != 0 {
		t.Fatalf("16-bit text count header = %x", got)
	}
}

func TestBroadcastPlayerTextVariadicCEntry4D9FD0(t *testing.T) {
	world := lineMessageTestWorld(t)
	player := world.Players.NewRaw(1)
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	update.Player = player
	unit.ObjClass, unit.UpdateData = object.ClassPlayer, unsafe.Pointer(update)
	player.PlayerUnit = unit
	t.Cleanup(func() { player.PlayerUnit = nil })

	for _, tc := range []struct {
		name, format, message string
	}{
		{"formatted words", "%S %04d%s", "xp 0037ź"},
		{"length byte wraps", strings.Repeat("x", 256), strings.Repeat("x", 255)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			format, freeFormat := alloc.CString16(tc.format)
			t.Cleanup(freeFormat)
			const tag = 0x10
			if got := cfixture.BroadcastMessage4D9FD0(tag, format, 37); got != 0 {
				t.Fatalf("broadcast result = %d, want 0", got)
			}
			text, freeText := alloc.CString16(tc.message)
			t.Cleanup(freeText)
			var buffer [11 + 2*255]byte
			want := broadcastPlayerTextPacket4D9FD0(&buffer, tag, text)
			if got := world.NetList.CopyPacketsA(player.PlayerIndex(), netlist.Kind1); !bytes.Equal(got, want) {
				t.Fatalf("broadcast packet = %x, want %x", got, want)
			}
		})
	}
}

func TestBoundedWideStringsTerminateAtCapacity(t *testing.T) {
	format, freeFormat := CWString("%s|%s")
	t.Cleanup(freeFormat)
	first, freeFirst := CWString("가")
	t.Cleanup(freeFirst)
	second, freeSecond := CWString("나나다")
	t.Cleanup(freeSecond)
	buffer := [7]wchar2_t{0xa5, 0xa5, 0xa5, 0xa5, 0xa5, 0xa5, 0xa5}
	formatWideStrings(buffer[1:6], format, first, second)
	if want := [7]wchar2_t{0xa5, '가', '|', '나', '나', 0, 0xa5}; buffer != want {
		t.Fatalf("bounded format = %x, want %x", buffer, want)
	}
}
