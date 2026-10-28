package opennox

import (
	"encoding/json"
	"fmt"
	"image"
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Observations contain copied values only. In particular, no selected object,
// player update record, client drawable or duration-spell pointer is retained.
type e2eTargetObjectObservation struct {
	Present       bool         `json:"present"`
	Extent        uint32       `json:"extent"`
	ScriptID      int32        `json:"script_id"`
	TypeIndex     uint16       `json:"type_index"`
	NetCode       uint32       `json:"netcode"`
	Class         object.Class `json:"class"`
	Flags         object.Flags `json:"flags"`
	Position      types.Pointf `json:"position"`
	HealthPresent bool         `json:"health_present"`
	Health        uint16       `json:"health"`
	MaximumHealth uint16       `json:"maximum_health"`
}

type e2eCharmTargetObservation struct {
	ID      uint16                     `json:"id"`
	Level   uint32                     `json:"level"`
	Flag20  uint32                     `json:"flag20"`
	Pos     types.Pointf               `json:"position"`
	Pos2    types.Pointf               `json:"position2"`
	Frame60 uint32                     `json:"frame60"`
	Frame64 uint32                     `json:"frame64"`
	Frame68 uint32                     `json:"frame68"`
	Target  e2eTargetObjectObservation `json:"target"`
}

type e2ePlayerStatInputs struct {
	Present     bool      `json:"present"`
	Level       uint8     `json:"level"`
	Class       uint8     `json:"class"`
	GameFlags   uint32    `json:"game_flags"`
	Base        [4]uint32 `json:"base_bits32"`
	Maximum     [4]uint32 `json:"maximum_bits32"`
	Warrior     [4]uint32 `json:"warrior_bits32"`
	Multipliers [4]uint32 `json:"multipliers_bits32"`
}

func e2eClassStatWords(s server.ClassStats) [4]uint32 {
	return [4]uint32{math.Float32bits(s.Health), math.Float32bits(s.Mana), math.Float32bits(s.Speed), math.Float32bits(s.Strength)}
}

type e2ePlayerTargetObservation struct {
	Frame                 uint32                       `json:"frame"`
	Mouse                 image.Point                  `json:"mouse"`
	ClientCursor          uint32                       `json:"client_cursor"`
	ClientDrawablePresent bool                         `json:"client_drawable_present"`
	ClientDrawableNetCode uint32                       `json:"client_drawable_netcode"`
	ClientDrawableType    uint32                       `json:"client_drawable_type"`
	ClientGuidePresent    bool                         `json:"client_guide_present"`
	ClientGuideLevels     [41]uint32                   `json:"client_guide_levels"`
	Player                e2eTargetObjectObservation   `json:"player"`
	State                 uint32                       `json:"state"`
	Mana                  uint16                       `json:"mana"`
	MaximumMana           uint16                       `json:"maximum_mana"`
	StatInputs            e2ePlayerStatInputs          `json:"stat_inputs"`
	Poison                byte                         `json:"poison"`
	Stamina               byte                         `json:"stamina"`
	EquippedWeapon        e2eTargetObjectObservation   `json:"equipped_weapon"`
	CursorTarget          e2eTargetObjectObservation   `json:"cursor_target"`
	ServerGuidePresent    bool                         `json:"server_guide_present"`
	ServerGuideLevels     [41]uint32                   `json:"server_guide_levels"`
	CharmLevel            uint32                       `json:"charm_level"`
	OwnedCharmCount       int                          `json:"owned_charm_count"`
	OwnedCharms           [8]e2eCharmTargetObservation `json:"owned_charms"`
	CharmOutputTruncated  bool                         `json:"charm_output_truncated"`
	DurationScanTruncated bool                         `json:"duration_scan_truncated"`
}

func e2eReadTargetObject(obj *server.Object) e2eTargetObjectObservation {
	if obj == nil {
		return e2eTargetObjectObservation{}
	}
	out := e2eTargetObjectObservation{
		Present: true, Extent: obj.Extent, ScriptID: obj.ScriptIDVal,
		TypeIndex: obj.TypeInd, NetCode: obj.NetCode, Class: obj.Class(), Flags: obj.Flags(), Position: obj.PosVec,
	}
	if hp := obj.HealthData; hp != nil {
		out.HealthPresent, out.Health, out.MaximumHealth = true, hp.Cur, hp.Max
	}
	return out
}

func e2eReadPlayerTarget(frame uint32, player *server.Object, durations *server.DurSpell, worlds ...*server.Server) (e2ePlayerTargetObservation, error) {
	if player == nil || !player.Class().Has(object.ClassPlayer) || player.UpdateData == nil {
		return e2ePlayerTargetObservation{}, fmt.Errorf("player-target observation requires an original player update record")
	}
	ud := player.UpdateDataPlayer()
	out := e2ePlayerTargetObservation{
		Frame: frame, Player: e2eReadTargetObject(player), State: uint32(ud.State),
		Mana: ud.ManaCur, MaximumMana: ud.ManaMax, Poison: player.Poison540, Stamina: ud.Stamina,
		EquippedWeapon: e2eReadTargetObject(ud.EquippedWeapon), CursorTarget: e2eReadTargetObject(ud.CursorObj),
	}
	if pl := ud.Player; pl != nil {
		out.ServerGuidePresent, out.ServerGuideLevels = true, pl.BeastScrollLvl
		out.CharmLevel = pl.SpellLvl[spell.SPELL_CHARM]
		if len(worlds) == 1 && worlds[0] != nil {
			world := worlds[0]
			class := pl.Info().PlayerClass()
			out.StatInputs = e2ePlayerStatInputs{Present: true, Level: pl.Level, Class: uint8(class), GameFlags: uint32(noxflags.GetGame()),
				Base: e2eClassStatWords(*world.Players.BaseStats()), Maximum: e2eClassStatWords(*world.Players.ClassStats(class)),
				Warrior: e2eClassStatWords(world.Players.Stats.Warrior), Multipliers: e2eClassStatWords(*world.Players.ClassStatsMult(class))}
		}
	}
	// A damaged or cyclic list must not hang the game loop. Extra records are
	// reported as truncated; this observer never unlinks or repairs the list.
	sd := durations
	for scanned := 0; sd != nil && scanned < 1024; scanned++ {
		if sd.Caster16 == player && sd.Spell == uint32(spell.SPELL_CHARM) {
			if out.OwnedCharmCount < len(out.OwnedCharms) {
				out.OwnedCharms[out.OwnedCharmCount] = e2eCharmTargetObservation{
					ID: sd.ID, Level: sd.Level, Flag20: sd.Flag20, Pos: sd.Pos, Pos2: sd.Pos2,
					Frame60: sd.Frame60, Frame64: sd.Frame64, Frame68: sd.Frame68, Target: e2eReadTargetObject(sd.Target48),
				}
			} else {
				out.CharmOutputTruncated = true
			}
			out.OwnedCharmCount++
		}
		sd = sd.Next
	}
	out.DurationScanTruncated = sd != nil
	return out, nil
}

// ObservePlayerTarget runs on the game-loop thread. It reads the existing
// cursor and spell records without running target selection, changing input,
// casting, learning a guide or modifying any object or spell state.
func (sc *e2eScenario) ObservePlayerTarget(name string) {
	sc.add(0, name, func() {
		if noxServer == nil || noxClient == nil {
			e2eError(fmt.Errorf("player-target observation requires a current server and client"))
			return
		}
		out, err := e2eReadPlayerTarget(noxServer.Frame(), noxServer.Players.HostUnit(), noxServer.S().Spells.Dur.List, noxServer.S())
		if err != nil {
			e2eError(err)
			return
		}
		out.Mouse, out.ClientCursor = noxClient.Inp.GetMousePos(), uint32(noxClient.Nox_client_getCursorType())
		if dr := legacy.Nox_xxx_clientGetSpriteAtCursor_476F90(); dr != nil {
			out.ClientDrawablePresent, out.ClientDrawableNetCode, out.ClientDrawableType = true, dr.NetCode32, dr.TypeIDVal
		}
		if pl := legacy.Get_dword_8531A0_2576(); pl != nil {
			out.ClientGuidePresent, out.ClientGuideLevels = true, pl.BeastScrollLvl
		}
		data, err := json.Marshal(out)
		if err != nil {
			e2eError(fmt.Errorf("marshal player-target observation: %w", err))
			return
		}
		e2eLog.Printf("PLAYER TARGET OBSERVATION: %s", data)
	})
}
