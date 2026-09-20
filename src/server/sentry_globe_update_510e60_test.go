package server

import (
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/common"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

func sentryGlobeTestObject510E60(data *SentryUpdateData) *Object {
	return &Object{UpdateData: unsafe.Pointer(data)}
}

func TestSentryGlobeUpdate510E60TracesDamagesAndLinksNativePointers(t *testing.T) {
	data := &SentryUpdateData{Field8: math.Float32bits(0.25)}
	owner := &Object{ObjClass: object.ClassPlayer}
	source := sentryGlobeTestObject510E60(data)
	source.ObjFlags = object.FlagEnabled
	source.ObjOwner = owner
	source.PosVec = types.Ptf(100, 200)

	valid := &Object{PosVec: types.Ptf(125, 202), HealthData: new(HealthData)}
	valid.Shape.Circle.R = 3
	blocked := &Object{PosVec: types.Ptf(125, 202), HealthData: new(HealthData), ObjFlags: object.FlagNoCollide}
	blocked.Shape.Circle.R = 3
	shortOutsideQuest := &Object{PosVec: types.Ptf(125, 202), HealthData: new(HealthData), ObjFlags: object.FlagShort, ObjClass: object.ClassMonster}
	shortOutsideQuest.Shape.Circle.R = 3

	var state sentryGlobeState510E60
	var gotRect types.Rectf
	var damaged, sounded []*Object
	state.update(source, sentryGlobeUpdateDeps510E60{
		traceRay: func(start, end types.Pointf) (types.Pointf, bool) {
			if start != source.PosVec || end != (types.Ptf(700, 200)) {
				t.Fatalf("trace = %v -> %v, want %v -> {700 200}", start, end, source.PosVec)
			}
			return types.Ptf(150, 200), false
		},
		eachInRect: func(rect types.Rectf, fn func(*Object) bool) {
			gotRect = rect
			for _, candidate := range []*Object{valid, blocked, shortOutsideQuest} {
				if !fn(candidate) {
					t.Fatal("sentry scan stopped early")
				}
			}
		},
		damage: func(target, parent, gotSource *Object, damage int, damageType object.DamageType) {
			if parent != owner || gotSource != source || damage != 500 || damageType != object.DamageZapRay {
				t.Fatalf("damage args = %p/%p/%d/%d", parent, gotSource, damage, damageType)
			}
			damaged = append(damaged, target)
		},
		audio: func(target *Object) { sounded = append(sounded, target) },
	})

	if state.head != source || !source.ObjFlags.Has(object.FlagMarked) {
		t.Fatalf("list head/marker = %p/%v, want source/marked", state.head, source.ObjFlags)
	}
	if source.Pos39 != (types.Ptf(150, 200)) {
		t.Fatalf("beam end = %v, want {150 200}", source.Pos39)
	}
	if got := math.Float32frombits(data.Field0); got != 0.25 {
		t.Fatalf("angle = %v, want 0.25", got)
	}
	wantRect := types.RectFromPointsf(types.Ptf(100, 200), types.Ptf(150, 200))
	if gotRect != wantRect {
		t.Fatalf("scan rect = %v, want %v", gotRect, wantRect)
	}
	if len(damaged) != 1 || damaged[0] != valid || len(sounded) != 1 || sounded[0] != valid {
		t.Fatalf("hits = damage %p sound %p, want valid only", damaged, sounded)
	}
}

func TestSentryGlobeUpdate510E60DestroyedRemovesAndDisabledResetsAngle(t *testing.T) {
	data := &SentryUpdateData{Field0: math.Float32bits(1), Field4: math.Float32bits(2.5)}
	source := sentryGlobeTestObject510E60(data)
	var state sentryGlobeState510E60
	state.update(source, sentryGlobeUpdateDeps510E60{})
	if state.head != source || !source.ObjFlags.Has(object.FlagMarked) {
		t.Fatal("initial disabled update did not link the sentry")
	}
	if data.Field0 != data.Field4 {
		t.Fatal("disabled update did not reset the angle")
	}

	source.ObjFlags |= object.FlagDestroyed
	state.update(source, sentryGlobeUpdateDeps510E60{})
	if state.head != nil || source.ObjFlags.Has(object.FlagMarked) {
		t.Fatalf("destroyed sentry remained linked: head=%p flags=%v", state.head, source.ObjFlags)
	}
}

func TestSentryGlobeSendToPlayer511100PacketAndDisabledReset(t *testing.T) {
	visibleData := new(SentryUpdateData)
	visible := sentryGlobeTestObject510E60(visibleData)
	visible.ObjFlags = object.FlagEnabled | object.FlagMarked
	visible.PosVec = types.Ptf(90.4, 95.6)
	visible.Pos39 = types.Ptf(110.6, 105.4)
	disabledData := &SentryUpdateData{Field0: math.Float32bits(1), Field4: math.Float32bits(3)}
	disabled := sentryGlobeTestObject510E60(disabledData)
	disabled.ObjFlags = object.FlagMarked
	visible.InvNextItem = disabled
	disabled.Field125 = visible

	nl := netlist.New()
	nl.Init()
	s := &Server{NetList: nl}
	s.Players.list = make([]Player, common.MaxPlayers)
	const index = ntype.PlayerInd(2)
	player := &s.Players.list[index]
	player.Active = 1
	player.PlayerInd = byte(index)
	player.Field10 = 20
	player.Field12 = 20
	player.Pos3632Vec = types.Ptf(100, 100)
	s.sentryGlobe510E60.head = visible

	s.SentryGlobeSendToPlayer511100(int(index))
	packet := s.NetList.CopyPacketsA(index, netlist.Kind1)
	if len(packet) != 9 || packet[0] != byte(netmsg.MSG_FX_SENTRY_RAY) {
		t.Fatalf("packet = %v, want 9-byte sentry ray", packet)
	}
	for offset, want := range map[int]uint16{1: 90, 3: 96, 5: 111, 7: 105} {
		if got := binary.LittleEndian.Uint16(packet[offset : offset+2]); got != want {
			t.Fatalf("packet coordinate at %d = %d, want %d", offset, got, want)
		}
	}
	if disabledData.Field0 != disabledData.Field4 {
		t.Fatal("network walk did not reset a disabled sentry angle")
	}
}

func TestSentryGlobeRound419A30NegativeIsZero(t *testing.T) {
	if got := sentryGlobeRound419A30(-0.25); got != 0 {
		t.Fatalf("round(-0.25) = %d, want 0", got)
	}
}
