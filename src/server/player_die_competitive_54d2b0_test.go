package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

type playerDieCompetitiveRecord54D2B0 struct {
	Params  []uint32   `json:"params"`
	Events  [][]uint32 `json:"events"`
	Units   [][]uint32 `json:"units"`
	Updates [][]uint32 `json:"updates"`
	Players [][]uint32 `json:"players"`
	Fault   bool       `json:"fault"`
}

// This fixture mirrors only the explicitly declared original service
// boundaries. The scoring roots' internals have independent original-byte
// transcripts; real production death binding is tested separately in legacy.
func playerDieCompetitiveRecordNative54D2B0(t *testing.T, parameters []uint32) playerDieCompetitiveRecord54D2B0 {
	t.Helper()
	flags, rivals, sourceID, trackingMode, limit, deaths, mutation, invalid := parameters[0], parameters[1], parameters[2], parameters[3], parameters[4], parameters[5], parameters[6], parameters[7]
	const frame, rate = uint32(0x12344700), uint32(30)
	units := [5]Object{}
	updates := [2]PlayerUpdateData{}
	players := [3]Player{}
	trade := new(TradeSession)
	for i := range units {
		units[i].ObjClass = object.Class(0x80000004)
		if i == 3 {
			units[i].ObjClass = object.Class(0x80000000)
		} else if i == 4 {
			units[i].ObjClass = object.Class(0x80000002)
		}
		units[i].ObjFlags = object.Flags(0x8000)
		units[i].NetCode = uint32(0x12340001 + i)
		units[i].TypeInd = uint16(0x1234 + i)
		units[i].Buffs = 0x12345678
		units[i].BuffsDur[2], units[i].BuffsPower[2] = 77, 8
	}
	units[0].UpdateData = unsafe.Pointer(&updates[0])
	units[0].HealthData = &HealthData{Cur: 0, Max: 20}
	units[3].ObjOwner = &units[1]
	if sourceID != 0 {
		ind := sourceID
		if ind == 2 {
			ind = 0
		}
		units[0].Obj130 = &units[ind]
	}
	units[0].Field131 = 8
	if sourceID == 4 {
		units[0].Field131 = 16
	}
	for i := range updates {
		updates[i].Player = &players[0]
		if i != 0 {
			updates[i].Player = &players[2]
		}
		updates[i].ManaCur, updates[i].Field47_0 = uint16(17+i), uint8(9+i)
		for j := range updates[i].TrapSpells {
			updates[i].TrapSpells[j] = uint32(1 + j + 10*i)
		}
		updates[i].TrapSpellsCnt, updates[i].SpellCastStart = uint32(0xaabbcc05+i), uint32(99+i)
		updates[i].Field75, updates[i].Field76 = 2, 0xaabbcc02
		updates[i].ExtraLives, updates[i].Trade70 = 0x80000001, trade
	}
	for i := range players {
		players[i].PlayerUnit = &units[i]
		players[i].Field3600 = 0x80000000
		if trackingMode == 0 && i == 0 {
			players[i].Field3600 = 0
		}
		players[i].Field3604 = 0x81234567
		players[i].SetLastAggressorFrame(frame - 256)
		if trackingMode == 1 {
			players[i].SetLastAggressorFrame(frame - 300)
		}
		*(*uint32)(unsafe.Pointer(&players[i].Active)) = 0x80000000
		if trackingMode == 5 && i == 2 {
			*(*uint32)(unsafe.Pointer(&players[i].Active)) = 0
		}
		players[i].NetCodeVal = uint32(0x12340001 + i)
		players[i].Field2140 = uint32(0x11223344 + i)
		if i == 0 {
			players[i].Field2140 = deaths
		}
		players[i].ProtUnitManaCur = uint32(0x12345678 + i)
	}
	idUnit := func(unit *Object) uint32 {
		if unit == nil {
			return 0xffffffff
		}
		for i := range units {
			if unit == &units[i] {
				return uint32(i)
			}
		}
		panic("unrecognized object identity")
	}
	idPlayer := func(player *Player) uint32 {
		if player == nil {
			return 0xffffffff
		}
		for i := range players {
			if player == &players[i] {
				return uint32(i)
			}
		}
		panic("unrecognized Player identity")
	}
	idUpdate := func(pointer unsafe.Pointer) uint32 {
		if pointer == nil {
			return 0xffffffff
		}
		for i := range updates {
			if pointer == unsafe.Pointer(&updates[i]) {
				return uint32(i)
			}
		}
		panic("unrecognized update identity")
	}
	result := playerDieCompetitiveRecord54D2B0{Params: parameters, Events: make([][]uint32, 0)}
	event := func(values ...uint32) {
		result.Events = append(result.Events, values)
		if mutation == 0 || uint32(len(result.Events)) != mutation {
			return
		}
		units[0].UpdateData = unsafe.Pointer(&updates[1])
		updates[0].Player = &players[2]
		if invalid != 0 {
			updates[0].Player = nil
		}
		players[0].Field3604 = 0xfedcba98
		players[2].Field2140 = 1
		players[2].Info().SetIsFemale(1)
		flags ^= 0x510
		limit = 0xffff
	}
	armed := false
	rt := PlayerDieRuntime54D2B0{
		GameFlag: func(mask uint32) bool {
			answer := flags&mask != 0
			if armed {
				value := uint32(0)
				if answer {
					value = 1
				}
				event(2, mask, flags, value)
			}
			return answer
		},
		Frame:    func() uint32 { event(3, frame); return frame },
		TickRate: func() uint32 { event(4, rate); return rate },
		PlayerByIndex: func(index uint32) *Player {
			event(5, index)
			switch trackingMode {
			case 2, 5:
				return &players[2]
			case 3:
				return &players[1]
			case 4:
				return &players[0]
			}
			return nil
		},
		ObjectByNetCode: func(code uint32) *Object {
			event(6, code)
			switch trackingMode {
			case 2:
				return &units[2]
			case 3:
				return &units[1]
			}
			return &units[0]
		},
		InformText: func(code int, packet [14]byte) {
			values := []uint32{7, uint32(code)}
			for _, value := range packet {
				values = append(values, uint32(value))
			}
			event(values...)
		},
		ResetAbility:      func(unit *Object, ability int32) { event(8, idUnit(unit), uint32(ability)) },
		GameplayHasRivals: func() bool { event(9, rivals); return rivals != 0 },
		PrepareAnkhType:   func() { armed = true; event(1, 77) },
		CancelPendingSave: func() { event(26) },
		Audio:             func(id int, unit *Object) { event(10, uint32(id), idUnit(unit)) },
		SetPlayerState: func(unit *Object, state PlayerState) bool {
			event(11, idUnit(unit), uint32(state))
			unit.UpdateDataPlayer().State = state
			return false
		},
		RemoveActionShadow: func(unit *Object) { event(17, idUnit(unit)) },
		DropAllItems:       func(unit *Object) int32 { event(18, idUnit(unit)); return -1 },
		NotifyPlayerDied:   func(unit *Object) { event(19, idUnit(unit)) },
		ProtectMana:        func(token uint32, delta int16) { event(20, token, uint32(uint16(delta))) },
		SetBuffFlags:       func(unit *Object, value uint32) { event(21, idUnit(unit), value); unit.Buffs = value },
		CancelAbilities:    func(unit *Object) { event(22, idUnit(unit)) },
		CancelSpells:       func(unit *Object) { event(23, idUnit(unit)) },
		CancelTrade: func(session *TradeSession) {
			if session != trade {
				t.Fatal("unexpected native trade identity")
			}
			event(24)
		},
		Quest: &PlayerDieQuestRuntime54D2B0{
			SendStats:    func(uint8, [14]byte) { t.Fatal("unexpected zero-life branch") },
			RecordDeath:  func(unit *Object) unsafe.Pointer { event(25, idUnit(unit)); return unsafe.Pointer(unit) },
			ResetPlayer:  func(*Object) { t.Fatal("unexpected zero-life reset") },
			Penalty:      func(*Object) { t.Fatal("unexpected zero-life penalty") },
			BalanceFloat: func(string) float32 { t.Fatal("unexpected zero-life balance"); return 0 },
		},
		Unsupported: func(reason string, _ *Object) { t.Fatalf("valid fixture rejected: %s", reason) },
	}
	competitive := PlayerDieCompetitiveRuntime54D2B0{
		Arena: func(victim, killer, assist *Object, tracking uint32) {
			event(12, idUnit(victim), idUnit(killer), idUnit(assist), tracking)
		},
		Kotr:          func(victim, killer *Object) { event(13, idUnit(victim), idUnit(killer)) },
		Elimination:   func(victim, killer *Object) { event(14, idUnit(victim), idUnit(killer)) },
		GameDataLimit: func(mask uint16) uint16 { value := uint16(limit); event(15, uint32(mask), uint32(value)); return value },
		RemoveSpawned: func(unit *Object) { event(16, idUnit(unit)) },
	}
	func() {
		defer func() {
			if recover() != nil {
				result.Fault = true
			}
		}()
		if !PlayerDieCompetitiveNative54D2B0(&units[0], rt, competitive) {
			t.Fatal("complete native death not handled")
		}
	}()
	for i := range units {
		unit := &units[i]
		result.Units = append(result.Units, []uint32{idUpdate(unit.UpdateData), uint32(unit.ObjFlags), unit.Buffs, uint32(unit.BuffsDur[2]), uint32(unit.BuffsPower[2])})
	}
	for i := range updates {
		update := &updates[i]
		row := []uint32{idPlayer(update.Player), uint32(update.ManaCur), uint32(update.State), uint32(update.Field47_0), update.SpellCastStart, update.TrapSpellsCnt}
		row = append(row, update.TrapSpells[:]...)
		tradeActive := uint32(0)
		if update.Trade70 != nil {
			tradeActive = 1
		}
		row = append(row, update.Field76, update.ExtraLives, tradeActive)
		result.Updates = append(result.Updates, row)
	}
	for i := range players {
		player := &players[i]
		female := uint32(0)
		if player.Info().IsFemale() {
			female = 1
		}
		result.Players = append(result.Players, []uint32{player.Field3600, player.Field3604, player.Field2140, player.ProtUnitManaCur, female})
	}
	return result
}

func TestPlayerDieCompetitive54D2B0OriginalTranscript(t *testing.T) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	write := func(parameters []uint32) {
		t.Helper()
		if err := encoder.Encode(playerDieCompetitiveRecordNative54D2B0(t, parameters)); err != nil {
			t.Fatal(err)
		}
	}
	modes := []uint32{0x2100, 0x2010, 0x2400, 0x2110, 0x2500, 0x2410, 0x2510, 0x2800, 0x1000, 0x2000}
	for _, mode := range modes {
		for _, rivals := range []uint32{0, 0x80000000} {
			for source := uint32(0); source < 5; source++ {
				for tracking := uint32(0); tracking < 6; tracking++ {
					for _, limit := range []uint32{0, 1, 0xffff} {
						for _, deaths := range []uint32{0, 1, 0xffff, 0x7fffffff, 0x80000000, 0xffffffff} {
							write([]uint32{mode, rivals, source, tracking, limit, deaths, 0, 0})
						}
					}
				}
			}
		}
	}
	for _, mode := range modes {
		for callback := uint32(1); callback <= 32; callback++ {
			for _, invalid := range []uint32{0, 1} {
				write([]uint32{mode, 0x80000000, 3, 2, 1, 1, callback, invalid})
			}
		}
	}
	if path := os.Getenv("NOX_COMPETITIVE_DEATH_SNAPSHOT"); path != "" {
		if err := os.WriteFile(path, output.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	// Frozen from unchanged 0054D2B0 and 004EC580 instructions, with
	// explicitly declared scoring/state/audio/network/cleanup services.
	const want = "bde16fe883fb22ea791f0ca5bfbce196ef4423ce980f33f4f67dc4e07988be91"
	if got := fmt.Sprintf("%x", sha256.Sum256(output.Bytes())); got != want {
		t.Fatalf("original complete death transcript = %s, want %s", got, want)
	}
}

func TestPlayerDieCompetitive54D2B0AdmissionIsAtomic(t *testing.T) {
	for _, remove := range []func(*PlayerDieCompetitiveRuntime54D2B0){
		func(r *PlayerDieCompetitiveRuntime54D2B0) { r.Arena = nil },
		func(r *PlayerDieCompetitiveRuntime54D2B0) { r.Kotr = nil },
		func(r *PlayerDieCompetitiveRuntime54D2B0) { r.Elimination = nil },
		func(r *PlayerDieCompetitiveRuntime54D2B0) { r.GameDataLimit = nil },
		func(r *PlayerDieCompetitiveRuntime54D2B0) { r.RemoveSpawned = nil },
	} {
		unit, update, player := playerDieFixture54D2B0()
		var events []string
		runtime := playerDieRuntime54D2B0(t, &events)
		runtime.GameFlag = func(mask uint32) bool { return mask == playerDieElimMode54D2B0 }
		var reason string
		runtime.Unsupported = func(value string, _ *Object) { reason = value }
		competitive := PlayerDieCompetitiveRuntime54D2B0{
			Arena: func(*Object, *Object, *Object, uint32) {}, Kotr: func(*Object, *Object) {},
			Elimination: func(*Object, *Object) {}, GameDataLimit: func(uint16) uint16 { return 0 }, RemoveSpawned: func(*Object) {},
		}
		remove(&competitive)
		beforeUnit, beforeUpdate, beforePlayer := *unit, *update, *player
		if PlayerDieCompetitiveNative54D2B0(unit, runtime, competitive) || reason != "missing competitive death service" {
			t.Fatalf("incomplete service set admitted: %q", reason)
		}
		if *unit != beforeUnit || *update != beforeUpdate || *player != beforePlayer || len(events) != 0 {
			t.Fatal("incomplete competitive admission changed state")
		}
	}
}
