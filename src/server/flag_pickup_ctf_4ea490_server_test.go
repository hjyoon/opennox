package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestFlagPickupCTFFlagIndex4ECBD0MatchesMaterialTable(t *testing.T) {
	tests := []struct {
		name string
		want uint32
	}{
		{"MaterialTeamRed", 1},
		{"MaterialTeamGreen", 3},
		{"MaterialTeamBlue", 2},
		{"MaterialTeamYellow", 5},
		{"MaterialTeamCyan", 4},
		{"MaterialTeamViolet", 6},
		{"MaterialTeamBlack", 7},
		{"MaterialTeamWhite", 8},
		{"MaterialTeamOrange", 9},
		{"materialteamred", 0},
		{"UnknownMaterial", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name, freeName := alloc.CString(tc.name)
			defer freeName()
			material := &ModifierEff{name0: name}
			data := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, material}}
			obj := &Object{ObjClass: object.ClassFlag, InitData: unsafe.Pointer(data)}
			if got := flagPickupCTFFlagIndex4ECBD0(obj); got != tc.want {
				t.Fatalf("flag index = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestFlagPickupCTFFlagIndex4ECBD0GatesClassAndNilMaterial(t *testing.T) {
	data := &ModifierInitData{}
	nonFlag := &Object{ObjClass: object.ClassPlayer, InitData: unsafe.Pointer(data)}
	if got := flagPickupCTFFlagIndex4ECBD0(nonFlag); got != 0 {
		t.Fatalf("non-flag index = %d", got)
	}
	flag := &Object{ObjClass: object.ClassFlag, InitData: unsafe.Pointer(data)}
	if got := flagPickupCTFFlagIndex4ECBD0(flag); got != 0 {
		t.Fatalf("nil-material index = %d", got)
	}
}

func defaultFlagPickupCTFNativeDeps4EA490() flagPickupCTFNativeDeps4EA490 {
	return flagPickupCTFNativeDeps4EA490{
		flagIndex:       func(*Object) uint32 { return 0 },
		gameData:        func(uint32) uint16 { return 0 },
		moveHome:        func(*Object, *FlagUpdateData4EA490) {},
		informReturn:    func(uint32) {},
		informFlag:      func(uint32, uint32, uint32) {},
		flagStatus:      func(uint8, uint8, uint8, uint16) int32 { return 0 },
		reportLesson:    func(*Object) {},
		teamByID:        func(uint8) *Team { return nil },
		changeTeamScore: func(*Team, int32) {},
		observerMode:    func() uint32 { return 0 },
		observerUpdate:  func(*Player, *Player) {},
		detachInventory: func(*Object, *Object) {},
		createAt:        func(*Object, *Object, types.Pointf) {},
		raise:           func(*Object, float32) {},
		markMinimap:     func(*Object, uint32) {},
		firstTeam:       func() *Team { return nil },
		nextTeam:        func(*Team) *Team { return nil },
		setGameFlags:    func(uint32) {},
		flagWinner:      func(*Team, uint32) {},
		teamEligible:    func(*Team) int32 { return 0 },
		forceDrop:       func(*Object, *Object) {},
		finalizeDelete:  func(*Object) {},
		inventoryPut:    func(*Object, *Object, int32) {},
		reportObject:    func(uint32, *Object) {},
		unmarkMinimap:   func(*Object, uint32) {},
		purgeBuffs:      func(*Object) {},
		priorityMessage: func(*Object, string, uint32) {},
	}
}

func TestFlagPickupCTF4EA490NativeLayout(t *testing.T) {
	wantObjectUpdate := uintptr(748)
	wantObjectTeam := uintptr(48)
	wantObjectPos := uintptr(56)
	wantObjectNet := uintptr(36)
	wantObjectHolder := uintptr(492)
	wantObjectNext := uintptr(496)
	wantObjectFirst := uintptr(504)
	wantPlayerUpdatePlayer := uintptr(276)
	wantPlayerLessons := uintptr(2136)
	wantTeamLessons := uintptr(52)
	wantTeamID := uintptr(57)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectUpdate = 872
		wantObjectTeam = 52
		wantObjectPos = 60
		wantObjectNet = 40
		wantObjectHolder = 520
		wantObjectNext = 528
		wantObjectFirst = 544
		wantPlayerUpdatePlayer = 320
		wantPlayerLessons = 2140
		wantTeamLessons = 56
		wantTeamID = 65
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"FlagUpdateData size", unsafe.Sizeof(FlagUpdateData4EA490{}), 12},
		{"FlagUpdateData.Home", unsafe.Offsetof(FlagUpdateData4EA490{}.Home), 0},
		{"FlagUpdateData.State", unsafe.Offsetof(FlagUpdateData4EA490{}.State), 8},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantObjectTeam},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantObjectPos},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantObjectNet},
		{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), wantObjectHolder},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantObjectNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantObjectFirst},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerUpdatePlayer},
		{"Player.WeaponEquip", unsafe.Offsetof(Player{}.WeaponEquip), 4},
		{"Player.Lessons", unsafe.Offsetof(Player{}.Lessons), wantPlayerLessons},
		{"Team.Lessons", unsafe.Offsetof(Team{}.Lessons), wantTeamLessons},
		{"Team.IDVal", unsafe.Offsetof(Team{}.IDVal), wantTeamID},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestFlagPickupCTFNative4EA490ReturnHomeUsesTypedUpdate(t *testing.T) {
	update := &FlagUpdateData4EA490{Home: types.Pointf{X: 1, Y: 2}, State: 99}
	targetUpdate := &PlayerUpdateData{}
	source := &Object{
		TeamVal:    ObjectTeam{ID: 5},
		PosVec:     types.Pointf{X: 7, Y: 2},
		UpdateData: unsafe.Pointer(update),
	}
	target := &Object{
		NetCode:    0x12345678,
		TeamVal:    ObjectTeam{ID: 5},
		UpdateData: unsafe.Pointer(targetUpdate),
	}
	deps := defaultFlagPickupCTFNativeDeps4EA490()
	deps.flagIndex = func(got *Object) uint32 {
		if got != source {
			t.Fatal("wrong flag-index object")
		}
		return 0x102
	}
	deps.moveHome = func(got *Object, gotUpdate *FlagUpdateData4EA490) {
		if got != source || gotUpdate != update {
			t.Fatal("move did not receive typed cached update")
		}
		got.PosVec = gotUpdate.Home
		target.TeamVal.ID = 9
	}
	deps.informReturn = func(netCode uint32) {
		if netCode != 0x12345678 {
			t.Fatalf("return net code = %#x", netCode)
		}
	}
	deps.flagStatus = func(teamID, status, index uint8, carrier uint16) int32 {
		if teamID != 9 || status != 0 || index != 2 || carrier != 0 {
			t.Fatalf("status = %d/%d/%d/%#x", teamID, status, index, carrier)
		}
		return -1
	}
	flagPickupCTFNative4EA490(source, target, nil, deps)
	if source.PosVec != update.Home || update.State != 0 {
		t.Fatalf("source/state = %v/%d", source.PosVec, update.State)
	}
}

func TestFlagPickupCTFNative4EA490ScoresThroughNamedFields(t *testing.T) {
	sourceUpdate := &FlagUpdateData4EA490{Home: types.Pointf{X: 10, Y: 20}}
	itemUpdate := &FlagUpdateData4EA490{Home: types.Pointf{X: 30, Y: 40}, State: 7}
	player := &Player{Lessons: 41}
	targetUpdate := &PlayerUpdateData{Player: player}
	item := &Object{
		ObjClass:   object.ClassFlag,
		TeamVal:    ObjectTeam{ID: 4},
		UpdateData: unsafe.Pointer(itemUpdate),
	}
	source := &Object{
		TeamVal:    ObjectTeam{ID: 2},
		PosVec:     sourceUpdate.Home,
		UpdateData: unsafe.Pointer(sourceUpdate),
	}
	target := &Object{
		ObjClass:     object.ClassPlayer,
		NetCode:      0xabcdef01,
		TeamVal:      ObjectTeam{ID: 2},
		InvFirstItem: item,
		UpdateData:   unsafe.Pointer(targetUpdate),
	}
	team := &Team{Lessons: 5}
	deps := defaultFlagPickupCTFNativeDeps4EA490()
	deps.flagIndex = func(obj *Object) uint32 {
		if obj == item {
			return 0x123
		}
		return 7
	}
	deps.teamByID = func(id uint8) *Team {
		if id != 2 {
			t.Fatalf("team id = %d", id)
		}
		return team
	}
	deps.changeTeamScore = func(got *Team, score int32) {
		if got != team || score != 6 {
			t.Fatalf("team score = %p/%d", got, score)
		}
	}
	deps.detachInventory = func(owner, got *Object) {
		if owner != target || got != item {
			t.Fatal("wrong detach objects")
		}
		itemUpdate.Home = types.Pointf{X: 50, Y: 60}
	}
	deps.createAt = func(got, owner *Object, pos types.Pointf) {
		if got != item || owner != nil || pos != itemUpdate.Home {
			t.Fatalf("create = %p/%p/%v", got, owner, pos)
		}
	}
	deps.flagStatus = func(teamID, status, index uint8, carrier uint16) int32 {
		if teamID != 4 || status != 0 || index != 0x23 || carrier != 0 {
			t.Fatalf("status = %d/%d/%#x/%#x", teamID, status, index, carrier)
		}
		return 0
	}
	deps.informFlag = func(code, netCode, index uint32) {
		if code != 5 || netCode != 0xabcdef01 || index != 0x123 {
			t.Fatalf("inform = %d/%#x/%#x", code, netCode, index)
		}
	}
	flagPickupCTFNative4EA490(source, target, nil, deps)
	if player.Lessons != 42 {
		t.Fatalf("player lessons = %d", player.Lessons)
	}
	if itemUpdate.State != 0 {
		t.Fatalf("item state = %d", itemUpdate.State)
	}
}

func TestFlagPickupCTFNative4EA490EnemyKeepsCachedPointers(t *testing.T) {
	sourceUpdate := &FlagUpdateData4EA490{State: 7}
	replacementSourceUpdate := &FlagUpdateData4EA490{State: 8}
	oldPlayer := &Player{WeaponEquip: 0x20}
	newPlayer := &Player{WeaponEquip: 0x40}
	targetUpdate := &PlayerUpdateData{Player: oldPlayer}
	replacementTargetUpdate := &PlayerUpdateData{Player: newPlayer}
	carriedFlag := &Object{ObjClass: object.ClassFlag}
	nonFlag := &Object{InvNextItem: carriedFlag}
	source := &Object{
		TeamVal:    ObjectTeam{ID: 1},
		UpdateData: unsafe.Pointer(sourceUpdate),
	}
	target := &Object{
		NetCode:      0x12345678,
		TeamVal:      ObjectTeam{ID: 2},
		InvFirstItem: nonFlag,
		UpdateData:   unsafe.Pointer(targetUpdate),
	}
	team := &Team{}
	indexLoads := 0
	deps := defaultFlagPickupCTFNativeDeps4EA490()
	deps.flagIndex = func(got *Object) uint32 {
		indexLoads++
		if got != source {
			t.Fatal("wrong index object")
		}
		if indexLoads == 1 {
			return 0x123
		}
		return 0x456
	}
	deps.teamByID = func(id uint8) *Team {
		if id != 1 {
			t.Fatalf("team id = %d", id)
		}
		return team
	}
	deps.teamEligible = func(got *Team) int32 {
		if got != team {
			t.Fatal("wrong eligible team")
		}
		return 1
	}
	deps.forceDrop = func(owner, item *Object) {
		if owner != target || item != carriedFlag {
			t.Fatal("wrong force-drop objects")
		}
	}
	deps.finalizeDelete = func(got *Object) {
		if got != source {
			t.Fatal("wrong finalized object")
		}
		source.UpdateData = unsafe.Pointer(replacementSourceUpdate)
	}
	deps.inventoryPut = func(owner, item *Object, mode int32) {
		if owner != target || item != source || mode != 1 {
			t.Fatal("wrong inventory-put args")
		}
		target.UpdateData = unsafe.Pointer(replacementTargetUpdate)
	}
	deps.reportObject = func(recipient uint32, got *Object) {
		if recipient != 255 || got != source {
			t.Fatal("wrong report-object args")
		}
	}
	deps.informFlag = func(code, netCode, index uint32) {
		if code != 6 || netCode != 0x12345678 || index != 0x456 {
			t.Fatalf("inform = %d/%#x/%#x", code, netCode, index)
		}
		target.NetCode = 0x89abcdef
	}
	deps.flagStatus = func(teamID, status, index uint8, carrier uint16) int32 {
		if teamID != 1 || status != 1 || index != 0x23 || carrier != 0xcdef {
			t.Fatalf("status = %d/%d/%#x/%#x", teamID, status, index, carrier)
		}
		return 0
	}
	purged := false
	deps.purgeBuffs = func(got *Object) {
		if got != target {
			t.Fatal("wrong purge target")
		}
		purged = true
	}
	flagPickupCTFNative4EA490(source, target, nil, deps)
	if oldPlayer.WeaponEquip != 0x21 || newPlayer.WeaponEquip != 0x40 {
		t.Fatalf("old/new weapon flags = %#x/%#x", oldPlayer.WeaponEquip, newPlayer.WeaponEquip)
	}
	if sourceUpdate.State != 0 || replacementSourceUpdate.State != 8 {
		t.Fatalf("old/new source states = %d/%d", sourceUpdate.State, replacementSourceUpdate.State)
	}
	if !purged || indexLoads != 2 {
		t.Fatalf("purged/index loads = %t/%d", purged, indexLoads)
	}
}
