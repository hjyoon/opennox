package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type questPenaltyLegacyServer54CBD0 struct {
	Server
	native *server.Server
	delete func(*server.Object)
}

func (s *questPenaltyLegacyServer54CBD0) S() *server.Server                 { return s.native }
func (s *questPenaltyLegacyServer54CBD0) DelayedDelete(item *server.Object) { s.delete(item) }

func questPenaltyLegacyFixture54CBD0(t *testing.T) (*server.Server, *questPenaltyLegacyServer54CBD0, *server.Object, *server.PlayerUpdateData, *server.Player) {
	t.Helper()
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	outer := &questPenaltyLegacyServer54CBD0{native: s, delete: func(*server.Object) { t.Fatal("unexpected deletion") }}
	oldServer := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = oldServer })
	types := questGemTypeCache54D080()
	oldTypes := [3]uint32{*types[0], *types[1], *types[2]}
	t.Cleanup(func() {
		for slot := range types {
			*types[slot] = oldTypes[slot]
		}
	})
	for slot := range types {
		*types[slot] = uint32(11 + slot)
	}
	u, freeUnit := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	p, freePlayer := alloc.New(server.Player{})
	t.Cleanup(freeUnit)
	t.Cleanup(freeUpdate)
	t.Cleanup(freePlayer)
	u.UpdateData, update.Player = unsafe.Pointer(update), p
	questPenaltyLegacyHighPointers54CBD0(t, unsafe.Pointer(u), unsafe.Pointer(update), unsafe.Pointer(p))
	return s, outer, u, update, p
}

func questPenaltyLegacyHighPointers54CBD0(t *testing.T, pointers ...unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range pointers {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("C-owned pointer %p must exceed 4 GiB", pointer)
			}
		}
	}
}

func TestQuestPenalty54CBD0NativeCGoAllLossHelpersRNGAndGold(t *testing.T) {
	s, _, u, update, p := questPenaltyLegacyFixture54CBD0(t)
	// Independent sorted eligible IDs from the sealed GAME.EXE tables, not
	// queried from the loss helper under test. Each second draw uses the
	// remaining candidates after the preceding clear and packet.
	spellIDs := []int{1, 4, 5, 8, 10, 12, 13, 14, 16, 21, 22, 23, 24, 26, 29, 35, 36, 37, 38, 39, 42, 43, 50, 51, 52, 54, 58, 60, 61, 62, 64, 67, 71, 72, 74, 128, 129, 130, 132, 134, 135, 136}
	guideIDs := []int{2, 3, 4, 9, 10, 11, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 26, 27, 29, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}
	for _, class := range []uint8{0, 1, 2, 128, 255} {
		for seed := 0; seed < 48; seed++ {
			t.Run(fmt.Sprintf("class-%d-seed-%d", class, seed), func(t *testing.T) {
				*p = server.Player{PlayerInd: 255, GoldVal: []uint32{0, 1, 2, 3, 0x7fffffff, 0x80000000, 0xfffffffe, math.MaxUint32}[seed%8], Level: 9}
				p.Info().SetPlayerClass(player.Class(class))
				for id := range p.SpellLvl {
					p.SpellLvl[id] = uint32(id) | 0x80000000
				}
				for id := range p.BeastScrollLvl {
					p.BeastScrollLvl[id] = 1
				}
				u.ObjClass, u.Worth = 0, 0xaabbccdd // Original dispatcher has no Player-class bit guard.
				beforeUnit, beforeUpdate, wantPlayer := *u, *update, *p
				wantPlayer.GoldVal -= wantPlayer.GoldVal >> 1
				s.Rand.Logic, s.Rand.Other = prand.New(seed), prand.New(seed+11)
				otherIndex := s.Rand.Other.Index()
				wantRNG := prand.New(seed)
				type expected struct {
					packet []byte
					player server.Player
				}
				var wantPackets []expected
				if class == 1 || class == 2 {
					candidates := append([]int(nil), spellIDs...)
					for i := 0; i < 2; i++ {
						ordinal := wantRNG.IntClamp(0, len(candidates)-1)
						id := candidates[ordinal]
						wantPlayer.SpellLvl[id] = 0
						wantPackets = append(wantPackets, expected{[]byte{0xf0, 0x11, byte(id), 0}, wantPlayer})
						candidates = append(candidates[:ordinal], candidates[ordinal+1:]...)
					}
				}
				if class == 2 {
					candidates := append([]int(nil), guideIDs...)
					for i := 0; i < 2; i++ {
						ordinal := wantRNG.IntClamp(0, len(candidates)-1)
						id := candidates[ordinal]
						wantPlayer.BeastScrollLvl[id] = 0
						wantPackets = append(wantPackets, expected{[]byte{0xf0, 0x13, byte(id), 0}, wantPlayer})
						candidates = append(candidates[:ordinal], candidates[ordinal+1:]...)
					}
				}
				if class == 0 {
					id := wantRNG.IntClamp(1, 5)
					wantPlayer.SpellLvl[id] = 0
					wantPackets = append(wantPackets, expected{[]byte{0xf0, 0x12, byte(id), 0}, wantPlayer})
				}
				packets := 0
				s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
					if packets >= len(wantPackets) {
						t.Fatalf("unexpected packet %x", packet)
					}
					want := wantPackets[packets]
					packets++
					if *p != want.player || recipient != 255 || !reflect.DeepEqual(packet, want.packet) || related != nil || remove != 1 || sequence != 0 {
						t.Fatalf("packet=%d/%x/%p/%d/%d or player state mismatch", recipient, packet, related, remove, sequence)
					}
					return -257 // Ability's signed low-byte result is ignored by the dispatcher.
				}
				questPenaltyCEntry54CBD0(u)
				if packets != len(wantPackets) || *p != wantPlayer || *u != beforeUnit || *update != beforeUpdate || s.Rand.Logic.Index() != wantRNG.Index() || s.Rand.Other.Index() != otherIndex {
					t.Fatalf("class=%d seed=%d packets=%d or native/RNG state mismatch", class, seed, packets)
				}
				if s.Rand.Logic.IntClamp(-123, 456) != wantRNG.IntClamp(-123, 456) {
					t.Fatal("logic RNG state diverged")
				}
			})
		}
	}
}

func TestQuestPenalty54CBD0NativeCGoOrderedInventoryAndCachedClass(t *testing.T) {
	s, outer, u, entryUpdate, entryPlayer := questPenaltyLegacyFixture54CBD0(t)
	liveUpdate, freeLive := alloc.New(server.PlayerUpdateData{})
	livePlayer, freePlayer := alloc.New(server.Player{})
	latePlayer, freeLate := alloc.New(server.Player{})
	finalPlayer, freeFinal := alloc.New(server.Player{})
	for _, free := range []func(){freeLive, freePlayer, freeLate, freeFinal} {
		t.Cleanup(free)
	}
	entryPlayer.GoldVal = 101
	entryPlayer.Info().SetPlayerClass(player.Class(2))
	livePlayer.GoldVal = 200
	livePlayer.Info().SetPlayerClass(player.Class(2))
	latePlayer.Info().SetPlayerClass(player.Class(0))
	finalPlayer.Info().SetPlayerClass(player.Class(255))
	liveUpdate.Player = livePlayer
	items := make([]*server.Object, 10)
	for slot := range items {
		item, free := alloc.New(server.Object{})
		t.Cleanup(free)
		items[slot] = item
		questPenaltyLegacyHighPointers54CBD0(t, unsafe.Pointer(item))
	}
	for slot := range items {
		if slot+1 < len(items) {
			items[slot].InvNextItem = items[slot+1]
		}
	}
	for slot, id := range []uint16{11, 12, 12, 13, 13, 13} {
		items[slot].TypeInd = id
	}
	items[0].Worth, items[3].Worth = 11, 21
	items[6].ObjClass, items[6].ObjFlags = 0x1000, 0x100
	items[7].ObjClass = 0x1000 // A spare, using the ordinary class callback.
	items[8].ObjClass, items[8].ObjFlags, items[8].TypeInd = 0x02000000, 0x100, 65535
	items[9].ObjClass, items[9].ObjFlags, items[9].TypeInd = 0x02000000, 0x100, 32768
	u.InvFirstItem, u.Worth = items[0], 0xaabbccdd
	beforeUnit, beforeEntry, beforeLive, beforeEntryPlayer, beforeLivePlayer, beforeLate, beforeFinal := *u, *entryUpdate, *liveUpdate, *entryPlayer, *livePlayer, *latePlayer, *finalPlayer
	wantItems := make([]server.Object, len(items))
	for slot, item := range items {
		wantItems[slot] = *item
	}
	oldCanUse := Nox_xxx_playerClassCanUseItem_57B3D0
	t.Cleanup(func() { Nox_xxx_playerClassCanUseItem_57B3D0 = oldCanUse })
	var deleted []*server.Object
	canUseCalls := 0
	Nox_xxx_playerClassCanUseItem_57B3D0 = func(item *server.Object, class player.Class) bool {
		canUseCalls++
		if item != items[7] || class != 2 || len(deleted) != 4 || entryPlayer.GoldVal != 51 || livePlayer.GoldVal != 215 {
			t.Fatalf("spare callback class=%d deletes=%d gold=%d/%d", class, len(deleted), entryPlayer.GoldVal, livePlayer.GoldVal)
		}
		return true
	}
	s.Rand.Logic, s.Rand.Other = prand.New(81), prand.New(92)
	wantRNG := prand.New(81)
	firstArmor := items[8+wantRNG.IntClamp(0, 1)]
	secondArmor := items[8]
	if firstArmor == secondArmor {
		secondArmor = items[9]
	}
	_ = wantRNG.IntClamp(0, 0)
	wantDeleted := []*server.Object{items[0], items[1], items[3], items[4], items[6], firstArmor, secondArmor}
	outer.delete = func(item *server.Object) {
		ordinal := len(deleted)
		if ordinal >= len(wantDeleted) || item != wantDeleted[ordinal] || entryPlayer.GoldVal != 51 {
			t.Fatalf("deletion %d item=%p gold=%d", ordinal, item, entryPlayer.GoldVal)
		}
		if ordinal == 0 {
			if livePlayer.GoldVal != 200 {
				t.Fatal("odd gem credited before deletion")
			}
			u.UpdateData = unsafe.Pointer(liveUpdate)
		} else if ordinal == 1 || ordinal == 2 {
			if livePlayer.GoldVal != 205 {
				t.Fatal("first odd credit/order mismatch")
			}
		} else if livePlayer.GoldVal != 215 {
			t.Fatal("second odd credit/order mismatch")
		}
		deleted = append(deleted, item)
		// Legal callback mutation: remove the current item from the live
		// inventory so the next helper must recount the remaining candidates.
		if u.InvFirstItem == item {
			u.InvFirstItem = item.InvNextItem
		} else {
			for previous := u.InvFirstItem; previous != nil; previous = previous.InvNextItem {
				if previous.InvNextItem == item {
					previous.InvNextItem = item.InvNextItem
					for slot, candidate := range items {
						if candidate == previous {
							wantItems[slot].InvNextItem = item.InvNextItem
						}
					}
					break
				}
			}
		}
		if ordinal == 5 {
			entryUpdate.Player = latePlayer // Dispatcher must read this late class zero, not unit's class two.
		} else if ordinal == 6 {
			liveUpdate.Player = finalPlayer // Later helpers must load the current unit update, not dispatcher entry.
		}
	}
	s.NetSendPacketXxx = func(int, []byte, *server.Object, int, int) int {
		t.Fatal("final raw class 255 must not send a learned-item loss packet")
		return 0
	}
	questPenaltyCEntry54CBD0(u)
	if !reflect.DeepEqual(deleted, wantDeleted) || canUseCalls != 1 || s.Rand.Logic.Index() != wantRNG.Index() || s.Rand.Other.Index() != 92 {
		t.Fatalf("deletes=%v class callbacks=%d or RNG mismatch", deleted, canUseCalls)
	}
	beforeUnit.UpdateData, beforeUnit.InvFirstItem = unsafe.Pointer(liveUpdate), items[2]
	beforeEntry.Player, beforeLive.Player = latePlayer, finalPlayer
	beforeEntryPlayer.GoldVal, beforeLivePlayer.GoldVal = 51, 215
	if *u != beforeUnit || *entryUpdate != beforeEntry || *liveUpdate != beforeLive || *entryPlayer != beforeEntryPlayer || *livePlayer != beforeLivePlayer || *latePlayer != beforeLate || *finalPlayer != beforeFinal {
		t.Fatal("unrelated native fields changed")
	}
	for slot, item := range items {
		if *item != wantItems[slot] {
			t.Fatalf("unrelated inventory fields changed at %d", slot)
		}
	}
	questPenaltyLegacyHighPointers54CBD0(t, unsafe.Pointer(liveUpdate), unsafe.Pointer(livePlayer), unsafe.Pointer(latePlayer), unsafe.Pointer(finalPlayer))
}

func TestQuestPenalty54CBD0NativeCGoPacketChangesLaterHelperBindings(t *testing.T) {
	s, _, u, entryUpdate, entryPlayer := questPenaltyLegacyFixture54CBD0(t)
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	second, freeSecond := alloc.New(server.Player{})
	third, freeThird := alloc.New(server.Player{})
	for _, free := range []func(){freeUpdate, freeSecond, freeThird} {
		t.Cleanup(free)
	}
	entryPlayer.Info().SetPlayerClass(player.Class(1))
	entryPlayer.PlayerInd, entryPlayer.GoldVal = 7, 11
	entryPlayer.SpellLvl[1], entryPlayer.SpellLvl[4] = 1, 1
	second.Info().SetPlayerClass(player.Class(2))
	second.PlayerInd, second.GoldVal, second.SpellLvl[128] = 9, 77, 1
	third.Info().SetPlayerClass(player.Class(2))
	third.PlayerInd, third.GoldVal, third.BeastScrollLvl[40] = 255, 99, 1
	third.SpellLvl[1], third.SpellLvl[2], third.SpellLvl[5] = 1, 1, 1
	update.Player = second
	beforeUnit, beforeEntryUpdate, beforeUpdate, wantEntry, wantSecond, wantThird := *u, *entryUpdate, *update, *entryPlayer, *second, *third
	wantRNG := prand.New(81)
	firstSpell := []int{1, 4}[wantRNG.IntClamp(0, 1)]
	_ = wantRNG.IntClamp(1, 1) // The second spell helper's fresh singleton.
	_ = wantRNG.IntClamp(1, 1) // The first guide helper's fresh singleton.
	// The second guide rejects the class changed by its predecessor's
	// packet callback, so only the final ability helper draws next.
	ability := []int{1, 2, 5}[wantRNG.IntClamp(0, 2)]
	wantEntry.GoldVal, wantEntry.SpellLvl[firstSpell] = 6, 0
	wantSecond.SpellLvl[128] = 0
	wantThird.BeastScrollLvl[40] = 0
	wantThird.Info().SetPlayerClass(player.Class(0))
	wantThird.SpellLvl[ability] = 0
	packets := 0
	s.Rand.Logic, s.Rand.Other = prand.New(81), prand.New(92)
	s.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
		if related != nil || remove != 1 || sequence != 0 {
			t.Fatal("loss packet transport mismatch")
		}
		var want []byte
		switch packets {
		case 0:
			want = []byte{0xf0, 0x11, byte(firstSpell), 0}
			if recipient != 7 || *entryPlayer != wantEntry {
				t.Fatal("first spell/gold state mismatch")
			}
			u.UpdateData = unsafe.Pointer(update)
		case 1:
			want = []byte{0xf0, 0x11, 128, 0}
			if recipient != 9 || *second != wantSecond {
				t.Fatal("second spell did not load the new update")
			}
			update.Player = third
		case 2:
			want = []byte{0xf0, 0x13, 40, 0}
			before := wantThird
			before.Info().SetPlayerClass(player.Class(2))
			before.SpellLvl[ability] = 1
			if recipient != 255 || *third != before {
				t.Fatal("guide did not load the new Player link")
			}
			third.Info().SetPlayerClass(player.Class(0))
		case 3:
			want = []byte{0xf0, 0x12, byte(ability), 0}
			if recipient != 255 || *third != wantThird {
				t.Fatal("ability did not observe the late Warrior class")
			}
		default:
			t.Fatalf("unexpected packet %d/%x", recipient, data)
		}
		if !reflect.DeepEqual(data, want) {
			t.Fatalf("packet=%x, want %x", data, want)
		}
		packets++
		return -257
	}
	questPenaltyCEntry54CBD0(u)
	beforeUnit.UpdateData, beforeUpdate.Player = unsafe.Pointer(update), third
	if packets != 4 || *u != beforeUnit || *entryUpdate != beforeEntryUpdate || *update != beforeUpdate || *entryPlayer != wantEntry || *second != wantSecond || *third != wantThird || s.Rand.Logic.Index() != wantRNG.Index() || s.Rand.Other.Index() != 92 {
		t.Fatal("fresh helper bindings, untouched fields or RNG state mismatch")
	}
	questPenaltyLegacyHighPointers54CBD0(t, unsafe.Pointer(update), unsafe.Pointer(second), unsafe.Pointer(third))
}

func TestQuestPenalty54CBD0NativeCGoRepeatedSingletonAndEmptyLosses(t *testing.T) {
	s, _, u, _, p := questPenaltyLegacyFixture54CBD0(t)
	for _, class := range []uint8{0, 1, 2, 255} {
		for _, learned := range []bool{false, true} {
			*p = server.Player{GoldVal: 1}
			p.Info().SetPlayerClass(player.Class(class))
			if learned {
				p.SpellLvl[1], p.BeastScrollLvl[2] = 1, 1
			}
			s.Rand.Logic = prand.New(12)
			wantRNG := prand.New(12)
			var want [][]byte
			if class == 0 {
				if learned {
					_ = wantRNG.IntClamp(1, 1)
					want = append(want, []byte{0xf0, 0x12, 1, 0})
				} else {
					_ = wantRNG.IntClamp(1, 0)
				}
			}
			if class == 1 || class == 2 {
				if learned {
					_ = wantRNG.IntClamp(1, 1)
					want = append(want, []byte{0xf0, 0x11, 1, 0})
				} else {
					_ = wantRNG.IntClamp(1, 0)
				}
				_ = wantRNG.IntClamp(1, 0)
			}
			if class == 2 {
				if learned {
					_ = wantRNG.IntClamp(1, 1)
					want = append(want, []byte{0xf0, 0x13, 2, 0})
				} else {
					_ = wantRNG.IntClamp(1, 0)
				}
				_ = wantRNG.IntClamp(1, 0)
			}
			var packets [][]byte
			s.NetSendPacketXxx = func(_ int, data []byte, _ *server.Object, _, _ int) int {
				packets = append(packets, append([]byte(nil), data...))
				return 0
			}
			before := *p
			questPenaltyCEntry54CBD0(u)
			if learned && (class == 0 || class == 1 || class == 2) {
				before.SpellLvl[1] = 0
			}
			if learned && class == 2 {
				before.BeastScrollLvl[2] = 0
			}
			if !reflect.DeepEqual(packets, want) || *p != before || s.Rand.Logic.Index() != wantRNG.Index() {
				t.Fatalf("class=%d learned=%v packets=%v, want %v or state mismatch", class, learned, packets, want)
			}
		}
	}
}
