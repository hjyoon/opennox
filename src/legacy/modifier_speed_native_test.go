package legacy

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type originalSpeedModifierRow struct {
	Function string
	Name     string
	Mode     string
	Inputs   struct {
		ValueBits                uint32       `json:"value_bits"`
		CoefficientBits          uint32       `json:"coefficient_bits"`
		OwnerShape               string       `json:"owner_shape"`
		ClassWord                object.Class `json:"class_word"`
		InitialFlags             uint32       `json:"initial_flags"`
		PlayerIndex              uint8        `json:"player_index"`
		RemainingSpeedItem       bool         `json:"remaining_speed_item"`
		RemainingSlot            int          `json:"remaining_slot"`
		UnusedDisengageFloatBits uint32       `json:"unused_disengage_float_bits"`
	}
	ValueBits uint32 `json:"value_bits"`
	FlagsWord uint32 `json:"flags_word"`
	Services  []struct {
		Service     string
		PlayerIndex uint8 `json:"player_index"`
		Kind        uint8
		ValueBits   uint32 `json:"value_bits"`
		Sound       sound.ID
	}
}

func TestModifierSpeedOriginalNativeC(t *testing.T) {
	raw, err := os.ReadFile("testdata/modifier_speed_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256                       string `json:"original_sha256"`
		OriginalUnchanged                    bool   `json:"original_unchanged"`
		OriginalInventorySpeedBranchExecuted bool   `json:"original_inventory_speed_branch_executed"`
		RawRows                              int    `json:"raw_rows"`
		Rows                                 []originalSpeedModifierRow
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		!fixture.OriginalUnchanged || !fixture.OriginalInventorySpeedBranchExecuted ||
		fixture.RawRows != 272 || len(fixture.Rows) != fixture.RawRows {
		t.Fatal("incomplete independent original speed/inventory observations")
	}

	srv := objectStatusTestServer(t, 68)
	setImportantRecipientMaskC(^uint32(0))
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	owner, freeOwner := alloc.NewHigh(server.Object{})
	t.Cleanup(freeOwner)
	player, freePlayer := alloc.NewHigh(server.Player{})
	t.Cleanup(freePlayer)
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	item, freeItem := alloc.NewHigh(server.Object{})
	t.Cleanup(freeItem)
	attrs, freeAttrs := alloc.NewHigh(server.ModifierInitData{})
	t.Cleanup(freeAttrs)
	modifier, freeModifier := alloc.NewHigh(server.ModifierEff{})
	t.Cleanup(freeModifier)
	remaining, freeRemaining := alloc.NewHigh(server.ModifierEff{})
	t.Cleanup(freeRemaining)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, address := range []unsafe.Pointer{owner.CObj(), player.C(), unsafe.Pointer(update), item.CObj(), unsafe.Pointer(attrs), modifier.C(), remaining.C()} {
			if uintptr(address) <= uintptr(math.MaxUint32) {
				t.Fatalf("fixture address %p must exceed 4 GiB", address)
			}
		}
	}

	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	for _, id := range []sound.ID{sound.SoundHasteOn, sound.SoundHasteOff} {
		file := &doorSoundFile53AC50{
			name:                              id.String(),
			playerAttackHammerAudioFile538960: playerAttackHammerAudioFile538960{data: []byte{7, 1, 'x', 0, 0}},
		}
		if !srv.Audio.ReadAVNT(file) {
			t.Fatalf("could not load sound %v", id)
		}
	}
	type soundObservation struct {
		id     sound.ID
		kind   int
		owner  *server.Object
		packet unsafe.Pointer
	}
	var sounds []soundObservation
	srv.Audio.OnSound(func(id sound.ID, kind int, obj *server.Object, _ types.Pointf) {
		packet, _ := importantNativeListC()
		sounds = append(sounds, soundObservation{id, kind, obj, packet})
	})

	callbacks := map[string]unsafe.Pointer{
		"SpeedEngage":    modifierEngagePointerNative4DFBB0(16),
		"SpeedDisengage": modifierDisengagePointerNative4DFBB0(16),
	}
	counts := make(map[string]int, len(callbacks))
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		counts[row.Function]++
		t.Run(row.Function+"/"+row.Name, func(t *testing.T) {
			in := row.Inputs
			*owner = server.Object{
				ObjClass: in.ClassWord, NetCode: 0x2357, Field110: in.InitialFlags,
				SpeedBonus: math.Float32frombits(in.ValueBits), SpeedBase: 18.25,
				UpdateData: unsafe.Pointer(update), InvFirstItem: item,
			}
			player.PlayerInd = in.PlayerIndex
			update.Player = player
			*item = server.Object{ObjClass: object.ClassWand, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(attrs)}
			*attrs = server.ModifierInitData{}
			attrs.Modifiers[in.RemainingSlot] = remaining
			remaining.Engage112 = modifierEngagePointerNative4DFBB0(1)
			if in.RemainingSpeedItem {
				remaining.Engage112 = callbacks["SpeedEngage"]
			}
			modifier.EngageFloat120 = math.Float32frombits(in.CoefficientBits)
			modifier.DisengageFloat128 = math.Float32frombits(in.UnusedDisengageFloatBits)

			arg := owner.CObj()
			if in.OwnerShape == "null" {
				arg = nil
			}
			callback := callbacks[row.Function]
			if callback == nil {
				t.Fatalf("native %s callback is missing", row.Function)
			}
			srv.Audio.Reset()
			sounds = sounds[:0]
			oldPacket, _ := importantNativeListC()
			ccall.CallVoidPtr3(callback, modifier.C(), arg, item.CObj())
			if got := math.Float32bits(owner.SpeedBonus); got != row.ValueBits || owner.Field110 != row.FlagsWord || owner.SpeedBase != 18.25 {
				t.Errorf("speed bits %08x flags %08x base %v, want %08x %08x 18.25", got, owner.Field110, owner.SpeedBase, row.ValueBits, row.FlagsWord)
			}
			packet, _ := importantNativeListC()
			if len(row.Services) == 0 {
				if len(sounds) != 0 || packet != oldPacket {
					t.Fatal("guarded native callback emitted a sound or speed report")
				}
				return
			}
			if len(row.Services) != 2 {
				t.Fatalf("original speed service count %d, want two", len(row.Services))
			}
			report, audio := row.Services[0], row.Services[1]
			if report.Service != "speed-report" || audio.Service != "sound" {
				t.Fatal("invalid original speed service order")
			}
			if packet == oldPacket {
				t.Fatal("native callback omitted the speed report")
			}
			want := [8]byte{0: 0x68, 1: 0x57, 2: 0x23, 7: report.Kind}
			binary.LittleEndian.PutUint32(want[3:7], report.ValueBits)
			checkObjectStatusPacket(t, packet, report.PlayerIndex, want[:], 0, 1)
			if len(sounds) != 1 || sounds[0] != (soundObservation{audio.Sound, int(audio.Kind), owner, packet}) {
				t.Errorf("queued sound observations %+v, want %v after the packet", sounds, audio.Sound)
			}
		})
	}
	if counts["SpeedEngage"] != 34 || counts["SpeedDisengage"] != 34 {
		t.Fatalf("incomplete original startup contexts %v", counts)
	}
}
