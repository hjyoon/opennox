package server

import (
	"fmt"
	"reflect"
	"testing"
)

type ballCollideTestData4EBA00 struct {
	name    string
	carrier *ballCollideTestObject4EBA00
}

type ballCollideTestTeam4EBA00 struct {
	name    string
	members int32
	kind    uint8
}

type ballCollideTestObject4EBA00 struct {
	name       string
	data       *ballCollideTestData4EBA00
	teamID     uint8
	owner      *ballCollideTestObject4EBA00
	class      uint8
	ownedFirst *ballCollideTestObject4EBA00
	ownedNext  *ballCollideTestObject4EBA00
	typeInd    uint16
	netCode    uint32
	flags      uint32
	hasTeam    bool
}

type ballCollideTestFixture4EBA00 struct {
	events        []string
	teams         map[uint8]*ballCollideTestTeam4EBA00
	frames        []uint32
	frameReads    int
	feedback      uint32
	carrierResult *ballCollideTestObject4EBA00
	onMessage     func()
	onSetOwner    func()
	onHasTeam     func()
	onChangeTeam  func()
	onCreateTeam  func()
	onCarrier     func()
	onAudio       func(uint32)
}

func ballCollideObjectName4EBA00(obj *ballCollideTestObject4EBA00) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func ballCollideDataName4EBA00(data *ballCollideTestData4EBA00) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func ballCollideTeamName4EBA00(team *ballCollideTestTeam4EBA00) string {
	if team == nil {
		return "nil"
	}
	return team.name
}

func (f *ballCollideTestFixture4EBA00) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *ballCollideTestFixture4EBA00) hooks() ballCollideHooks4EBA00[
	*ballCollideTestObject4EBA00,
	*ballCollideTestTeam4EBA00,
	*ballCollideTestData4EBA00,
] {
	return ballCollideHooks4EBA00[
		*ballCollideTestObject4EBA00,
		*ballCollideTestTeam4EBA00,
		*ballCollideTestData4EBA00,
	]{
		loadUpdateData: func(obj *ballCollideTestObject4EBA00) *ballCollideTestData4EBA00 {
			if obj == nil {
				f.event("update:nil")
				panic("nil ball UpdateData")
			}
			f.event("update:%s=%s", obj.name, ballCollideDataName4EBA00(obj.data))
			return obj.data
		},
		loadTeamID: func(obj *ballCollideTestObject4EBA00) uint8 {
			f.event("team-id:%s=%#x", ballCollideObjectName4EBA00(obj), obj.teamID)
			return obj.teamID
		},
		findTeamByID: func(id uint8) *ballCollideTestTeam4EBA00 {
			team := f.teams[id]
			f.event("find-team:%#x=%s", id, ballCollideTeamName4EBA00(team))
			return team
		},
		loadCarrier: func(data *ballCollideTestData4EBA00) *ballCollideTestObject4EBA00 {
			if data == nil {
				f.event("carrier:nil")
				panic("nil Ball update-data carrier")
			}
			f.event("carrier:%s=%s", data.name, ballCollideObjectName4EBA00(data.carrier))
			return data.carrier
		},
		teamMemberCount: func(team *ballCollideTestTeam4EBA00) int32 {
			f.event("members:%s=%d", ballCollideTeamName4EBA00(team), team.members)
			return team.members
		},
		loadFrame: func() uint32 {
			if f.frameReads >= len(f.frames) {
				panic("unexpected frame read")
			}
			frame := f.frames[f.frameReads]
			f.frameReads++
			f.event("frame:%d=%#x", f.frameReads, frame)
			return frame
		},
		loadFeedbackFrame: func() uint32 {
			f.event("feedback=%#x", f.feedback)
			return f.feedback
		},
		priorityMessage: func(obj *ballCollideTestObject4EBA00, message string, flags int32) {
			f.event("message:%s:%s:%d", ballCollideObjectName4EBA00(obj), message, flags)
			if f.onMessage != nil {
				f.onMessage()
			}
		},
		storeFeedback: func(frame uint32) {
			f.event("store-feedback=%#x", frame)
			f.feedback = frame
		},
		audio: func(id uint32, obj *ballCollideTestObject4EBA00) {
			f.event("audio:%d:%s", id, ballCollideObjectName4EBA00(obj))
			if f.onAudio != nil {
				f.onAudio(id)
			}
		},
		loadOwner: func(obj *ballCollideTestObject4EBA00) *ballCollideTestObject4EBA00 {
			f.event("owner:%s=%s", ballCollideObjectName4EBA00(obj), ballCollideObjectName4EBA00(obj.owner))
			return obj.owner
		},
		loadClassLow: func(obj *ballCollideTestObject4EBA00) uint8 {
			f.event("class:%s=%#x", ballCollideObjectName4EBA00(obj), obj.class)
			return obj.class
		},
		loadOwnedFirst: func(obj *ballCollideTestObject4EBA00) *ballCollideTestObject4EBA00 {
			f.event("owned-first:%s=%s", ballCollideObjectName4EBA00(obj), ballCollideObjectName4EBA00(obj.ownedFirst))
			return obj.ownedFirst
		},
		loadTypeInd: func(obj *ballCollideTestObject4EBA00) uint16 {
			f.event("type:%s=%#x", ballCollideObjectName4EBA00(obj), obj.typeInd)
			return obj.typeInd
		},
		loadOwnedNext: func(obj *ballCollideTestObject4EBA00) *ballCollideTestObject4EBA00 {
			f.event("owned-next:%s=%s", ballCollideObjectName4EBA00(obj), ballCollideObjectName4EBA00(obj.ownedNext))
			return obj.ownedNext
		},
		setOwner: func(owner, ball *ballCollideTestObject4EBA00) {
			f.event("set-owner:%s->%s", ballCollideObjectName4EBA00(owner), ballCollideObjectName4EBA00(ball))
			ball.owner = owner
			if f.onSetOwner != nil {
				f.onSetOwner()
			}
		},
		hasTeam: func(obj *ballCollideTestObject4EBA00) int32 {
			value := obj.hasTeam
			f.event("has-team:%s=%t", ballCollideObjectName4EBA00(obj), value)
			if f.onHasTeam != nil {
				f.onHasTeam()
			}
			if value {
				return 1
			}
			return 0
		},
		loadNetCode: func(obj *ballCollideTestObject4EBA00) uint32 {
			f.event("net-code:%s=%#x", ballCollideObjectName4EBA00(obj), obj.netCode)
			return obj.netCode
		},
		changeTeam: func(obj *ballCollideTestObject4EBA00, team *ballCollideTestTeam4EBA00, netCode uint32, flags int32) {
			f.event("change-team:%s:%s:%#x:%d", ballCollideObjectName4EBA00(obj), ballCollideTeamName4EBA00(team), netCode, flags)
			if f.onChangeTeam != nil {
				f.onChangeTeam()
			}
		},
		createTeam: func(id uint8, obj *ballCollideTestObject4EBA00, active int32, netCode uint32, flags int32) {
			f.event("create-team:%#x:%s:%d:%#x:%d", id, ballCollideObjectName4EBA00(obj), active, netCode, flags)
			if f.onCreateTeam != nil {
				f.onCreateTeam()
			}
		},
		loadTeamKind: func(team *ballCollideTestTeam4EBA00) uint8 {
			f.event("team-kind:%s=%#x", ballCollideTeamName4EBA00(team), team.kind)
			return team.kind
		},
		loadNetCode16: func(obj *ballCollideTestObject4EBA00) uint16 {
			value := uint16(obj.netCode)
			f.event("net-code16:%s=%#x", ballCollideObjectName4EBA00(obj), value)
			return value
		},
		ballStatus: func(state uint8, netCode uint16) {
			f.event("status:%#x:%#x", state, netCode)
		},
		carrierState: func(ball, target *ballCollideTestObject4EBA00) *ballCollideTestObject4EBA00 {
			f.event("carrier-state:%s:%s", ballCollideObjectName4EBA00(ball), ballCollideObjectName4EBA00(target))
			if ball.data != nil {
				ball.data.carrier = target
			}
			if f.onCarrier != nil {
				f.onCarrier()
			}
			return f.carrierResult
		},
		loadFlags: func(obj *ballCollideTestObject4EBA00) uint32 {
			f.event("flags:%s=%#x", ballCollideObjectName4EBA00(obj), obj.flags)
			return obj.flags
		},
		storeFlags: func(obj *ballCollideTestObject4EBA00, flags uint32) {
			f.event("store-flags:%s=%#x", ballCollideObjectName4EBA00(obj), flags)
			obj.flags = flags
		},
		purgeBuffs: func(obj *ballCollideTestObject4EBA00) {
			f.event("purge:%s", ballCollideObjectName4EBA00(obj))
		},
	}
}

func assertBallCollideEvents4EBA00(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event order mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestBallCollide4EBA00NilTargetCachesUpdateWithoutDereference(t *testing.T) {
	ball := &ballCollideTestObject4EBA00{name: "ball"}
	f := &ballCollideTestFixture4EBA00{}
	ballCollide4EBA00(ball, nil, struct{ untouched uint32 }{0xdeadbeef}, f.hooks())
	assertBallCollideEvents4EBA00(t, f.events, []string{
		"update:ball=nil",
		"owner:ball=nil",
		"audio:928:ball",
	})
}

func TestBallCollide4EBA00TargetLookupPrecedesOwnerAndClassGates(t *testing.T) {
	oldCarrier := &ballCollideTestObject4EBA00{name: "old"}
	owner := &ballCollideTestObject4EBA00{name: "owner"}
	for _, tc := range []struct {
		name        string
		ballOwner   *ballCollideTestObject4EBA00
		targetClass uint8
		want        []string
	}{
		{
			name:      "owned ball",
			ballOwner: owner,
			want: []string{
				"update:ball=data", "team-id:target=0x7", "find-team:0x7=nil",
				"carrier:data=old", "owner:ball=owner", "audio:928:ball",
			},
		},
		{
			name:        "non-player target",
			targetClass: 0x80,
			want: []string{
				"update:ball=data", "team-id:target=0x7", "find-team:0x7=nil",
				"carrier:data=old", "owner:ball=nil", "class:target=0x80", "audio:928:ball",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ball := &ballCollideTestObject4EBA00{
				name: "ball", owner: tc.ballOwner,
				data: &ballCollideTestData4EBA00{name: "data", carrier: oldCarrier},
			}
			target := &ballCollideTestObject4EBA00{name: "target", teamID: 7, class: tc.targetClass}
			f := &ballCollideTestFixture4EBA00{teams: map[uint8]*ballCollideTestTeam4EBA00{}}
			ballCollide4EBA00(ball, target, nil, f.hooks())
			assertBallCollideEvents4EBA00(t, f.events, tc.want)
		})
	}
}

func TestBallCollide4EBA00RepeatedCarrierFeedbackReloadsFrame(t *testing.T) {
	team := &ballCollideTestTeam4EBA00{name: "red", members: 2}
	target := &ballCollideTestObject4EBA00{name: "target", teamID: 1}
	ball := &ballCollideTestObject4EBA00{
		name: "ball", data: &ballCollideTestData4EBA00{name: "data", carrier: target},
	}
	f := &ballCollideTestFixture4EBA00{
		teams:    map[uint8]*ballCollideTestTeam4EBA00{1: team},
		frames:   []uint32{146, 222},
		feedback: 100,
	}
	ballCollide4EBA00(ball, target, nil, f.hooks())
	assertBallCollideEvents4EBA00(t, f.events, []string{
		"update:ball=data", "team-id:target=0x1", "find-team:0x1=red",
		"carrier:data=target", "members:red=2", "frame:1=0x92", "feedback=0x64",
		"message:target:objcoll.c:CantPickupBall:0", "frame:2=0xde",
		"store-feedback=0xde", "audio:928:ball",
	})
	if f.feedback != 222 {
		t.Fatalf("feedback frame = %d, want reloaded 222", f.feedback)
	}
}

func TestBallCollide4EBA00FeedbackUsesUnsignedStrictBoundary(t *testing.T) {
	for _, tc := range []struct {
		name        string
		frame       uint32
		last        uint32
		wantMessage bool
	}{
		{name: "exactly 45", frame: 145, last: 100},
		{name: "wrapping 48", frame: 0x10, last: 0xffffffe0, wantMessage: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			team := &ballCollideTestTeam4EBA00{name: "team", members: 2}
			target := &ballCollideTestObject4EBA00{name: "target", teamID: 1}
			ball := &ballCollideTestObject4EBA00{
				name: "ball", data: &ballCollideTestData4EBA00{name: "data", carrier: target},
			}
			frames := []uint32{tc.frame}
			if tc.wantMessage {
				frames = append(frames, tc.frame+9)
			}
			message := false
			f := &ballCollideTestFixture4EBA00{
				teams: map[uint8]*ballCollideTestTeam4EBA00{1: team}, frames: frames, feedback: tc.last,
				onMessage: func() { message = true },
			}
			ballCollide4EBA00(ball, target, nil, f.hooks())
			if message != tc.wantMessage {
				t.Fatalf("message = %t, want %t", message, tc.wantMessage)
			}
		})
	}
}

func TestBallCollide4EBA00OwnedTypeMatchReturnsSilently(t *testing.T) {
	item2 := &ballCollideTestObject4EBA00{name: "item2", typeInd: 0x44}
	item1 := &ballCollideTestObject4EBA00{name: "item1", typeInd: 0x33, ownedNext: item2}
	target := &ballCollideTestObject4EBA00{name: "target", class: 4, ownedFirst: item1}
	ball := &ballCollideTestObject4EBA00{
		name: "ball", typeInd: 0x44,
		data: &ballCollideTestData4EBA00{name: "data", carrier: &ballCollideTestObject4EBA00{name: "old"}},
	}
	f := &ballCollideTestFixture4EBA00{teams: map[uint8]*ballCollideTestTeam4EBA00{}}
	ballCollide4EBA00(ball, target, nil, f.hooks())
	assertBallCollideEvents4EBA00(t, f.events, []string{
		"update:ball=data", "team-id:target=0x0", "find-team:0x0=nil", "carrier:data=old",
		"owner:ball=nil", "class:target=0x4", "owned-first:target=item1", "type:ball=0x44",
		"type:item1=0x33", "owned-next:item1=item2", "type:item2=0x44",
	})
}

func TestBallCollide4EBA00ExistingTeamPickupUsesCachedTeamAndLivePostCallbacks(t *testing.T) {
	team := &ballCollideTestTeam4EBA00{name: "red", kind: 1}
	target := &ballCollideTestObject4EBA00{name: "target", teamID: 1, class: 4, netCode: 0xaaaa0001}
	ball := &ballCollideTestObject4EBA00{
		name: "ball", hasTeam: true, netCode: 0x11110001, flags: 2,
		data: &ballCollideTestData4EBA00{name: "data", carrier: &ballCollideTestObject4EBA00{name: "old"}},
	}
	f := &ballCollideTestFixture4EBA00{
		teams:         map[uint8]*ballCollideTestTeam4EBA00{1: team},
		carrierResult: &ballCollideTestObject4EBA00{name: "ignored-result"},
	}
	f.onHasTeam = func() { ball.netCode = 0x22220002 }
	f.onChangeTeam = func() {
		team.kind = 2
		target.netCode = 0xbbbb1234
	}
	f.onAudio = func(id uint32) {
		if id == ballCollidePickupAudio4EBA00 {
			ball.flags = 0x100
		}
	}
	ballCollide4EBA00(ball, target, struct{ unread byte }{7}, f.hooks())
	assertBallCollideEvents4EBA00(t, f.events, []string{
		"update:ball=data", "team-id:target=0x1", "find-team:0x1=red", "carrier:data=old",
		"owner:ball=nil", "class:target=0x4", "owned-first:target=nil", "set-owner:target->ball",
		"has-team:ball=true", "net-code:ball=0x22220002", "change-team:ball:red:0x22220002:0",
		"team-kind:red=0x2", "net-code16:target=0x1234", "status:0x4:0x1234",
		"carrier-state:ball:target", "audio:927:ball", "flags:ball=0x100",
		"store-flags:ball=0x140", "purge:target",
	})
	if ball.owner != target || ball.data.carrier != target || ball.flags != 0x140 {
		t.Fatalf("pickup result owner=%s carrier=%s flags=%#x", ballCollideObjectName4EBA00(ball.owner), ballCollideObjectName4EBA00(ball.data.carrier), ball.flags)
	}
}

func TestBallCollide4EBA00NoTeamPathReloadsTargetIDAfterSourceNetCode(t *testing.T) {
	team := &ballCollideTestTeam4EBA00{name: "blue", members: -1, kind: 1}
	target := &ballCollideTestObject4EBA00{name: "target", teamID: 1, class: 4, netCode: 0x11110001}
	ball := &ballCollideTestObject4EBA00{
		name: "ball", netCode: 0x22220002,
		data: &ballCollideTestData4EBA00{name: "data", carrier: target},
	}
	f := &ballCollideTestFixture4EBA00{teams: map[uint8]*ballCollideTestTeam4EBA00{1: team}}
	f.onSetOwner = func() { target.teamID = 7 }
	f.onHasTeam = func() { ball.netCode = 0xdeadbeef }
	f.onCreateTeam = func() {
		team.kind = 2
		target.netCode = 0x12345678
	}
	ballCollide4EBA00(ball, target, nil, f.hooks())
	assertBallCollideEvents4EBA00(t, f.events, []string{
		"update:ball=data", "team-id:target=0x1", "find-team:0x1=blue", "carrier:data=target",
		"members:blue=-1", "owner:ball=nil", "class:target=0x4", "owned-first:target=nil",
		"set-owner:target->ball", "has-team:ball=false", "net-code:ball=0xdeadbeef",
		"team-id:target=0x7", "create-team:0x7:ball:1:0xdeadbeef:0",
		"team-kind:blue=0x2", "net-code16:target=0x5678", "status:0x4:0x5678",
		"carrier-state:ball:target", "audio:927:ball", "flags:ball=0x0",
		"store-flags:ball=0x40", "purge:target",
	})
}

func TestBallCollide4EBA00MissingCachedTeamSkipsTeamMutationButPicksUp(t *testing.T) {
	target := &ballCollideTestObject4EBA00{name: "target", teamID: 9, class: 4}
	ball := &ballCollideTestObject4EBA00{
		name: "ball", hasTeam: true,
		data: &ballCollideTestData4EBA00{name: "data", carrier: &ballCollideTestObject4EBA00{name: "old"}},
	}
	f := &ballCollideTestFixture4EBA00{teams: map[uint8]*ballCollideTestTeam4EBA00{}}
	ballCollide4EBA00(ball, target, nil, f.hooks())
	assertBallCollideEvents4EBA00(t, f.events, []string{
		"update:ball=data", "team-id:target=0x9", "find-team:0x9=nil", "carrier:data=old",
		"owner:ball=nil", "class:target=0x4", "owned-first:target=nil", "set-owner:target->ball",
		"has-team:ball=true", "carrier-state:ball:target", "audio:927:ball",
		"flags:ball=0x0", "store-flags:ball=0x40", "purge:target",
	})
}

func TestBallCollide4EBA00NilUpdateFaultFollowsTargetTeamLookup(t *testing.T) {
	ball := &ballCollideTestObject4EBA00{name: "ball"}
	target := &ballCollideTestObject4EBA00{name: "target", teamID: 3}
	f := &ballCollideTestFixture4EBA00{teams: map[uint8]*ballCollideTestTeam4EBA00{}}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update-data did not fault")
		}
		assertBallCollideEvents4EBA00(t, f.events, []string{
			"update:ball=nil", "team-id:target=0x3", "find-team:0x3=nil", "carrier:nil",
		})
	}()
	ballCollide4EBA00(ball, target, nil, f.hooks())
}
