package opennox

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eObjectiveGameMode(mode string) (noxflags.GameFlag, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "ctf":
		return noxflags.GameModeCTF, nil
	case "flagball":
		return noxflags.GameModeFlagBall, nil
	case "kotr":
		return noxflags.GameModeKOTR, nil
	default:
		return 0, fmt.Errorf("unknown E2E objective game mode %q", mode)
	}
}

// SetObjectiveGameMode changes the same mode preference used by server
// options. The subsequent ordinary map switch still performs map validation,
// objective setup, player respawn, and client map loading.
func (sc *e2eScenario) SetObjectiveGameMode(mode, name string) {
	sc.add(0, name, func() {
		flag, err := e2eObjectiveGameMode(mode)
		if err != nil {
			e2eError(err)
			return
		}
		settings := getCurrentSettings2()
		settings.Field52 = (settings.Field52 &^ uint16(noxflags.GameModeMask)) | uint16(flag)
		e2eLog.Printf("OBJECTIVE MODE PREFERENCE: mode=%s flags=%#x", flag, settings.Field52)
	})
}

func (sc *e2eScenario) AssertObjectiveGameMode(mode, name string) {
	sc.add(0, name, func() {
		flag, err := e2eObjectiveGameMode(mode)
		if err != nil {
			e2eError(err)
			return
		}
		if got := noxflags.GetGame() & noxflags.GameModeMask; got != flag {
			e2eError(fmt.Errorf("objective mode = %s, want %s", got, flag))
			return
		}
		player := noxServer.Players.Host()
		if player == nil || player.PlayerUnit == nil {
			e2eError(fmt.Errorf("objective mode has no live host player"))
			return
		}
		var objectiveType uint16
		switch flag {
		case noxflags.GameModeFlagBall:
			objectiveType = uint16(noxServer.Types.IndByID("GameBall"))
		case noxflags.GameModeKOTR:
			objectiveType = uint16(noxServer.Types.IndByID("Crown"))
		}
		count := 0
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Flags().HasAny(object.FlagDead | object.FlagDestroyed) {
				continue
			}
			if flag == noxflags.GameModeCTF {
				if !obj.Class().Has(object.ClassFlag) {
					continue
				}
			} else if obj.TypeInd != objectiveType {
				continue
			}
			count++
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(obj.CObj()) <= math.MaxUint32 {
				e2eError(fmt.Errorf("objective pointer %p is not above 4 GiB", obj))
				return
			}
			if flag != noxflags.GameModeCTF && !player.MinimapTracks(obj) {
				e2eError(fmt.Errorf("host minimap does not track %s objective %p", flag, obj))
				return
			}
			if flag == noxflags.GameModeFlagBall {
				if obj.UpdateData == nil {
					e2eError(fmt.Errorf("FlagBall has no native update data"))
					return
				}
				update := (*server.GameBallUpdateData4EA800)(obj.UpdateData)
				if update.PossessionDuration == 0 || update.ResetVelocity <= 0 {
					e2eError(fmt.Errorf("FlagBall setup parameters are missing: duration=%d velocity=%g", update.PossessionDuration, update.ResetVelocity))
					return
				}
			}
			e2eLog.Printf("OBJECTIVE OBJECT: mode=%s obj=%p update=%p team=%d minimap=%t pos=(%.3f,%.3f)",
				flag, obj, obj.UpdateData, obj.TeamVal.ID, player.MinimapTracks(obj), obj.PosVec.X, obj.PosVec.Y)
		}
		if count == 0 || (flag == noxflags.GameModeCTF && count < 2) || (flag == noxflags.GameModeFlagBall && count != 1) {
			e2eError(fmt.Errorf("objective mode %s has unexpected objective count %d", flag, count))
			return
		}
		e2eLog.Printf("OBJECTIVE MODE READY: mode=%s map=%q frame=%d objectives=%d player=%p",
			flag, legacy.Nox_xxx_mapGetMapName_409B40(), noxServer.Frame(), count, player.PlayerUnit)
	})
}

func (sc *e2eScenario) ResetFlagball(name string) {
	sc.add(0, name, func() {
		if !noxflags.HasGame(noxflags.GameModeFlagBall) {
			e2eError(fmt.Errorf("cannot reset FlagBall outside FlagBall mode"))
			return
		}
		ballType := uint16(noxServer.Types.IndByID("GameBall"))
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.TypeInd != ballType || obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				continue
			}
			if noxServer.GameBallReset417F50(obj) != 1 {
				e2eError(fmt.Errorf("FlagBall reset failed for %p", obj))
				return
			}
			e2eLog.Printf("FLAGBALL RESET: old=%p frame=%d", obj, noxServer.Frame())
			return
		}
		e2eError(fmt.Errorf("no live FlagBall to reset"))
	})
}
