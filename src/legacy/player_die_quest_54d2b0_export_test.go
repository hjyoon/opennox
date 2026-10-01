package legacy

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// These original scalar fields are private in server. Resolve the declared
// native member offsets rather than writing the original PE32 offsets into a
// C-owned native Player allocation.
func playerDieQuestTestWord54D2B0(t *testing.T, p *server.Player, name string) *uint32 {
	t.Helper()
	field, ok := reflect.TypeOf(*p).FieldByName(name)
	if !ok || field.Type.Kind() != reflect.Uint32 {
		t.Fatalf("missing DWORD member %s", name)
	}
	return (*uint32)(unsafe.Add(unsafe.Pointer(p), field.Offset))
}

func TestPlayerDieQuest54D2B0RealCEntryNativeLivesStatsAndPenalty(t *testing.T) {
	for _, lives := range []uint32{0, 1, 2, 0x80000000, 0xffffffff} {
		t.Run(fmt.Sprintf("%08x", lives), func(t *testing.T) {
			s, _, unit, update, p := questPenaltyLegacyFixture54CBD0(t)
			health, freeHealth := alloc.New(server.HealthData{Cur: 0, Max: 450})
			t.Cleanup(freeHealth)
			questPenaltyLegacyHighPointers54CBD0(t, unsafe.Pointer(health))
			unit.ObjClass, unit.ObjFlags, unit.HealthData = object.ClassPlayer, object.FlagDead|object.FlagEnabled, health
			unit.NetCode, update.ManaCur, update.ExtraLives, update.TrapSpellsCnt = 0x1234, 17, lives, 0xaabbcc05
			update.TrapSpells = [5]uint32{1, 2, 3, 4, 5}
			p.PlayerUnit, p.PlayerInd, p.GoldVal = unit, 31, 101
			p.Info().SetPlayerClass(player.Class(255))
			*playerDieQuestTestWord54D2B0(t, p, "field4660") = 19
			*playerDieQuestTestWord54D2B0(t, p, "field4692") = 0x81
			*playerDieQuestTestWord54D2B0(t, p, "field4668") = 0xaabb1234
			*playerDieQuestTestWord54D2B0(t, p, "field4672") = 0xccdd5678
			*playerDieQuestTestWord54D2B0(t, p, "field4664") = 0xeeff9abc
			*playerDieQuestTestWord54D2B0(t, p, "field4688") = 0x1122def0
			update.RespawnMarkers[31] = 0x77

			oldFlags, oldAnkh := noxflags.GetGame(), playerDieAnkhType54D2B0
			noxflags.ResetGame()
			noxflags.SetGame(noxflags.GameModeQuest)
			playerDieAnkhType54D2B0 = 11 // Already-initialized type-cache branch.
			t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags); playerDieAnkhType54D2B0 = oldAnkh })
			// legacy is below the package that registers these two common
			// callbacks. Only those missing outer boundaries are supplied;
			// playerDieCall and all Quest services remain the production binding.
			oldState, oldAbilities := Nox_xxx_playerSetState_4FA020, Nox_xxx_playerCancelAbils_4FC180
			var states, abilities int
			Nox_xxx_playerSetState_4FA020 = func(got *server.Object, state server.PlayerState) bool {
				states++
				if got != unit || state != server.PlayerState3 {
					t.Fatal("native state boundary")
				}
				update.State = state
				return false
			}
			Nox_xxx_playerCancelAbils_4FC180 = func(got *server.Object) {
				abilities++
				if got != unit {
					t.Fatal("native ability boundary")
				}
			}
			t.Cleanup(func() { Nox_xxx_playerSetState_4FA020, Nox_xxx_playerCancelAbils_4FC180 = oldState, oldAbilities })

			var packets [][]byte
			s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
				packets = append(packets, append([]byte(nil), packet...))
				if related != nil {
					t.Fatal("related object must be nil")
				}
				if len(packets) == 1 {
					if recipient != 255 || len(packet) != 3 || binary.LittleEndian.Uint16(packet[1:]) != 0x1234 || remove != 0 || sequence != 1 {
						t.Fatalf("death notification = %d/%x/%d/%d", recipient, packet, remove, sequence)
					}
				} else {
					want := []byte{0xf0, 2, 0x34, 0x12, 0x78, 0x56, 0xbc, 0x9a, 0xf0, 0xde, 0, 0, 0, 0}
					if lives != 0 || len(packets) != 2 || recipient != 31 || !reflect.DeepEqual(packet, want) || remove != 1 || sequence != 0 || p.GoldVal != 101 || *playerDieQuestTestWord54D2B0(t, p, "field4660") != 19 {
						t.Fatalf("pre-reset Quest statistics = %d/%x/%d/%d", recipient, packet, remove, sequence)
					}
				}
				return -77 // Both original notification/statistics returns are ignored.
			}
			playerDieExportCall54D2B0(unit)
			if states != 1 || abilities != 1 || update.State != server.PlayerState3 || update.ManaCur != 0 || update.TrapSpells != [5]uint32{} || update.TrapSpellsCnt != 0xaabbcc00 {
				t.Fatal("common native death did not finish")
			}
			if lives != 0 {
				if len(packets) != 1 || update.ExtraLives != lives-1 || p.GoldVal != 101 || *playerDieQuestTestWord54D2B0(t, p, "field4660") != 20 || *playerDieQuestTestWord54D2B0(t, p, "field4692") != 0x83 || update.RespawnMarkers[31] != 0x77 {
					t.Fatal("native positive-life branch changed non-record state")
				}
			} else {
				// An empty balance file supplies its existing zero default. The
				// stock starting-life value requires a separate headless play check.
				if len(packets) != 2 || p.GoldVal != 51 || update.ExtraLives != 0 || update.RespawnMarkers[31] != 0 || *playerDieQuestTestWord54D2B0(t, p, "field4660") != 0 || *playerDieQuestTestWord54D2B0(t, p, "field4692") != 63 {
					t.Fatal("native zero-life reset/penalty/balance sequence")
				}
			}
		})
	}
}
