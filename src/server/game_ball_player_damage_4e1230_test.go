package server

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type gameBallDamageWorld4E1230 struct {
	cache                               uint32
	source, target, wrong, ball, second *Object
	team                                *Team
	events                              []string
	faultAt                             int
}

func newGameBallDamageWorld4E1230() *gameBallDamageWorld4E1230 {
	w := &gameBallDamageWorld4E1230{source: &Object{TeamVal: ObjectTeam{ID: 0xab}},
		target: &Object{ObjClass: object.ClassPlayer}, team: &Team{IDVal: 0xab}}
	w.second = &Object{TypeInd: 0x2468, ObjFlags: object.Flags(0x44)}
	w.ball = &Object{TypeInd: 0x2468, ObjFlags: object.Flags(0xa5b6c7ff), NetCode: 0x12345678,
		TeamVal: ObjectTeam{ID: 2}, ObjOwner: w.target, Obj130: w.source, Field128: w.second}
	w.wrong = &Object{TypeInd: 7, Field128: w.ball}
	w.target.Field129 = w.wrong
	return w
}

func (w *gameBallDamageWorld4E1230) event(s string) {
	w.events = append(w.events, s)
	if w.faultAt == len(w.events) {
		panic(s)
	}
}

func (w *gameBallDamageWorld4E1230) hooks() gameBallPlayerDamageHooks4E1230[*Object, *ObjectTeam, *Team] {
	return gameBallPlayerDamageHooks4E1230[*Object, *ObjectTeam, *Team]{
		loadClassLow:   func(o *Object) uint8 { w.event("class"); return uint8(o.ObjClass) },
		loadTypeCache:  func() uint32 { w.event("cache"); return w.cache },
		lookupType:     func(s string) uint32 { w.event("lookup:" + s); return 0x2468 },
		storeTypeCache: func(v uint32) { w.event("store-cache"); w.cache = v },
		firstOwned:     func(o *Object) *Object { w.event("first"); return o.Field129 },
		nextOwned:      func(o *Object) *Object { w.event("next"); return o.Field128 },
		loadType:       func(o *Object) uint16 { w.event(fmt.Sprintf("type:%d", o.TypeInd)); return o.TypeInd },
		loadFlags:      func(o *Object) uint32 { w.event("flags"); return uint32(o.ObjFlags) },
		storeFlags:     func(o *Object, flags uint32) { w.event("store-flags"); o.ObjFlags = object.Flags(flags) },
		applyForce: func(victim, ball *Object, force float32) {
			w.event("force")
			if victim != w.target || ball != w.ball || force != 30 || ball.ObjFlags.Has(object.FlagNoCollide) {
				panic("bad force")
			}
		},
		clearOwner: func(o *Object) { w.event("clear-owner"); o.ObjOwner = nil; w.wrong.Field128 = o.Field128 },
		carrierState: func(ball, victim *Object) {
			w.event("carrier-state")
			if ball != w.ball || victim != w.target || ball.ObjOwner != nil || ball.Obj130 != w.source {
				panic("bad carrier state")
			}
		},
		loadTeam:   func(o *Object) *ObjectTeam { w.event("team"); return o.TeamPtr() },
		hasTeam:    func(v *ObjectTeam) bool { w.event("has-team"); return v.Has() },
		loadTeamID: func(o *Object) uint8 { w.event("source-team"); return uint8(o.TeamVal.ID) },
		findTeam: func(id uint8) *Team {
			w.event("find-team")
			if id == uint8(w.team.ID()) {
				return w.team
			}
			return nil
		},
		loadNetCode: func(o *Object) uint32 { w.event("netcode"); return o.NetCode },
		changeTeam: func(v *ObjectTeam, team *Team, code uint32, flags int32) {
			w.event("change-team")
			if v != w.ball.TeamPtr() || team != w.team || code != w.ball.NetCode || flags != 0 {
				panic("bad change-team")
			}
			v.ID = team.ID()
		},
		createTeam: func(id uint8, v *ObjectTeam, active int32, code uint32, flags int32) {
			w.event("create-team")
			if id != uint8(w.source.TeamVal.ID) || v != w.ball.TeamPtr() || active != 1 || code != w.ball.NetCode || flags != 0 {
				panic("bad create-team")
			}
			v.ID = TeamID(id)
		},
		audio: func(id uint32, o *Object, kind int32, code uint32) {
			w.event("audio")
			if id != 926 || o != w.target || kind != 0 || code != 0 {
				panic("bad audio")
			}
		},
	}
}

func gameBallDamageSuccessEvents4E1230() []string {
	return []string{"class", "cache", "lookup:GameBall", "store-cache", "first", "type:7", "next", "type:9320",
		"flags", "store-flags", "force", "clear-owner", "carrier-state", "team", "has-team", "source-team", "find-team", "netcode", "change-team", "audio"}
}

func TestGameBallPlayerDamage4E1230TraceAndFaultBoundaries(t *testing.T) {
	want := gameBallDamageSuccessEvents4E1230()
	for fault := 0; fault <= len(want); fault++ {
		t.Run(fmt.Sprint(fault), func(t *testing.T) {
			w := newGameBallDamageWorld4E1230()
			w.faultAt = fault
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				gameBallPlayerDamage4E1230(w.source, w.target, 30, w.hooks())
			}()
			n := len(want)
			if fault != 0 {
				n = fault
			}
			if !reflect.DeepEqual(w.events, want[:n]) || (recovered != nil) != (fault != 0) {
				t.Fatalf("events=%v fault=%v", w.events, recovered)
			}
			if fault == 0 && (w.ball.ObjOwner != nil || w.wrong.Field128 != w.second || uint32(w.ball.ObjFlags) != 0xa5b6c7bf || w.ball.TeamVal.ID != 0xab || w.ball.Obj130 != w.source || w.second.ObjFlags != 0x44) {
				t.Fatal("release state")
			}
		})
	}
}

func TestGameBallPlayerDamage4E1230GatesAndFullCacheWidth(t *testing.T) {
	for _, damage := range []int32{-2147483648, -1, 0, 29} {
		w := newGameBallDamageWorld4E1230()
		gameBallPlayerDamage4E1230(w.source, w.target, damage, w.hooks())
		if !reflect.DeepEqual(w.events, []string{"class"}) || w.ball.ObjOwner != w.target {
			t.Fatalf("damage %d: %v", damage, w.events)
		}
	}
	w := newGameBallDamageWorld4E1230()
	w.target.ObjClass = object.ClassMonster
	gameBallPlayerDamage4E1230(w.source, w.target, 2147483647, w.hooks())
	if !reflect.DeepEqual(w.events, []string{"class"}) {
		t.Fatal(w.events)
	}
	w = newGameBallDamageWorld4E1230()
	w.cache = 0x12468
	gameBallPlayerDamage4E1230(w.source, w.target, 30, w.hooks())
	if w.ball.ObjOwner != w.target || !reflect.DeepEqual(w.events, []string{"class", "cache", "first", "type:7", "next", "type:9320", "next", "type:9320", "next"}) {
		t.Fatal(w.events)
	}
	w = newGameBallDamageWorld4E1230()
	w.target.Field129 = nil
	gameBallPlayerDamage4E1230(w.source, w.target, 30, w.hooks())
	if w.cache != 0x2468 || !reflect.DeepEqual(w.events, gameBallDamageSuccessEvents4E1230()[:5]) {
		t.Fatal(w.events)
	}
}

func TestGameBallPlayerDamage4E1230TeamBranchesAndLiveReads(t *testing.T) {
	w := newGameBallDamageWorld4E1230()
	h := w.hooks()
	h.carrierState = func(ball, victim *Object) {
		w.event("carrier-state")
		ball.TeamVal.ID = 0
		ball.NetCode = 0xfedcba98
		w.source.TeamVal.ID = 0xfe
	}
	gameBallPlayerDamage4E1230(w.source, w.target, 31, h)
	want := append(gameBallDamageSuccessEvents4E1230()[:15], "netcode", "source-team", "create-team", "audio")
	if !reflect.DeepEqual(w.events, want) || w.ball.TeamVal.ID != 0xfe {
		t.Fatalf("teamless: %v", w.events)
	}
	w = newGameBallDamageWorld4E1230()
	w.source.TeamVal.ID = 0
	gameBallPlayerDamage4E1230(w.source, w.target, 30, w.hooks())
	want = append(gameBallDamageSuccessEvents4E1230()[:17], "audio")
	if !reflect.DeepEqual(w.events, want) || w.ball.ObjOwner != nil || w.ball.TeamVal.ID != 2 {
		t.Fatalf("unknown source team: %v", w.events)
	}
	w = newGameBallDamageWorld4E1230()
	h = w.hooks()
	h.findTeam = func(id uint8) *Team { w.event("find-team"); w.ball.NetCode = 0x87654321; return w.team }
	gameBallPlayerDamage4E1230(w.source, w.target, 30, h)
	if !reflect.DeepEqual(w.events, gameBallDamageSuccessEvents4E1230()) {
		t.Fatal(w.events)
	}
}

func TestGameBallPlayerDamage4E1230NilSourceFaultAfterRelease(t *testing.T) {
	w := newGameBallDamageWorld4E1230()
	defer func() {
		if recover() == nil || w.ball.ObjOwner != nil || !reflect.DeepEqual(w.events, gameBallDamageSuccessEvents4E1230()[:16]) {
			t.Fatalf("nil source: %v", w.events)
		}
	}()
	gameBallPlayerDamage4E1230(nil, w.target, 30, w.hooks())
}

func TestGameBallPlayerDamageNative4E1230OwnershipAndCarrierRecord(t *testing.T) {
	s := creatureMonitoredTestServer500CC0(t)
	s.SetFrame(0x89abcdef)
	update := &GameBallUpdateData4EA800{Ticks: 0x123456789abcdef, PossessionDuration: 900, ResetVelocity: 55, Reserved: 7}
	source := &Object{TeamVal: ObjectTeam{ID: 0xab}}
	target := &Object{ObjClass: object.ClassPlayer, TeamVal: ObjectTeam{ID: 2}, PosVec: types.Ptf(34, 56)}
	ball := &Object{TypeInd: 0x2468, ObjFlags: object.FlagNoCollide, ObjOwner: target, Obj130: source, NetCode: 0x9876, UpdateData: unsafe.Pointer(update), serverHandle: s.handle}
	wrong := &Object{Field128: ball}
	target.Field129 = wrong
	cache := uint32(0x2468)
	called := false
	s.GameBallOnPlayerDamage4E1230(source, target, 30, GameBallPlayerDamageRuntime4E1230{
		LoadTypeCache: func() uint32 { return cache }, StoreTypeCache: func(uint32) { t.Fatal("unexpected cache store") },
		ApplyForce: func(got *Object, pos types.Pointf, force float64) {
			if got != ball || pos != target.PosVec || force != 30 || ball.ObjOwner != target || ball.ObjFlags.Has(object.FlagNoCollide) {
				t.Fatal("force ordering")
			}
		},
		CreateTeam: func(id TeamID, value *ObjectTeam, active int32, code uint32, flags int32) {
			if id != 0xab || value != ball.TeamPtr() || active != 1 || code != 0x9876 || flags != 0 || ball.ObjOwner != nil || update.Carrier != target || update.TeamID != 2 || update.CarrierFrame != 0x89abcdef {
				t.Fatal("native team create")
			}
			called = true
		},
	})
	if !called || wrong.Field128 != nil || ball.ObjOwner != nil || ball.Obj130 != source || update.Ticks != 0x123456789abcdef || update.PossessionDuration != 900 || update.ResetVelocity != 55 || update.Reserved != 7 {
		t.Fatal("native release damaged unrelated fields")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(ball)) <= 0xffffffff || uintptr(ball.UpdateData) <= 0xffffffff) {
		t.Fatal("native fixtures require high pointers")
	}
}
