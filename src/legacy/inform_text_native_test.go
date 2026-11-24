package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestInformTextNativeCEntries(t *testing.T) {
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.NetList.Init()
	t.Cleanup(s.NetList.Free)
	old := GetServer
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: s} }
	t.Cleanup(func() { GetServer = old })

	recipients := [...]ntype.PlayerInd{3, 31}
	units, freeUnits := alloc.NewHigh([len(recipients)]server.Object{})
	t.Cleanup(freeUnits)
	updates, freeUpdates := alloc.NewHigh([len(recipients)]server.PlayerUpdateData{})
	t.Cleanup(freeUpdates)
	for i, index := range recipients {
		player := s.Players.ByIndRaw(index)
		player.Active, player.PlayerInd = 1, byte(index)
		player.PlayerUnit = &units[i]
		units[i].ObjClass = object.ClassPlayer
		units[i].UpdateData = unsafe.Pointer(&updates[i])
		updates[i].Player = player
	}
	unitless := s.Players.ByIndRaw(5)
	unitless.Active, unitless.PlayerInd = 1, 5

	number := int32(-0x76543211)
	for _, code := range []int32{0, 1, 2, 12, 13, 16, 20, 21} {
		if result := callNetInformTextC(3, code, &number); result != 1 {
			t.Fatalf("unicast code %d returned %d, want 1", code, result)
		}
		want := []byte{0xa9, byte(code), 0xef, 0xcd, 0xab, 0x89}
		if got := s.NetList.CopyPacketsA(3, netlist.Kind1); !bytes.Equal(got, want) {
			t.Fatalf("unicast code %d sent %x, want %x", code, got, want)
		}
	}
	if result := callNetInformTextC(3, 17, nil); result != 1 {
		t.Fatalf("unicast code 17 returned %d, want 1", result)
	}
	if got := s.NetList.CopyPacketsA(3, netlist.Kind1); !bytes.Equal(got, []byte{0xa9, 17}) {
		t.Fatalf("unicast code 17 sent %x", got)
	}
	for _, code := range []int32{-99, 0x103} {
		if result := callNetInformTextC(3, code, nil); result != code {
			t.Fatalf("unknown unicast code %d returned %d", code, result)
		}
		if result := callNetInformTextBroadcastC(code, nil); result != int64(code) {
			t.Fatalf("unknown broadcast code %d returned %d", code, result)
		}
	}

	for _, tc := range []struct {
		size  int
		codes []int32
	}{
		{6, []int32{3, 4, 8, 18, 19, 21}},
		{10, []int32{5, 6, 7, 9, 10, 11}},
		{11, []int32{14}},
	} {
		for _, code := range tc.codes {
			payload := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0xa5}
			wantInput := bytes.Clone(payload)
			var wantPacket []byte
			if tc.size == 6 {
				wantPacket = []byte{0xa9, byte(code), 1, 2, 3, 4}
			} else {
				wantInput[0], wantInput[1] = 0xa9, byte(code)
				wantPacket = wantInput[:tc.size]
			}
			if result := callNetInformTextBroadcastC(code, &payload[0]); result != 0 {
				t.Fatalf("broadcast code %d returned %d, want 0", code, result)
			}
			if !bytes.Equal(payload, wantInput) {
				t.Fatalf("broadcast code %d left payload %x, want %x", code, payload, wantInput)
			}
			for _, index := range recipients {
				if got := s.NetList.CopyPacketsA(index, netlist.Kind1); !bytes.Equal(got, wantPacket) {
					t.Fatalf("broadcast code %d for slot %d = %x, want %x", code, index, got, wantPacket)
				}
			}
			if got := s.NetList.CopyPacketsA(5, netlist.Kind1); len(got) != 0 {
				t.Fatalf("unitless slot received %x", got)
			}
		}
	}
}
