package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultBallCollideNativeDeps4EBA00() ballCollideNativeDeps4EBA00 {
	return ballCollideNativeDeps4EBA00{
		teamByID:        func(uint8) *Team { return nil },
		teamMemberCount: func(*Team) int32 { return 0 },
		loadFrame:       func() uint32 { return 0 },
		loadFeedback:    func() uint32 { return 0 },
		priorityMessage: func(*Object, string, int32) {},
		storeFeedback:   func(uint32) {},
		audio:           func(uint32, *Object) {},
		setOwner:        func(*Object, *Object) {},
		changeTeam:      func(*ObjectTeam, *Team, uint32, int32) int32 { return 0 },
		createTeam:      func(TeamID, *ObjectTeam, int32, uint32, int32) {},
		ballStatus:      func(uint8, uint16) int32 { return 0 },
		carrierState:    func(*Object, *Object) *Object { return nil },
		purgeBuffs:      func(*Object) {},
	}
}

func TestBallCollide4EBA00NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantType := uintptr(4)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantNetCode := uintptr(36)
	wantTeam := uintptr(48)
	wantOwner := uintptr(508)
	wantOwnedNext := uintptr(512)
	wantOwnedFirst := uintptr(516)
	wantUpdate := uintptr(748)
	wantTeamSize := uintptr(80)
	wantUpdateSize := uintptr(32)
	wantUpdateTeamID := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantType = 8
		wantClass = 12
		wantFlags = 20
		wantNetCode = 40
		wantTeam = 52
		wantOwner = 552
		wantOwnedNext = 560
		wantOwnedFirst = 568
		wantUpdate = 872
		wantTeamSize = 88
		wantUpdateSize = 40
		wantUpdateTeamID = 8
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantTeam},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantOwnedNext},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantOwnedFirst},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"ObjectTeam size", unsafe.Sizeof(ObjectTeam{}), 8},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"Team size", unsafe.Sizeof(Team{}), wantTeamSize},
		{"Team.ColorInd", unsafe.Offsetof(Team{}.ColorInd), 56},
		{"GameBallUpdateData size", unsafe.Sizeof(GameBallUpdateData4EA800{}), wantUpdateSize},
		{"GameBallUpdateData.Carrier", unsafe.Offsetof(GameBallUpdateData4EA800{}.Carrier), 0},
		{"GameBallUpdateData.TeamID", unsafe.Offsetof(GameBallUpdateData4EA800{}.TeamID), wantUpdateTeamID},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestBallCollideNative4EBA00OwnedTypeMatchUsesNativeList(t *testing.T) {
	oldCarrier := &Object{}
	update := &GameBallUpdateData4EA800{Carrier: oldCarrier}
	ball := &Object{TypeInd: 0x44, UpdateData: unsafe.Pointer(update)}
	match := &Object{TypeInd: 0x44}
	first := &Object{TypeInd: 0x33, Field128: match}
	target := &Object{
		ObjClass: object.ClassPlayer,
		Field129: first,
	}
	events := make([]string, 0, 2)
	deps := defaultBallCollideNativeDeps4EBA00()
	deps.teamByID = func(id uint8) *Team {
		events = append(events, "team")
		if id != 0 {
			t.Fatalf("team ID = %#x, want 0", id)
		}
		return nil
	}
	deps.audio = func(uint32, *Object) { t.Fatal("duplicate path emitted audio") }
	deps.setOwner = func(*Object, *Object) { t.Fatal("duplicate path set owner") }
	deps.carrierState = func(*Object, *Object) *Object {
		t.Fatal("duplicate path changed carrier")
		return nil
	}
	deps.purgeBuffs = func(*Object) { t.Fatal("duplicate path purged buffs") }
	ballCollideNative4EBA00(ball, target, &types.Pointf{X: 12, Y: -3}, deps)
	if want := []string{"team"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if update.Carrier != oldCarrier || ball.ObjOwner != nil || target.Field129 != first || first.Field128 != match {
		t.Fatal("duplicate path mutated native object state")
	}
}

func TestBallCollideNative4EBA00ExistingTeamUsesLiveNativeFields(t *testing.T) {
	oldCarrier := &Object{}
	update := &GameBallUpdateData4EA800{Carrier: oldCarrier}
	ball := &Object{
		TypeInd:    0x44,
		ObjFlags:   object.FlagEnabled,
		NetCode:    0x11223344,
		TeamVal:    ObjectTeam{ID: 3},
		UpdateData: unsafe.Pointer(update),
	}
	target := &Object{
		ObjClass: object.ClassPlayer,
		NetCode:  0xaabb3344,
		TeamVal:  ObjectTeam{ID: 7},
	}
	team := &Team{ColorInd: TeamRed}
	events := make([]string, 0, 8)
	deps := defaultBallCollideNativeDeps4EBA00()
	deps.teamByID = func(id uint8) *Team {
		events = append(events, "team")
		if id != 7 {
			t.Fatalf("team ID = %#x, want 7", id)
		}
		return team
	}
	deps.teamMemberCount = func(*Team) int32 {
		t.Fatal("non-carrier path counted team members")
		return 0
	}
	deps.setOwner = func(owner, gotBall *Object) {
		events = append(events, "owner")
		if owner != target || gotBall != ball {
			t.Fatalf("owner args = %p/%p", owner, gotBall)
		}
		gotBall.ObjOwner = owner
	}
	deps.changeTeam = func(value *ObjectTeam, gotTeam *Team, netCode uint32, flags int32) int32 {
		events = append(events, "change-team")
		if value != &ball.TeamVal || gotTeam != team || netCode != 0x11223344 || flags != 0 {
			t.Fatalf("change-team args = %p/%p/%#x/%d", value, gotTeam, netCode, flags)
		}
		team.ColorInd = TeamBlue
		target.NetCode = 0x55667788
		return -17
	}
	deps.createTeam = func(TeamID, *ObjectTeam, int32, uint32, int32) {
		t.Fatal("existing-team path created team")
	}
	deps.ballStatus = func(state uint8, netCode uint16) int32 {
		events = append(events, "status")
		if state != ballCollideTeamStateSpecial4EBA00 || netCode != 0x7788 {
			t.Fatalf("status args = %#x/%#x", state, netCode)
		}
		return -29
	}
	deps.carrierState = func(gotBall, gotTarget *Object) *Object {
		events = append(events, "carrier")
		if gotBall != ball || gotTarget != target {
			t.Fatalf("carrier args = %p/%p", gotBall, gotTarget)
		}
		update.Carrier = gotTarget
		return oldCarrier
	}
	deps.audio = func(id uint32, obj *Object) {
		events = append(events, "audio")
		if id != ballCollidePickupAudio4EBA00 || obj != ball {
			t.Fatalf("audio args = %d/%p", id, obj)
		}
		ball.ObjFlags = object.FlagDestroyed
	}
	deps.purgeBuffs = func(obj *Object) {
		events = append(events, "purge")
		if obj != target {
			t.Fatalf("purge object = %p", obj)
		}
	}
	ballCollideNative4EBA00(ball, target, nil, deps)
	if want := []string{"team", "owner", "change-team", "status", "carrier", "audio", "purge"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if ball.ObjOwner != target || update.Carrier != target {
		t.Fatalf("owner/carrier = %p/%p, want target %p", ball.ObjOwner, update.Carrier, target)
	}
	wantFlags := object.FlagDestroyed | object.Flags(ballCollideNoCollideFlag4EBA00)
	if ball.ObjFlags != wantFlags {
		t.Fatalf("flags = %#x, want post-audio live flags %#x", ball.ObjFlags, wantFlags)
	}
}

func TestBallCollideNative4EBA00CreatesNativeObjectTeam(t *testing.T) {
	update := &GameBallUpdateData4EA800{Carrier: &Object{}}
	ball := &Object{NetCode: 0x89abcdef, UpdateData: unsafe.Pointer(update)}
	target := &Object{
		ObjClass: object.ClassPlayer,
		NetCode:  0x13572468,
		TeamVal:  ObjectTeam{ID: 5},
	}
	team := &Team{ColorInd: TeamRed}
	deps := defaultBallCollideNativeDeps4EBA00()
	deps.teamByID = func(id uint8) *Team {
		if id != 5 {
			t.Fatalf("lookup ID = %d, want 5", id)
		}
		return team
	}
	deps.setOwner = func(owner, gotBall *Object) {
		if owner != target || gotBall != ball {
			t.Fatal("wrong owner arguments")
		}
		target.TeamVal.ID = 9
		ball.NetCode = 0x76543210
	}
	deps.createTeam = func(id TeamID, value *ObjectTeam, active int32, netCode uint32, flags int32) {
		if id != 9 || value != &ball.TeamVal || active != 1 || netCode != 0x76543210 || flags != 0 {
			t.Fatalf("create-team args = %d/%p/%d/%#x/%d", id, value, active, netCode, flags)
		}
	}
	deps.changeTeam = func(*ObjectTeam, *Team, uint32, int32) int32 {
		t.Fatal("teamless ball changed an existing team")
		return 0
	}
	deps.ballStatus = func(state uint8, netCode uint16) int32 {
		if state != ballCollideTeamStateNormal4EBA00 || netCode != 0x2468 {
			t.Fatalf("status args = %#x/%#x", state, netCode)
		}
		return 0
	}
	ballCollideNative4EBA00(ball, target, nil, deps)
}

func TestBallCollideServerDeps4EBA00PreservesRuntimeServices(t *testing.T) {
	s := &Server{}
	s.SetFrame(0xfedcba98)
	team := Team{IDVal: 6, active: 1, ind: 1}
	s.Teams.Arr = []Team{{}, team}
	gotFeedback := uint32(0)
	gotChange := false
	gotCreate := false
	gotStatus := false
	runtime := BallCollideRuntime4EBA00{
		TeamMemberCount: func(got *Team) int32 {
			if got != &s.Teams.Arr[1] {
				t.Fatal("wrong count team")
			}
			return -3
		},
		LoadFeedbackFrame: func() uint32 { return 0x10203040 },
		StoreFeedback:     func(frame uint32) { gotFeedback = frame },
		ChangeTeam: func(*ObjectTeam, *Team, uint32, int32) int32 {
			gotChange = true
			return -1
		},
		CreateTeam: func(TeamID, *ObjectTeam, int32, uint32, int32) { gotCreate = true },
		BallStatus: func(uint8, uint16) int32 { gotStatus = true; return -2 },
	}
	deps := ballCollideServerDeps4EBA00(s, runtime)
	if got := deps.teamByID(6); got != &s.Teams.Arr[1] {
		t.Fatalf("team lookup = %p, want %p", got, &s.Teams.Arr[1])
	}
	if got := deps.teamMemberCount(&s.Teams.Arr[1]); got != -3 {
		t.Fatalf("member count = %d, want -3", got)
	}
	if got := deps.loadFrame(); got != 0xfedcba98 {
		t.Fatalf("frame = %#x", got)
	}
	if got := deps.loadFeedback(); got != 0x10203040 {
		t.Fatalf("feedback = %#x", got)
	}
	deps.storeFeedback(0x55667788)
	deps.changeTeam(nil, nil, 0, 0)
	deps.createTeam(0, nil, 0, 0, 0)
	deps.ballStatus(0, 0)
	if gotFeedback != 0x55667788 || !gotChange || !gotCreate || !gotStatus {
		t.Fatalf("runtime forwarding = %#x/%t/%t/%t", gotFeedback, gotChange, gotCreate, gotStatus)
	}
}
