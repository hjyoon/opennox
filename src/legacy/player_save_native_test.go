package legacy

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func setPlayerSaveTestFlags(t *testing.T, flags noxflags.GameFlag) {
	t.Helper()
	previous := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(flags)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(previous)
	})
}

func writePlayerSaveTestPayload(t *testing.T, write func(*cryptfile.CryptFile) error) []byte {
	t.Helper()
	path := filepath.Join(t.TempDir(), "player-section.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	if err := write(cf); err != nil {
		_ = cf.Close()
		t.Fatal(err)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func readPlayerSaveTestPayload(t *testing.T, payload []byte, read func(*cryptfile.CryptFile) error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "player-section.bin")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	if err := read(cf); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerAttribWriteNative41A590ExactVersion5Payload(t *testing.T) {
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	info := &server.PlayerInfo{}
	info.SetName("AX")
	raw := unsafe.Slice((*byte)(info.C()), playerInfoPayloadSize41A590)
	for i := 50; i < 89; i++ {
		raw[i] = byte(i)
	}
	pl := &server.Player{QuestStage: 0x55667788}
	update := &server.PlayerUpdateData{ExtraLives: 0x11223344, Player: pl}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}

	got := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		return playerAttribWriteNative41A590(cf, unit, info)
	})
	want, err := hex.DecodeString(
		"0500" +
			"01000000" +
			"02" +
			"41005800" +
			"32333435363738393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f505152535455565758" +
			"44332211" +
			"88776655",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("attribute payload = %x, want %x", got, want)
	}
}

func TestPlayerAttribReadNative41A590UsesNativePlayerLinks(t *testing.T) {
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	sourceInfo := &server.PlayerInfo{}
	sourceInfo.SetName("AX")
	sourceRaw := unsafe.Slice((*byte)(sourceInfo.C()), playerInfoPayloadSize41A590)
	for i := 50; i < 89; i++ {
		sourceRaw[i] = byte(i)
	}
	sourcePlayer := &server.Player{QuestStage: 7}
	sourceUpdate := &server.PlayerUpdateData{ExtraLives: 3, Player: sourcePlayer}
	sourceUnit := &server.Object{UpdateData: unsafe.Pointer(sourceUpdate)}
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		return playerAttribWriteNative41A590(cf, sourceUnit, sourceInfo)
	})

	targetInfo := &server.PlayerInfo{}
	targetPlayer := &server.Player{
		PlayerInd:           4,
		ProtPlayerOrigName:  0x10,
		ProtPlayerField2239: 0x20,
		ProtPlayerField2235: 0x30,
		ProtPlayerClass:     0x40,
		QuestStage:          99,
	}
	targetUpdate := &server.PlayerUpdateData{Player: targetPlayer, ExtraLives: 99}
	targetUnit := &server.Object{NetCode: 0x5566, UpdateData: unsafe.Pointer(targetUpdate)}
	var events []string
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerAttribReadNative41A590(cf, targetUnit, targetInfo, playerAttribReadHooks41A590{
			questMode: func() bool { return false },
			disconnect: func(ind byte) {
				events = append(events, fmt.Sprintf("disconnect:%d", ind))
			},
			protectName: func(token uint32, name []byte) {
				events = append(events, fmt.Sprintf("name:%x:%x", token, name))
			},
			protectInt: func(token, value uint32) {
				events = append(events, fmt.Sprintf("int:%x:%x", token, value))
			},
			protectClass: func(token uint32, class byte) {
				events = append(events, fmt.Sprintf("class:%x:%x", token, class))
			},
			initColors: func(netCode uint32) {
				events = append(events, fmt.Sprintf("colors:%x", netCode))
			},
			maxExtraLives: func() int32 {
				events = append(events, "maximum")
				return 9
			},
			resetQuest: func(unit *server.Object) {
				if unit != targetUnit {
					t.Fatalf("reset unit = %p, want %p", unit, targetUnit)
				}
				events = append(events, "reset")
			},
			sendQuestStage: func(player *server.Player) {
				if player != targetPlayer {
					t.Fatalf("stage player = %p, want %p", player, targetPlayer)
				}
				events = append(events, fmt.Sprintf("stage:%d", player.QuestStage))
			},
		})
	})

	targetRaw := unsafe.Slice((*byte)(targetInfo.C()), playerInfoPayloadSize41A590)
	if !bytes.Equal(targetRaw[:89], sourceRaw[:89]) {
		t.Fatalf("attribute payload = %x, want %x", targetRaw[:89], sourceRaw[:89])
	}
	if targetPlayer.Name() != "AX" || targetUpdate.ExtraLives != 3 || targetPlayer.QuestStage != 7 {
		t.Fatalf("player state = name %q lives %d stage %d", targetPlayer.Name(), targetUpdate.ExtraLives, targetPlayer.QuestStage)
	}
	wantEvents := []string{
		"name:10:41005800",
		"int:20:39383736",
		"int:30:35343332",
		"class:40:42",
		"colors:5566",
		"maximum",
		"reset",
		"stage:7",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerStatusReadNative41AA30RestoresFixedWidthFields(t *testing.T) {
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	sourcePlayer := &server.Player{}
	sourceUpdate := &server.PlayerUpdateData{Player: sourcePlayer, ManaCur: 45, ManaMax: 90}
	sourceUnit := &server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(sourceUpdate),
		HealthData: &server.HealthData{Cur: 55, Max: 100},
		Poison540:  3,
		Field541:   4,
		Field542:   0x5566,
		Experience: 123.5,
		Direction1: server.Dir16(0x7788),
	}
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		return playerStatusWriteNative41AA30(cf, sourceUnit, &server.PlayerInfo{})
	})

	targetPlayer := &server.Player{ProtUnitExperience: 0x12345678}
	targetUpdate := &server.PlayerUpdateData{Player: targetPlayer, ManaCur: 1, ManaMax: 2}
	targetUnit := &server.Object{
		ObjClass:   object.ClassPlayer,
		NetCode:    0x3344,
		UpdateData: unsafe.Pointer(targetUpdate),
		HealthData: &server.HealthData{Cur: 3, Max: 4},
	}
	var events []string
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerStatusReadNative41AA30(cf, targetUnit, playerStatusReadHooks41AA30{
			playerExists: func(netCode uint32) bool {
				events = append(events, fmt.Sprintf("exists:%x", netCode))
				return true
			},
			coopMode: func() bool { return true },
			setMaxHP: func(unit *server.Object, value uint16) {
				events = append(events, fmt.Sprintf("max-hp:%d", value))
				unit.HealthData.Max = value
			},
			setHP: func(unit *server.Object, value uint16) {
				events = append(events, fmt.Sprintf("hp:%d", value))
				unit.HealthData.Cur = value
			},
			setMaxMana: func(_ *server.Object, value uint16) {
				events = append(events, fmt.Sprintf("max-mana:%d", value))
				targetUpdate.ManaMax = value
			},
			refreshMana: func(*server.Object) {
				events = append(events, "refresh-mana")
				targetUpdate.ManaCur = targetUpdate.ManaMax
			},
			storeCurrentHP: func(value uint16) {
				events = append(events, fmt.Sprintf("saved-hp:%d", value))
			},
			storeCurrentMana: func(value uint16) {
				events = append(events, fmt.Sprintf("saved-mana:%d", value))
			},
			setPoison: func(unit *server.Object, value byte) {
				events = append(events, fmt.Sprintf("poison:%d", value))
				unit.Poison540 = value
			},
			protectExperience: func(token uint32, value float32) {
				events = append(events, fmt.Sprintf("protect:%x:%.1f", token, value))
			},
			reportExperience: func(unit *server.Object) {
				if unit != targetUnit {
					t.Fatalf("experience unit = %p, want %p", unit, targetUnit)
				}
				events = append(events, "report")
			},
		})
	})

	if targetUnit.HealthData.Max != 100 || targetUnit.HealthData.Cur != 100 ||
		targetUpdate.ManaMax != 90 || targetUpdate.ManaCur != 90 {
		t.Fatalf("loaded maxima = HP %d/%d mana %d/%d",
			targetUnit.HealthData.Cur, targetUnit.HealthData.Max, targetUpdate.ManaCur, targetUpdate.ManaMax)
	}
	if targetUnit.Poison540 != 3 || targetUnit.Field541 != 4 || targetUnit.Field542 != 0x5566 ||
		targetUnit.Experience != 123.5 || targetUnit.Direction1 != 0x7788 || targetUnit.Direction2 != 0x7788 {
		t.Fatalf("loaded status = poison %d fields %d/%x experience %.1f direction %x/%x",
			targetUnit.Poison540, targetUnit.Field541, targetUnit.Field542, targetUnit.Experience,
			targetUnit.Direction1, targetUnit.Direction2)
	}
	wantEvents := []string{
		"exists:3344", "max-hp:100", "hp:100", "max-mana:90", "refresh-mana",
		"saved-hp:55", "saved-mana:45", "poison:3", "protect:12345678:123.5", "report",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerJournalWriteNative41BEC0OldestFirst(t *testing.T) {
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	newest := &server.PlayerJournal{Field3: 0x5678}
	copy(newest.EntryBuf[:], "new")
	oldest := &server.PlayerJournal{Field3: 0x1234, Prev: newest}
	copy(oldest.EntryBuf[:], "old")
	newest.Next = oldest
	pl := &server.Player{Journal: newest}
	update := &server.PlayerUpdateData{Player: pl}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}

	got := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		return playerJournalWriteNative41BEC0(cf, unit, &server.PlayerInfo{})
	})
	want, err := hex.DecodeString("0100010200036f6c643412036e65777856")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("journal payload = %x, want %x", got, want)
	}
}

func TestPlayerJournalReadNative41BEC0RestoresNewestFirstAtNativeWidth(t *testing.T) {
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		if err := cf.WriteU16(1); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU16(2); err != nil {
			return err
		}
		for _, value := range []struct {
			name      string
			entryType uint16
		}{{"old", 0x1234}, {"new", 0x5678}} {
			if err := cf.WriteU8(byte(len(value.name))); err != nil {
				return err
			}
			if _, err := cf.Write([]byte(value.name)); err != nil {
				return err
			}
			if err := cf.WriteU16(value.entryType); err != nil {
				return err
			}
		}
		return nil
	})

	player := &server.Player{Journal: &server.PlayerJournal{Field3: 1}}
	copy(player.Journal.EntryBuf[:], "stale")
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unit.UpdateData) <= uintptr(^uint32(0)) {
		t.Fatalf("test update-data pointer %p does not exercise the high-address path", update)
	}
	var events []string
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerJournalReadNative41BEC0(cf, unit, playerJournalReadHooks41BEC0{
			coopMode: func() bool { return true },
			removeEntries: func(got *server.Object, mask uint16) {
				if got != unit {
					t.Fatalf("remove unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("remove:%04x", mask))
				player.Journal = nil
			},
			addEntry: func(got *server.Object, name string, entryType uint16) {
				if got != unit {
					t.Fatalf("add unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("add:%s:%04x", name, entryType))
				entry := &server.PlayerJournal{Field3: entryType, Next: player.Journal}
				copy(entry.EntryBuf[:63], name)
				if player.Journal != nil {
					player.Journal.Prev = entry
				}
				player.Journal = entry
			},
		})
	})
	wantEvents := []string{"remove:ffff", "add:old:1234", "add:new:5678"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	newest := player.Journal
	if newest == nil || string(bytes.TrimRight(newest.EntryBuf[:], "\x00")) != "new" || newest.Field3 != 0x5678 {
		t.Fatalf("newest journal entry = %#v, want new/5678", newest)
	}
	oldest := newest.Next
	if oldest == nil || string(bytes.TrimRight(oldest.EntryBuf[:], "\x00")) != "old" || oldest.Field3 != 0x1234 {
		t.Fatalf("oldest journal entry = %#v, want old/1234", oldest)
	}
	if oldest.Prev != newest || newest.Prev != nil || oldest.Next != nil {
		t.Fatalf("journal links = newest.prev %p oldest.prev %p oldest.next %p", newest.Prev, oldest.Prev, oldest.Next)
	}
}

func TestPlayerInventoryQuestLimits41AC30(t *testing.T) {
	if !playerInventoryQuestLimits41AC30(false, func(string) int32 {
		t.Fatal("non-player limit check read inventory")
		return 0
	}, 0) {
		t.Fatal("non-player was rejected")
	}

	counts := make(map[string]int32)
	for _, name := range playerInventoryQuestStackNames41AC30 {
		counts[name] = 9
	}
	counts["InfinitePainWand"] = 3
	count := func(name string) int32 { return counts[name] }
	if !playerInventoryQuestLimits41AC30(true, count, 3) {
		t.Fatal("GAME.EXE boundary inventory was rejected")
	}
	counts["ShieldPotion"] = 10
	if playerInventoryQuestLimits41AC30(true, count, 3) {
		t.Fatal("ten quest potions were accepted")
	}
	counts["ShieldPotion"] = 9
	counts["InfinitePainWand"] = 4
	if playerInventoryQuestLimits41AC30(true, count, 3) {
		t.Fatal("staff count above the balance limit was accepted")
	}
}

func TestPlayerInventoryReadNative41AC30RestoresNativeObjectLinks(t *testing.T) {
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		if err := cf.WriteU16(3); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU32(77); err != nil {
			return err
		}
		if err := cf.WriteU32(1); err != nil {
			return err
		}
		if err := cf.WriteU8(5); err != nil {
			return err
		}
		if _, err := cf.Write([]byte("Sword")); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		for _, value := range []uint32{0x1234, 0x1234, 0x1234} {
			if err := cf.WriteU32(value); err != nil {
				return err
			}
		}
		return cf.WriteU8(7)
	})

	player := &server.Player{PlayerInd: 5, GoldVal: 9, ProtPlayerGold: 0xaabbccdd}
	update := &server.PlayerUpdateData{Player: player, CurTraps: 0x11223300}
	oldItem := &server.Object{ScriptIDVal: 0x7777}
	unit := &server.Object{UpdateData: unsafe.Pointer(update), InvFirstItem: oldItem}
	oldItem.InvHolder = unit
	created := &server.Object{ScriptIDVal: 0x1234, ObjFlags: object.FlagEquipped}
	var events []string

	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerInventoryReadNative41AC30(cf, unit, playerInventoryReadHooks41AC30{
			coopMode:  func() bool { return true },
			questMode: func() bool { return false },
			syncLevel: func(got *server.Object) {
				if got != unit {
					t.Fatalf("sync unit = %p, want %p", got, unit)
				}
				events = append(events, "sync")
			},
			protectGold: func(token uint32, delta int32) {
				events = append(events, fmt.Sprintf("gold:%x:%d", token, delta))
			},
			delayedDelete: func(item *server.Object) {
				events = append(events, fmt.Sprintf("delete:%x", uint32(item.ScriptIDVal)))
				if item == oldItem {
					unit.InvFirstItem = item.InvNextItem
				}
				item.InvHolder = nil
				item.InvNextItem = nil
			},
			newObject: func(name string) *server.Object {
				events = append(events, "new:"+name)
				return created
			},
			transferItem: func(item *server.Object) error {
				events = append(events, "xfer")
				return nil
			},
			questItemAllowed: func(*server.Object) bool { return true },
			placeWorld: func(item, owner *server.Object) {
				if item != created || owner != unit || item.PosVec.X != 2944 || item.PosVec.Y != 2944 {
					t.Fatalf("world placement = %p/%p/%v", item, owner, item.PosVec)
				}
				events = append(events, "world")
			},
			addPending: func() { events = append(events, "pending") },
			placeInventory: func(owner, item *server.Object) bool {
				events = append(events, "place")
				item.InvNextItem = owner.InvFirstItem
				owner.InvFirstItem = item
				item.InvHolder = owner
				return true
			},
			tryDequip: func(owner, item *server.Object) bool {
				events = append(events, "dequip")
				item.ObjFlags &^= object.FlagEquipped
				return true
			},
			tryEquip: func(owner, item *server.Object) bool {
				events = append(events, "equip")
				item.ObjFlags |= object.FlagEquipped
				return true
			},
			clearClientSelection: func() { events = append(events, "clear") },
			reportSecondary: func(ind byte, item *server.Object) {
				events = append(events, fmt.Sprintf("secondary:%d:%x", ind, uint32(item.ScriptIDVal)))
			},
			reportQuiver: func(ind byte, item *server.Object) {
				events = append(events, fmt.Sprintf("quiver:%d:%x", ind, uint32(item.ScriptIDVal)))
			},
			nextScriptID: func() uint32 { return 0x9999 },
			questLimits:  func(*server.Object) bool { return true },
			notifyLoaded: func(ind byte) {
				events = append(events, fmt.Sprintf("loaded:%d", ind))
			},
		})
	})

	if unit.InvFirstItem != created || created.InvHolder != unit {
		t.Fatalf("native inventory links = first %p holder %p, want %p/%p", unit.InvFirstItem, created.InvHolder, created, unit)
	}
	if player.GoldVal != 77 || update.CurTraps != 0x11223307 || !created.Flags().Has(object.FlagEquipped) {
		t.Fatalf("loaded inventory state = gold %d traps %#x flags %#x", player.GoldVal, update.CurTraps, created.ObjFlags)
	}
	wantEvents := []string{
		"sync", "gold:aabbccdd:-9", "gold:aabbccdd:77", "delete:7777", "new:Sword", "xfer",
		"world", "pending", "place", "dequip", "equip", "clear", "secondary:5:1234", "quiver:5:1234", "loaded:5",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerInventoryReadNative41AC30RejectsQuestItemBeforePlacement(t *testing.T) {
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		if err := cf.WriteU16(3); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU32(0); err != nil {
			return err
		}
		if err := cf.WriteU32(1); err != nil {
			return err
		}
		if err := cf.WriteU8(5); err != nil {
			return err
		}
		_, err := cf.Write([]byte("Sword"))
		return err
	})
	path := filepath.Join(t.TempDir(), "quest-inventory.bin")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()

	player := &server.Player{}
	unit := &server.Object{UpdateData: unsafe.Pointer(&server.PlayerUpdateData{Player: player})}
	created := &server.Object{}
	var events []string
	err = playerInventoryReadNative41AC30(cf, unit, playerInventoryReadHooks41AC30{
		coopMode:  func() bool { return false },
		questMode: func() bool { return true },
		syncLevel: func(*server.Object) {
			events = append(events, "sync")
		},
		protectGold: func(uint32, int32) {
			events = append(events, "gold")
		},
		newObject: func(name string) *server.Object {
			events = append(events, "new:"+name)
			return created
		},
		transferItem: func(item *server.Object) error {
			if item != created {
				t.Fatalf("transfer item = %p, want %p", item, created)
			}
			events = append(events, "xfer")
			return nil
		},
		questItemAllowed: func(item *server.Object) bool {
			if item != created || item.PosVec != (types.Pointf{X: 2944, Y: 2944}) {
				t.Fatalf("eligibility item/position = %p/%v, want %p/(2944,2944)", item, item.PosVec, created)
			}
			events = append(events, "eligible")
			return false
		},
		placeWorld: func(*server.Object, *server.Object) {
			t.Fatal("rejected Quest item was placed in the world")
		},
		addPending: func() {
			t.Fatal("rejected Quest item entered the pending list")
		},
	})
	if err == nil || err.Error() != "inventory[0] \"Sword\" is not valid in quest mode" {
		t.Fatalf("read error = %v, want exact Quest eligibility rejection", err)
	}
	want := []string{"sync", "gold", "gold", "new:Sword", "xfer", "eligible"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPlayerFieldbookReadNative41B420RestoresNamedGuides(t *testing.T) {
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		if err := cf.WriteU16(1); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU8(2); err != nil {
			return err
		}
		for _, name := range []string{"Spider", "Bat"} {
			if err := playerSaveWriteName41B420(cf, name); err != nil {
				return err
			}
		}
		return nil
	})

	unit := &server.Object{NetCode: 0x12345678}
	var events []string
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerFieldbookReadNative41B420(cf, unit, playerFieldbookReadHooks41B420{
			playerExists: func(netCode uint32) bool {
				events = append(events, fmt.Sprintf("exists:%x", netCode))
				return true
			},
			coopMode:  func() bool { return false },
			questMode: func() bool { return true },
			guideByName: func(name string) int {
				events = append(events, "name:"+name)
				if name == "Spider" {
					return 24
				}
				return 1
			},
			questGuideAllowed: func(guide int) bool {
				events = append(events, fmt.Sprintf("allowed:%d", guide))
				return true
			},
			awardGuide: func(got *server.Object, guide int) {
				if got != unit {
					t.Fatalf("award unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("award:%d", guide))
			},
		})
	})

	wantEvents := []string{
		"exists:12345678", "name:Spider", "allowed:24", "award:24",
		"name:Bat", "allowed:1", "award:1",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerFieldbookAwardLoadNative41B420UsesNativePlayerLink(t *testing.T) {
	player := &server.Player{PlayerInd: 7, Prot4640: 0x12345678}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var events []string
	if !playerFieldbookAwardLoadNative41B420(unit, 24, playerFieldbookAwardHooks41B420{
		awardProtection: func(token uint32, guide, level int) {
			events = append(events, fmt.Sprintf("protect:%x:%d:%d", token, guide, level))
		},
		relatedGuides: func(guide int) []int {
			events = append(events, fmt.Sprintf("related:%d", guide))
			return []int{7, 8, 25, 26}
		},
		reportAward: func(gotUnit *server.Object, gotPlayer *server.Player, guide int) {
			if gotUnit != unit || gotPlayer != player {
				t.Fatalf("report state = %p/%p, want %p/%p", gotUnit, gotPlayer, unit, player)
			}
			events = append(events, fmt.Sprintf("report:%d:%d", gotPlayer.PlayerInd, guide))
		},
	}) {
		t.Fatal("valid guide was not restored")
	}
	for _, guide := range []int{24, 7, 8, 25, 26} {
		if player.BeastScrollLvl[guide] != 1 {
			t.Fatalf("guide %d level = %d, want 1", guide, player.BeastScrollLvl[guide])
		}
	}
	wantEvents := []string{
		"protect:12345678:24:1", "related:24", "protect:12345678:7:1",
		"protect:12345678:8:1", "protect:12345678:25:1", "protect:12345678:26:1",
		"report:7:24",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if playerFieldbookAwardLoadNative41B420(unit, 24, playerFieldbookAwardHooks41B420{
		awardProtection: func(uint32, int, int) { t.Fatal("duplicate guide changed protection") },
		relatedGuides:   func(int) []int { t.Fatal("duplicate guide visited relations"); return nil },
		reportAward:     func(*server.Object, *server.Player, int) { t.Fatal("duplicate guide was reported") },
	}) {
		t.Fatal("duplicate guide was restored twice")
	}
}

func TestPlayerSpellbookReadNative41B660RestoresWarriorAbility(t *testing.T) {
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		if err := cf.WriteU16(3); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := playerSaveWriteName41B420(cf, "ABILITY_WARCRY"); err != nil {
			return err
		}
		return cf.WriteU32(5)
	})

	player := &server.Player{}
	player.Info().SetPlayerClass(0)
	unit := &server.Object{NetCode: 0x55667788}
	var events []string
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerSpellbookReadNative41B660(cf, unit, playerSpellbookReadHooks41B660{
			playerByID: func(netCode uint32) *server.Player {
				events = append(events, fmt.Sprintf("player:%x", netCode))
				return player
			},
			coopMode:  func() bool { return true },
			questMode: func() bool { return false },
			spellByName: func(name string) int {
				t.Fatalf("warrior version 3 resolved spell %q", name)
				return 0
			},
			abilityByName: func(name string) int {
				events = append(events, "ability:"+name)
				return 2
			},
			questSpellAllowed: func(int) bool {
				t.Fatal("non-Quest spellbook validated a spell")
				return false
			},
			grantSpell: func(*server.Object, int, int32) {
				t.Fatal("warrior version 3 granted a spell")
			},
			awardAbility: func(got *server.Object, ability int) {
				if got != unit {
					t.Fatalf("ability unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("award:%d", ability))
			},
		})
	})
	wantEvents := []string{"player:55667788", "ability:ABILITY_WARCRY", "award:2"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerSpellGrantLoadNative41B660RestoresFamilyAtNativeWidth(t *testing.T) {
	player := &server.Player{PlayerInd: 4, Prot4636: 0xaabbccdd}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var events []string
	if !playerSpellGrantLoadNative41B660(unit, 10, 2, playerSpellGrantLoadHooks41B660{
		coopOrQuest: func() bool { return true },
		questMode:   func() bool { return false },
		hasFlags: func(spellID int, flags things.SpellFlags) bool {
			if spellID == 10 && flags == things.SpellFlags(0x1000) {
				return true
			}
			return flags == things.SpellFlags(0x2000) && (spellID == 10 || spellID == 11)
		},
		validSpell: func(spellID int) bool { return spellID == 10 || spellID == 11 },
		awardProtection: func(token uint32, spellID, level int) {
			events = append(events, fmt.Sprintf("protect:%x:%d:%d", token, spellID, level))
		},
		reportAward: func(gotUnit *server.Object, gotPlayer *server.Player, spellID int) {
			if gotUnit != unit || gotPlayer != player {
				t.Fatalf("report state = %p/%p, want %p/%p", gotUnit, gotPlayer, unit, player)
			}
			events = append(events, fmt.Sprintf("report:%d:%d", gotPlayer.PlayerInd, spellID))
		},
	}) {
		t.Fatal("valid spell was not restored")
	}
	if player.SpellLvl[10] != 2 || player.SpellLvl[11] != 2 {
		t.Fatalf("family levels = %d/%d, want 2/2", player.SpellLvl[10], player.SpellLvl[11])
	}
	wantEvents := []string{
		"protect:aabbccdd:10:2", "protect:aabbccdd:10:2",
		"protect:aabbccdd:10:2", "report:4:10",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerEnchantmentReadNative41B9C0RestoresWarriorStateAtNativeWidth(t *testing.T) {
	const enchantName = "ENCHANT_SHIELD"
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		if err := cf.WriteU16(5); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU8(byte(len(enchantName))); err != nil {
			return err
		}
		if _, err := cf.Write([]byte(enchantName)); err != nil {
			return err
		}
		if err := cf.WriteU16(0); err != nil {
			return err
		}
		if err := cf.WriteU8(3); err != nil {
			return err
		}
		if err := cf.WriteU32(77); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU8(1); err != nil {
			return err
		}
		if err := cf.WriteU32(123); err != nil {
			return err
		}
		for _, cooldown := range []uint32{20, 0, 40, 50} {
			if err := cf.WriteU32(cooldown); err != nil {
				return err
			}
		}
		return nil
	})

	player := &server.Player{}
	player.Info().SetPlayerClass(0)
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unit.UpdateData) <= uintptr(^uint32(0)) {
		t.Fatalf("test update-data pointer %p does not exercise the high-address path", update)
	}
	var events []string
	fpsCalls := 0
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerEnchantmentReadNative41B9C0(cf, unit, playerEnchantmentReadHooks41B9C0{
			coopMode: func() bool { return true },
			clearPlayerSpells: func() {
				events = append(events, "clear-player")
			},
			clearDurationSpells: func() {
				events = append(events, "clear-duration")
			},
			parseEnchant: server.ParseEnchant,
			applyEnchant: func(got *server.Object, enchant server.EnchantID, power byte) {
				if got != unit {
					t.Fatalf("enchantment unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("apply:%d:%d", enchant, power))
				got.Buffs |= uint32(1) << enchant
				got.BuffsPower[enchant] = power
			},
			gameFPS: func() uint32 {
				fpsCalls++
				fps := uint32(29 + fpsCalls)
				events = append(events, fmt.Sprintf("fps:%d", fps))
				return fps
			},
			setShieldHealth: func(got *server.Object, health uint32) {
				if got != unit {
					t.Fatalf("shield unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("shield:%d", health))
			},
			stopBerserker: func() {
				events = append(events, "stop")
			},
			executeAbility: func(got *server.Object, ability server.Ability) {
				if got != unit {
					t.Fatalf("ability unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("execute:%d", ability))
			},
			setAbilityDuration: func(got *server.Object, ability server.Ability, duration int) {
				if got != unit {
					t.Fatalf("duration unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("duration:%d:%d", ability, duration))
			},
			setCooldown: func(got *server.Object, ability server.Ability, cooldown int) {
				if got != unit {
					t.Fatalf("cooldown unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("cooldown:%d:%d", ability, cooldown))
			},
			reportCooldown: func(got *server.Object, ability server.Ability) {
				if got != unit {
					t.Fatalf("report unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("report:%d", ability))
			},
		})
	})

	if !unit.HasEnchant(server.ENCHANT_SHIELD) || unit.BuffsPower[server.ENCHANT_SHIELD] != 3 {
		t.Fatalf("shield state = active %t, power %d; want active power 3", unit.HasEnchant(server.ENCHANT_SHIELD), unit.BuffsPower[server.ENCHANT_SHIELD])
	}
	if unit.BuffsDur[server.ENCHANT_SHIELD] != 30 {
		t.Fatalf("shield duration = %d, want first gameFPS value 30", unit.BuffsDur[server.ENCHANT_SHIELD])
	}
	wantEvents := []string{
		"clear-player", "clear-duration", "apply:26:3", "fps:30", "fps:31", "shield:77",
		"stop", "execute:4", "duration:4:123",
		"cooldown:2:20", "report:2", "cooldown:3:0",
		"cooldown:4:40", "report:4", "cooldown:5:50", "report:5",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerEnchantmentReadNative41B9C0Version4AbilityTailWithoutPayload(t *testing.T) {
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		if err := cf.WriteU16(4); err != nil {
			return err
		}
		if err := cf.WriteU8(0); err != nil {
			return err
		}
		if err := cf.WriteU8(0); err != nil {
			return err
		}
		if err := cf.WriteU8(0); err != nil {
			return err
		}
		if err := cf.WriteU32(^uint32(0)); err != nil {
			return err
		}
		for _, cooldown := range []uint32{1, 0, 3, 0, 5} {
			if err := cf.WriteU32(cooldown); err != nil {
				return err
			}
		}
		return nil
	})
	player := &server.Player{}
	player.Info().SetPlayerClass(0)
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}
	var events []string
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerEnchantmentReadNative41B9C0(cf, unit, playerEnchantmentReadHooks41B9C0{
			coopMode:          func() bool { return true },
			clearPlayerSpells: func() { events = append(events, "clear-player") },
			clearDurationSpells: func() {
				events = append(events, "clear-duration")
			},
			setCooldown: func(_ *server.Object, ability server.Ability, cooldown int) {
				events = append(events, fmt.Sprintf("cooldown:%d:%d", ability, cooldown))
			},
			reportCooldown: func(_ *server.Object, ability server.Ability) {
				events = append(events, fmt.Sprintf("report:%d", ability))
			},
		})
	})
	wantEvents := []string{
		"clear-player", "clear-duration",
		"cooldown:1:1", "report:1", "cooldown:2:0",
		"cooldown:3:3", "report:3", "cooldown:4:0",
		"cooldown:5:5", "report:5",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerAbilityCooldownStart41B9C0(t *testing.T) {
	if got := playerAbilityCooldownStart41B9C0(false); got != 1 {
		t.Fatalf("inactive Berserker start = %d, want 1", got)
	}
	if got := playerAbilityCooldownStart41B9C0(true); got != 2 {
		t.Fatalf("active Berserker start = %d, want 2", got)
	}
}

func TestPlayerGameReadNative41C080RestoresFixedMapBufferAtNativeWidth(t *testing.T) {
	mapPayload := []byte{'W', 'a', 'r', '0', '1', 'a', 0xff, 7, 8, 9, 10, 11}
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		if err := cf.WriteU16(5); err != nil {
			return err
		}
		if err := cf.WriteU32(0x11223344); err != nil {
			return err
		}
		if err := cf.WriteU16(6); err != nil {
			return err
		}
		if _, err := cf.Write(mapPayload); err != nil {
			return err
		}
		if err := cf.WriteU16(0xbeef); err != nil {
			return err
		}
		if err := cf.WriteU8(0xab); err != nil {
			return err
		}
		return cf.WriteU8(0x7c)
	})

	player := &server.Player{}
	for i := range player.SaveNameBuf {
		player.SaveNameBuf[i] = 0xcc
	}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unit.UpdateData) <= uintptr(^uint32(0)) {
		t.Fatalf("test update-data pointer %p does not exercise the high-address path", update)
	}
	var events []string
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerGameReadNative41C080(cf, unit, playerGameReadHooks41C080{
			onlineMode: func() bool { return false },
			readQuestJournal: func(got *cryptfile.CryptFile) error {
				value, err := got.ReadU16()
				if err != nil {
					return err
				}
				events = append(events, fmt.Sprintf("quest:%04x", value))
				return nil
			},
			transferWalls: func(got *cryptfile.CryptFile, gotUnit *server.Object) bool {
				if gotUnit != unit {
					t.Fatalf("wall unit = %p, want %p", gotUnit, unit)
				}
				value, err := got.ReadU8()
				if err != nil {
					t.Fatal(err)
				}
				events = append(events, fmt.Sprintf("walls:%02x", value))
				return true
			},
			setStage: func(stage byte) {
				events = append(events, fmt.Sprintf("stage:%02x", stage))
			},
		})
	})

	wantMap := []byte{'W', 'a', 'r', '0', '1', 'a', 0, 7, 8, 9, 10, 11}
	if !bytes.Equal(player.SaveNameBuf[:len(wantMap)], wantMap) {
		t.Fatalf("map-name buffer = %x, want %x", player.SaveNameBuf[:len(wantMap)], wantMap)
	}
	for index, value := range player.SaveNameBuf[len(wantMap):] {
		if value != 0xcc {
			t.Fatalf("map-name tail byte %d = %02x, want cc", index+len(wantMap), value)
		}
	}
	wantEvents := []string{"quest:beef", "walls:ab", "stage:7c"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerMapNamePayload41C080FixedBuffer(t *testing.T) {
	pl := &server.Player{}
	for i := range pl.SaveNameBuf {
		pl.SaveNameBuf[i] = 0xcc
	}
	got, err := playerMapNamePayload41C080(pl, "War01a")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{'W', 'a', 'r', '0', '1', 'a', 0, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc}
	if !bytes.Equal(got, want) {
		t.Fatalf("map-name payload = %x, want %x", got, want)
	}
	if len(pl.SaveNameBuf) != 32 {
		t.Fatalf("map-name buffer = %d bytes, want 32", len(pl.SaveNameBuf))
	}
	if delta := unsafe.Offsetof(pl.Field4792) - unsafe.Offsetof(pl.SaveNameBuf); delta != 32 {
		t.Fatalf("Field4792 follows map-name buffer by %d bytes, want 32", delta)
	}
	if _, err := playerMapNamePayload41C080(pl, "12345678901234567"); err == nil {
		t.Fatal("17-byte doubled map name fit in the 32-byte GAME.EXE buffer")
	}
}
