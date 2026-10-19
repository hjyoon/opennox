package opennox

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// Observations copy values only. No world object or update-data pointer is
// retained after a game-loop observation, including in the linked-shaft row.
type e2eElevatorActorObservation struct {
	Extent    uint32       `json:"extent"`
	ScriptID  int32        `json:"script_id"`
	Class     object.Class `json:"class"`
	Flags     object.Flags `json:"flags"`
	Pos       types.Pointf `json:"pos"`
	NewPos    types.Pointf `json:"new_pos"`
	Velocity  types.Pointf `json:"velocity"`
	Force     types.Pointf `json:"force"`
	Z         float32      `json:"z"`
	ZVelocity float32      `json:"z_velocity"`
	Shape     server.Shape `json:"shape"`
}

type e2eElevatorBoardingObservation struct {
	Frame               uint32                      `json:"frame"`
	Player              e2eElevatorActorObservation `json:"player"`
	PlayerState         uint32                      `json:"player_state"`
	PlayerHealth        uint16                      `json:"player_health"`
	PlayerMaximumHealth uint16                      `json:"player_maximum_health"`
	Elevator            e2eElevatorActorObservation `json:"elevator"`
	State               byte                        `json:"state"`
	Height              uint32                      `json:"height"`
	ShaftExtent         uint32                      `json:"shaft_extent"`
	LinkPresent         bool                        `json:"link_present"`
	Shaft               e2eElevatorActorObservation `json:"shaft"`
	LinkReciprocal      bool                        `json:"link_reciprocal"`
	PlayerInBox         bool                        `json:"player_in_box"`
	PlayerNewPosInBox   bool                        `json:"player_new_pos_in_box"`
	HeightDifference    float64                     `json:"height_difference"`
}

func e2eReadElevatorActor(obj *server.Object) e2eElevatorActorObservation {
	return e2eElevatorActorObservation{
		Extent: obj.Extent, ScriptID: obj.ScriptIDVal, Class: obj.Class(), Flags: obj.Flags(),
		Pos: obj.PosVec, NewPos: obj.NewPos, Velocity: obj.VelVec, Force: obj.ForceVec,
		Z: obj.ZVal, ZVelocity: obj.Field27, Shape: obj.Shape,
	}
}

func e2eReadElevatorBoarding(frame uint32, player, elevator *server.Object) (e2eElevatorBoardingObservation, error) {
	if player == nil || !player.Class().Has(object.ClassPlayer) {
		return e2eElevatorBoardingObservation{}, fmt.Errorf("missing original player for elevator observation")
	}
	if elevator == nil || !elevator.Class().Has(object.ClassElevator) || elevator.UpdateData == nil {
		return e2eElevatorBoardingObservation{}, fmt.Errorf("missing original Elevator update record")
	}
	data := elevator.UpdateDataElevator()
	out := e2eElevatorBoardingObservation{
		Frame: frame, Player: e2eReadElevatorActor(player), Elevator: e2eReadElevatorActor(elevator),
		State: data.Field_3, Height: data.Field_4, ShaftExtent: data.Field_2,
		PlayerInBox:       server.MapPointInBox57B850(&elevator.PosVec, &elevator.Shape, &player.PosVec),
		PlayerNewPosInBox: server.MapPointInBox57B850(&elevator.NewPos, &elevator.Shape, &player.NewPos),
		HeightDifference:  math.Abs(float64(player.ZVal) - float64(int32(data.Field_4))),
	}
	if player.HealthData != nil {
		out.PlayerHealth, out.PlayerMaximumHealth = player.HealthData.Cur, player.HealthData.Max
	}
	if player.UpdateData != nil {
		out.PlayerState = uint32(player.UpdateDataPlayer().State)
	}
	if shaft := elevator.ElevatorLinkFor(elevator.UpdateData); shaft != nil {
		out.LinkPresent = true
		out.Shaft = e2eReadElevatorActor(shaft)
		if shaft.Class().Has(object.ClassElevatorShaft) && shaft.UpdateData != nil {
			out.LinkReciprocal = shaft.ElevatorLinkFor(shaft.UpdateData) == elevator
		}
	}
	return out, nil
}

// ObserveElevatorBoarding reads on the game-loop thread. It neither boards a
// passenger nor changes input, position, height, flags, links or elevator state.
func (sc *e2eScenario) ObserveElevatorBoarding(extent uint32, name string) {
	sc.add(0, name, func() {
		if noxServer == nil || extent == 0 {
			e2eError(fmt.Errorf("elevator observation requires a current server and nonzero original extent"))
			return
		}
		var elevator *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Extent != extent {
				continue
			}
			if elevator != nil {
				e2eError(fmt.Errorf("ambiguous original elevator extent %d", extent))
				return
			}
			elevator = obj
		}
		observation, err := e2eReadElevatorBoarding(noxServer.Frame(), noxServer.Players.HostUnit(), elevator)
		if err != nil {
			e2eError(err)
			return
		}
		data, err := json.Marshal(observation)
		if err != nil {
			e2eError(fmt.Errorf("marshal original elevator observation: %w", err))
			return
		}
		e2eLog.Printf("ELEVATOR BOARDING OBSERVATION: %s", data)
	})
}
