package opennox

import (
	"net/http"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type debugMonsterCombat struct {
	StatusFlags uint32  `json:"status_flags"`
	CanCast     bool    `json:"can_cast_spells"`
	Aggression  float32 `json:"aggression"`
	Retreat     float32 `json:"retreat"`
	Regroup     float32 `json:"regroup"`
}

type debugObjectCombat struct {
	Ind       int    `json:"ind"`
	ScriptID  int32  `json:"script_id"`
	ID        string `json:"id"`
	Buffs     uint32 `json:"buffs"`
	AntiMagic struct {
		Active bool   `json:"active"`
		Frames uint16 `json:"frames"`
		Power  uint8  `json:"power"`
	} `json:"anti_magic"`
	Monster *debugMonsterCombat `json:"monster,omitempty"`
}

func snapshotObjectCombat(obj *server.Object) debugObjectCombat {
	out := debugObjectCombat{Ind: obj.Ind(), ScriptID: obj.ScriptIDVal, ID: obj.ID(), Buffs: obj.Buffs}
	out.AntiMagic.Active = obj.HasEnchant(server.ENCHANT_ANTI_MAGIC)
	out.AntiMagic.Frames = obj.BuffsDur[server.ENCHANT_ANTI_MAGIC]
	out.AntiMagic.Power = obj.BuffsPower[server.ENCHANT_ANTI_MAGIC]
	if obj.Class().Has(object.ClassMonster) && obj.UpdateData != nil {
		status := obj.UpdateDataMonster().StatusFlags
		out.Monster = &debugMonsterCombat{
			StatusFlags: uint32(status), CanCast: status.Has(object.MonStatusCanCastSpells),
			Aggression: obj.Aggression(), Retreat: obj.RetreatLevel(), Regroup: obj.RegroupLevel(),
		}
	}
	return out
}

func init() {
	// This observer copies native status on the server loop. Unlike the
	// indexed object route, no query parameter can modify object flags.
	http.HandleFunc("/debug/nox/objects/combat", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
			out := struct {
				Frame   uint32              `json:"frame"`
				Objects []debugObjectCombat `json:"objects"`
			}{Frame: noxServer.Frame(), Objects: make([]debugObjectCombat, 0)}
			for _, obj := range noxServer.Objs.AllObjects() {
				out.Objects = append(out.Objects, snapshotObjectCombat(obj))
			}
			return out
		})
	})
}
