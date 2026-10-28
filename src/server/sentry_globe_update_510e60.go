package server

import (
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
	sentryGlobeInsertRaw510E60(source, st.listRuntime510E60())
}

func (st *sentryGlobeState510E60) remove(source *Object) {
	SentryGlobeUnlinkRaw510FD0(source, st.listRuntime510E60())
}

func (st *sentryGlobeState510E60) listRuntime510E60() SentryGlobeListRuntime510E60 {
	return SentryGlobeListRuntime510E60{Head: func() *Object { return st.head }, SetHead: func(o *Object) { st.head = o }}
}

// update restores GAME.EXE 00510E60 and its scan callback 00511020 without
// storing native pointers in the PE32 sentry list. InvNextItem and Field125
// retain the original next/previous roles, but are native-width Go pointers.
func (st *sentryGlobeState510E60) update(source *Object, deps sentryGlobeUpdateDeps510E60) {
	if source == nil {
		return
	}
	h := sentryGlobeDepsRuntime510E60(deps)
	h.List = st.listRuntime510E60()
	SentryGlobeUpdateRaw510E60(source, h)
}

func sentryGlobeHitCandidate511020(source, candidate *Object, start, end types.Pointf, deps sentryGlobeUpdateDeps510E60) {
	if candidate == nil {
		return
	}
	ray := types.Rectf{Min: start, Max: end}
	SentryGlobeHitRaw511020(candidate, &SentryGlobeRayContext511020{Source: source, Ray: &ray}, sentryGlobeDepsRuntime510E60(deps))
}

func sentryGlobeProject511020(ray *types.Rectf, point, out *types.Pointf) int32 {
	value, ok := PointOnTheLine(ray.Min, ray.Max, *point)
	*out = value
	if ok {
		return 1
	}
	return 0
}

func sentryGlobeDepsRuntime510E60(deps sentryGlobeUpdateDeps510E60) SentryGlobeRawRuntime510E60 {
	h := SentryGlobeRawRuntime510E60{
		TraceRay: func(ray *types.Rectf, hit *types.Pointf, _ uint32) int32 {
			value, clear := deps.traceRay(ray.Min, ray.Max)
			*hit = value
			if clear {
				return 1
			}
			return 0
		},
		GameFlags: func(uint32) uint32 {
			if deps.questMode {
				return 1
			}
			return 0
		},
		PointOnLine: sentryGlobeProject511020, FindParent: (*Object).FindOwnerChainPlayer,
		Damage: deps.damage, Audio: func(_ uint32, target *Object) { deps.audio(target) },
	}
	h.EachInRect = func(rect *types.Rectf, context *SentryGlobeRayContext511020) {
		deps.eachInRect(*rect, func(candidate *Object) bool { SentryGlobeHitRaw511020(candidate, context, h); return true })
	}
	return h
}

// SentryGlobeRuntime510E60 binds all native entries to the same server list.
func (s *Server) SentryGlobeRuntime510E60() SentryGlobeRawRuntime510E60 {
	h := SentryGlobeRawRuntime510E60{
		List: s.sentryGlobe510E60.listRuntime510E60(),
		TraceRay: func(ray *types.Rectf, hit *types.Pointf, flags uint32) int32 {
			if s.MapTraceRayAt(ray.Min, ray.Max, hit, nil, MapTraceFlags(flags)) {
				return 1
			}
			return 0
		},
		GameFlags: func(mask uint32) uint32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return mask
			}
			return 0
		},
		PointOnLine: sentryGlobeProject511020, FindParent: (*Object).FindOwnerChainPlayer,
		Damage: func(target, parent, source *Object, damage int, typ object.DamageType) {
			target.CallDamage(parent, source, damage, typ)
		},
		Audio:  func(id uint32, target *Object) { s.Audio.EventObj(sound.ID(id), target, 0, 0) },
		Player: func(index int32) *Player { return s.Players.ByInd(ntype.PlayerInd(index)) },
		Packet: func(index, kind int32, packet []byte) int32 {
			if s.NetList.AddToMsgListCli(ntype.PlayerInd(index), netlist.Kind(kind), packet) {
				return 1
			}
			return 0
		},
	}
	h.EachInRect = func(rect *types.Rectf, context *SentryGlobeRayContext511020) {
		s.Map.EachObjInRect(*rect, func(candidate *Object) bool { SentryGlobeHitRaw511020(candidate, context, h); return true })
	}
	return h
}

// SentryGlobeUpdate510E60 binds the SentryGlobe update to native-width Object
// pointers rather than re-entering GAME.EXE's PE32 callback through cgo.
func (s *Server) SentryGlobeUpdate510E60(source *Object) {
	if s == nil {
		return
	}
	if source != nil {
		SentryGlobeUpdateRaw510E60(source, s.SentryGlobeRuntime510E60())
	}
}

func sentryGlobeRound419A30(value float32) uint32 {
	return sentryGlobeRoundRaw419A30(value)
}

func sentryGlobeRayPacket511250(source *Object) [9]byte {
	return sentryGlobePacketBytes511250(source)
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
	h := s.SentryGlobeRuntime510E60()
	h.Player = func(int32) *Player { return player }
	SentryGlobeSendRaw511100(int32(index), h)
}

// SentryGlobeReset510E50 preserves the original reset boundary: map teardown
// drops the list head while object teardown owns the individual links.
func (s *Server) SentryGlobeReset510E50() {
	if s != nil {
		SentryGlobeResetRaw510E50(s.sentryGlobe510E60.listRuntime510E60())
	}
}
