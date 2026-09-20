package server

import (
	"encoding/binary"
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
)

const sentryGlobeRange510E60 = float32(600)

type sentryGlobeState510E60 struct {
	head *Object
}

type sentryGlobeUpdateDeps510E60 struct {
	traceRay   func(types.Pointf, types.Pointf) (types.Pointf, bool)
	eachInRect func(types.Rectf, func(*Object) bool)
	questMode  bool
	damage     func(target, parent, source *Object, damage int, damageType object.DamageType)
	audio      func(*Object)
}

func (st *sentryGlobeState510E60) insert(source *Object) {
	source.Field125 = nil
	source.InvNextItem = st.head
	if st.head != nil {
		st.head.Field125 = source
	}
	st.head = source
	source.ObjFlags |= object.FlagMarked
}

func (st *sentryGlobeState510E60) remove(source *Object) {
	if source.ObjFlags.Has(object.FlagMarked) {
		prev := source.Field125
		next := source.InvNextItem
		if prev != nil {
			prev.InvNextItem = next
		} else {
			st.head = next
		}
		if next != nil {
			next.Field125 = prev
		}
	}
	source.ObjFlags &^= object.FlagMarked
}

// update restores GAME.EXE 00510E60 and its scan callback 00511020 without
// storing native pointers in the PE32 sentry list. InvNextItem and Field125
// retain the original next/previous roles, but are native-width Go pointers.
func (st *sentryGlobeState510E60) update(source *Object, deps sentryGlobeUpdateDeps510E60) {
	if source == nil {
		return
	}
	data := source.UpdateDataSentry()
	if !source.ObjFlags.Has(object.FlagMarked) {
		st.insert(source)
	}
	if source.ObjFlags.Has(object.FlagDestroyed) {
		st.remove(source)
	}
	if !source.ObjFlags.Has(object.FlagEnabled) {
		data.Field0 = data.Field4
		return
	}

	start := source.PosVec
	angle := math.Float32frombits(data.Field0)
	end := types.Ptf(
		float32(math.Cos(float64(angle))*float64(sentryGlobeRange510E60)+float64(start.X)),
		float32(math.Sin(float64(angle))*float64(sentryGlobeRange510E60)+float64(start.Y)),
	)
	if hit, clear := deps.traceRay(start, end); !clear {
		end = hit
	}
	source.Pos39 = end
	data.Field0 = math.Float32bits(angle + math.Float32frombits(data.Field8))

	rect := types.RectFromPointsf(start, end)
	deps.eachInRect(rect, func(candidate *Object) bool {
		sentryGlobeHitCandidate511020(source, candidate, start, end, deps)
		return true
	})
}

func sentryGlobeHitCandidate511020(source, candidate *Object, start, end types.Pointf, deps sentryGlobeUpdateDeps510E60) {
	if candidate == nil || candidate.ObjFlags.HasAny(object.FlagBelow|object.FlagNoCollide) {
		return
	}
	if candidate.ObjFlags.Has(object.FlagShort) &&
		(!deps.questMode || !candidate.ObjClass.Has(object.ClassMonster) || candidate.ObjFlags.Has(object.FlagDead)) {
		return
	}
	if candidate.HealthData == nil {
		return
	}
	point, ok := PointOnTheLine(start, end, candidate.PosVec)
	if !ok {
		return
	}
	delta := candidate.PosVec.Sub(point)
	distanceSquared := float64(delta.X)*float64(delta.X) + float64(delta.Y)*float64(delta.Y)
	radius := float64(candidate.Shape.Circle.R)
	if radius*radius <= distanceSquared {
		return
	}
	parent := source.FindOwnerChainPlayer()
	deps.damage(candidate, parent, source, 500, object.DamageZapRay)
	deps.audio(candidate)
}

// SentryGlobeUpdate510E60 binds the SentryGlobe update to native-width Object
// pointers rather than re-entering GAME.EXE's PE32 callback through cgo.
func (s *Server) SentryGlobeUpdate510E60(source *Object) {
	if s == nil {
		return
	}
	s.sentryGlobe510E60.update(source, sentryGlobeUpdateDeps510E60{
		traceRay: func(start, end types.Pointf) (types.Pointf, bool) {
			var hit types.Pointf
			clear := s.MapTraceRayAt(start, end, &hit, nil, 5)
			return hit, clear
		},
		eachInRect: s.Map.EachObjInRect,
		questMode:  noxflags.HasGame(noxflags.GameModeQuest),
		damage: func(target, parent, source *Object, damage int, damageType object.DamageType) {
			target.CallDamage(parent, source, damage, damageType)
		},
		audio: func(target *Object) {
			s.Audio.EventObj(sound.SoundSentryRayHit, target, 0, 0)
		},
	})
}

func sentryGlobeRound419A30(value float32) uint32 {
	if value < 0 {
		return 0
	}
	value += float32(8388608)
	return math.Float32bits(value) & 0x7fffff
}

func sentryGlobeRayPacket511250(source *Object) [9]byte {
	var packet [9]byte
	packet[0] = byte(netmsg.MSG_FX_SENTRY_RAY)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(sentryGlobeRound419A30(source.PosVec.X)))
	binary.LittleEndian.PutUint16(packet[3:5], uint16(sentryGlobeRound419A30(source.PosVec.Y)))
	binary.LittleEndian.PutUint16(packet[5:7], uint16(sentryGlobeRound419A30(source.Pos39.X)))
	binary.LittleEndian.PutUint16(packet[7:9], uint16(sentryGlobeRound419A30(source.Pos39.Y)))
	return packet
}

// SentryGlobeSendToPlayer511100 restores GAME.EXE 00511100/00511250 using
// the same native-width list owned by SentryGlobeUpdate510E60.
func (s *Server) SentryGlobeSendToPlayer511100(index int) {
	if s == nil {
		return
	}
	player := s.Players.ByInd(ntype.PlayerInd(index))
	if player == nil {
		return
	}
	halfWidth := float32(player.Field10)
	halfHeight := float32(player.Field12)
	viewport := types.Rectf{
		Min: types.Ptf(player.Pos3632Vec.X-halfWidth, player.Pos3632Vec.Y-halfHeight),
		Max: types.Ptf(player.Pos3632Vec.X+halfWidth, player.Pos3632Vec.Y+halfHeight),
	}
	for source := s.sentryGlobe510E60.head; source != nil; source = source.InvNextItem {
		if !source.ObjFlags.Has(object.FlagEnabled) {
			data := source.UpdateDataSentry()
			data.Field0 = data.Field4
			continue
		}
		beam := types.RectFromPointsf(source.PosVec, source.Pos39)
		if viewport.Min.X < beam.Max.X && viewport.Max.X > beam.Min.X &&
			viewport.Min.Y < beam.Max.Y && viewport.Max.Y > beam.Min.Y {
			packet := sentryGlobeRayPacket511250(source)
			s.NetList.AddToMsgListCli(ntype.PlayerInd(index), netlist.Kind1, packet[:])
		}
	}
}

// SentryGlobeReset510E50 preserves the original reset boundary: map teardown
// drops the list head while object teardown owns the individual links.
func (s *Server) SentryGlobeReset510E50() {
	if s != nil {
		s.sentryGlobe510E60.head = nil
	}
}
