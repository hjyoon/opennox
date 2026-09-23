package opennox

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

func TestConsumeObjectSpecialFlags527E50OrderAndClears(t *testing.T) {
	state := uint32(0x80000000 | 0x04000000 | 0x00800000 | 0x00020000 | 0x000000ab)
	var calls []string
	got := consumeObjectSpecialFlags527E50(&state, objectSpecialHooks527E50{
		reportAnimation: func() { calls = append(calls, "animation") },
		reportHealth:    func() { calls = append(calls, "health") },
		reportHidden:    func() { calls = append(calls, "hidden") },
		reportXStatus:   func() { calls = append(calls, "status") },
		reportHeight:    func() { calls = append(calls, "height") },
		reportEnchant:   func() { calls = append(calls, "enchant") },
		reportTeamBase:  func() { calls = append(calls, "team") },
		reportNPC:       func() { calls = append(calls, "npc") },
	})
	if !got {
		t.Fatal("pending special state was not consumed")
	}
	if want := []string{"health", "enchant", "npc"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
	if state != 0x800000ab {
		t.Fatalf("remaining state = %#08x, want %#08x", state, uint32(0x800000ab))
	}
	if consumeObjectSpecialFlags527E50(&state, objectSpecialHooks527E50{}) {
		t.Fatal("unrelated state was reported as consumed")
	}
}

func TestObjectPacketsNative519410UseNamedFields(t *testing.T) {
	obj := &server.Object{
		TypeInd:    0x1234,
		ObjClass:   object.ClassSimple,
		NetCode:    0x2345,
		PosVec:     types.Ptf(10.5, 11.5),
		Direction1: 0,
	}
	s := &Server{Server: &server.Server{}}
	if got, want := s.simpleObjectPacketNative5188A0(obj), [9]byte{47, 0x45, 0x23, 0x34, 0x12, 10, 0, 12, 0}; got != want {
		t.Fatalf("simple packet = % x, want % x", got, want)
	}
	obj.ObjClass = object.ClassComplex
	if got, want := s.phantomObjectPacketNative5187E0(obj), [11]byte{48, 0x45, 0x23, 0x34, 0x12, 10, 0, 12, 0, 0x40, 0xff}; got != want {
		t.Fatalf("phantom packet = % x, want % x", got, want)
	}
}

func TestNetSpriteUpdateStateNative518AE0MatchesFixedObjectProtocols(t *testing.T) {
	typeIndices := map[string]int{
		"TeleportPentagram": 11,
		"Spike":             12,
		"PressurePlate":     13,
	}
	hooks := netSpriteUpdateHooks518AE0{
		typeIndex: func(id string) int { return typeIndices[id] },
	}
	tests := []struct {
		name   string
		obj    *server.Object
		opcode netmsg.Op
		value  byte
		direct bool
	}{
		{
			name: "obelisk",
			obj: &server.Object{
				ObjClass:    object.ClassImmobile,
				ObjSubClass: object.SubClass(object.OtherVisibleObelisk),
				UpdateData:  unsafe.Pointer(&server.ObeliskUpdateData{Mana: 0x123}),
			},
			opcode: netmsg.MSG_OBELISK_CHARGE,
			value:  0x23,
			direct: true,
		},
		{
			name: "teleport pentagram",
			obj: &server.Object{
				TypeInd:    11,
				ObjClass:   object.ClassImmobile,
				UpdateData: unsafe.Pointer(&server.PentagramUpdateData{AnimationStep: 7}),
			},
			opcode: netmsg.MSG_DRAW_FRAME,
			value:  7,
		},
		{
			name: "spike raised",
			obj: &server.Object{
				TypeInd:  12,
				ObjClass: object.ClassImmobile,
				ObjFlags: 0,
			},
			opcode: netmsg.MSG_DRAW_FRAME,
			value:  1,
		},
		{
			name: "spike lowered",
			obj: &server.Object{
				TypeInd:  12,
				ObjClass: object.ClassImmobile,
				ObjFlags: object.FlagEquipped,
			},
			opcode: netmsg.MSG_DRAW_FRAME,
			value:  0,
		},
		{
			name: "pressure plate",
			obj: &server.Object{
				TypeInd:    13,
				ObjClass:   object.ClassImmobile | object.ClassTrigger,
				UpdateData: unsafe.Pointer(&server.TriggerUpdateData{Flags: 3}),
			},
			opcode: netmsg.MSG_PENTAGRAM_ACTIVATE,
			value:  1,
		},
		{
			name: "door",
			obj: &server.Object{
				ObjClass:   object.ClassImmobile | object.ClassDoor,
				UpdateData: unsafe.Pointer(&server.DoorUpdateData{CurrentDirection: 0x123}),
			},
			opcode: netmsg.MSG_DOOR_ANGLE,
			value:  0x23,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := netSpriteUpdateStateNative518AE0(tc.obj, hooks)
			if !ok {
				t.Fatal("fixed-object state was not selected")
			}
			if got.opcode != tc.opcode || got.value != tc.value || got.direct != tc.direct {
				t.Fatalf("state = %+v, want opcode=%d value=%d direct=%t", got, tc.opcode, tc.value, tc.direct)
			}
		})
	}
}

func TestNetSpriteUpdateStateNative518AE0ElevatorFramesAndNativeShaftLink(t *testing.T) {
	for _, tc := range []struct {
		height uint32
		frame  byte
	}{
		{height: 0, frame: 0},
		{height: 2, frame: 0},
		{height: 63, frame: 15},
		{height: 64, frame: 16},
		{height: 0x104, frame: 1},
	} {
		update := &server.ElevatorUpdateData{Field_4: tc.height}
		elevator := &server.Object{
			ObjClass:   object.ClassImmobile | object.ClassElevator,
			UpdateData: unsafe.Pointer(update),
		}
		got, ok := netSpriteUpdateStateNative518AE0(elevator, netSpriteUpdateHooks518AE0{})
		if !ok || got.opcode != netmsg.MSG_DRAW_FRAME || got.value != tc.frame {
			t.Fatalf("height %#x state = %+v, ok=%t; want draw frame %d", tc.height, got, ok, tc.frame)
		}
	}

	elevatorUpdate := &server.ElevatorUpdateData{Field_4: 63}
	elevator := &server.Object{
		ObjClass:   object.ClassImmobile | object.ClassElevator,
		UpdateData: unsafe.Pointer(elevatorUpdate),
	}
	shaftUpdate := &server.ElevatorShaftUpdateData{Field_1: 0xffffffff}
	shaft := &server.Object{
		ObjClass:   object.ClassImmobile | object.ClassElevatorShaft,
		UpdateData: unsafe.Pointer(shaftUpdate),
	}
	got, ok := netSpriteUpdateStateNative518AE0(shaft, netSpriteUpdateHooks518AE0{
		elevatorLink: func(got *server.Object) *server.Object {
			if got != shaft {
				t.Fatalf("shaft link lookup object = %p, want %p", got, shaft)
			}
			return elevator
		},
	})
	if !ok || got.opcode != netmsg.MSG_DRAW_FRAME || got.value != 15 {
		t.Fatalf("shaft state = %+v, ok=%t; want linked draw frame 15", got, ok)
	}
	if shaftUpdate.Field_1 != 0xffffffff {
		t.Fatalf("legacy PE32 link slot changed to %#x", shaftUpdate.Field_1)
	}
}

func TestNetSendObjects2PlayerNative519410QueuesElevatorDrawFrame(t *testing.T) {
	list := netlist.New()
	list.Init()
	defer list.Free()
	base := &server.Server{NetList: list}
	s := &Server{Server: base}
	const playerIndex = 5
	player := &server.Player{PlayerInd: playerIndex}
	playerUpdate := &server.PlayerUpdateData{Player: player}
	recipient := &server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(playerUpdate),
	}
	elevatorUpdate := &server.ElevatorUpdateData{Field_4: 64}
	elevator := &server.Object{
		ObjClass:   object.ClassImmobile | object.ClassElevator,
		Extent:     0x1234,
		Field38:    1 << playerIndex,
		UpdateData: unsafe.Pointer(elevatorUpdate),
	}

	if !s.netSendObjects2PlayerNative519410(recipient, elevator) {
		t.Fatal("elevator state was not sent")
	}
	want := []byte{byte(netmsg.MSG_DRAW_FRAME), 0x34, 0x92, 16}
	if got := base.NetList.CopyPacketsA(playerIndex, netlist.Kind1); !reflect.DeepEqual(got, want) {
		t.Fatalf("elevator packet = % x, want % x", got, want)
	}
	bit := uint32(1 << playerIndex)
	if elevator.Field37&bit == 0 || elevator.Field38&bit != 0 {
		t.Fatalf("elevator visibility/dirty bits = %#x/%#x, want visible/clean", elevator.Field37, elevator.Field38)
	}
}

func TestHealthDeltaPacketNative4D8760DamageHealingAndDelay(t *testing.T) {
	obj := &server.Object{
		NetCode:    0x2345,
		Frame134:   100,
		HealthData: &server.HealthData{Cur: 80, Max: 120},
	}
	cached := uint16(100)
	if _, ok := healthDeltaPacketNative4D8760(102, obj, &cached); ok || cached != 100 {
		t.Fatalf("two-frame report = sent:%t cache:%d, want false/100", ok, cached)
	}
	packet, ok := healthDeltaPacketNative4D8760(103, obj, &cached)
	if !ok {
		t.Fatal("damage report was not emitted after three frames")
	}
	if packet[0] != byte(netmsg.MSG_REPORT_HEALTH_DELTA) ||
		binary.LittleEndian.Uint16(packet[1:]) != 0x2345 ||
		int16(binary.LittleEndian.Uint16(packet[3:])) != -20 {
		t.Fatalf("damage packet = % x, want opcode/id/delta 66/0x2345/-20", packet)
	}
	if cached != 80 {
		t.Fatalf("damage cache = %d, want 80", cached)
	}

	obj.HealthData.Cur = 90
	if _, ok := healthDeltaPacketNative4D8760(104, obj, &cached); ok || cached != 90 {
		t.Fatalf("healing report = sent:%t cache:%d, want false/90", ok, cached)
	}
	if _, ok := healthDeltaPacketNative4D8760(105, obj, &cached); ok || cached != 90 {
		t.Fatalf("unchanged report = sent:%t cache:%d, want false/90", ok, cached)
	}
}

func TestUnitHealthSampleNative4D8760UsesRecipientCaches(t *testing.T) {
	monsterUpdate := new(server.MonsterUpdateData)
	monster := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(monsterUpdate)}
	monsterSample := unitHealthSampleNative4D8760(monster, 7)
	if monsterSample != &monsterUpdate.HealthGraph103[7] {
		t.Fatalf("monster sample = %p, want %p", monsterSample, &monsterUpdate.HealthGraph103[7])
	}

	playerUpdate := new(server.PlayerUpdateData)
	player := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUpdate)}
	playerSample := unitHealthSampleNative4D8760(player, 9)
	if playerSample != &playerUpdate.HealthSamples[9] {
		t.Fatalf("player sample = %p, want %p", playerSample, &playerUpdate.HealthSamples[9])
	}

	generatorUpdate := new(server.MonsterGenUpdateData)
	generator := &server.Object{ObjClass: object.ClassMonsterGenerator, UpdateData: unsafe.Pointer(generatorUpdate)}
	generatorSample := unitHealthSampleNative4D8760(generator, 11)
	if generatorSample != &generatorUpdate.HealthSamples[11] {
		t.Fatalf("generator sample = %p, want %p", generatorSample, &generatorUpdate.HealthSamples[11])
	}
	if got := unitHealthSampleNative4D8760(player, 32); got != nil {
		t.Fatalf("out-of-range sample = %p, want nil", got)
	}
}

func TestReportUnitHealthDeltaNative4D8760SendsOriginalPacketMode(t *testing.T) {
	base := &server.Server{}
	base.SetFrame(50)
	var (
		recipient, remove, sequence int
		packet                      []byte
		related                     *server.Object
	)
	base.NetSendPacketXxx = func(gotRecipient int, gotPacket []byte, gotRelated *server.Object, gotRemove, gotSequence int) int {
		recipient, remove, sequence = gotRecipient, gotRemove, gotSequence
		packet = append(packet[:0], gotPacket...)
		related = gotRelated
		return 1
	}

	update := new(server.MonsterUpdateData)
	update.HealthGraph103[4] = 70
	obj := &server.Object{
		ObjClass:   object.ClassMonster,
		NetCode:    0x3456,
		Frame134:   47,
		HealthData: &server.HealthData{Cur: 55, Max: 100},
		UpdateData: unsafe.Pointer(update),
	}
	(&Server{Server: base}).reportUnitHealthDeltaNative4D8760(4, obj)

	want := []byte{byte(netmsg.MSG_REPORT_HEALTH_DELTA), 0x56, 0x34, 0xf1, 0xff}
	if recipient != 4 || remove != 1 || sequence != 0 || related != nil || !reflect.DeepEqual(packet, want) {
		t.Fatalf("health report = recipient:%d packet:% x related:%p remove:%d sequence:%d; want 4/% x/nil/1/0",
			recipient, packet, related, remove, sequence, want)
	}
	if update.HealthGraph103[4] != 55 {
		t.Fatalf("health cache = %d, want 55", update.HealthGraph103[4])
	}
}

func TestMonsterGeneratorHealthDeltaNative4D8760UsesRecipientCache(t *testing.T) {
	base := &server.Server{}
	base.SetFrame(80)
	var packet []byte
	base.NetSendPacketXxx = func(_ int, gotPacket []byte, _ *server.Object, _, _ int) int {
		packet = append(packet[:0], gotPacket...)
		return 1
	}

	update := new(server.MonsterGenUpdateData)
	update.HealthSamples[6] = 90
	obj := &server.Object{
		ObjClass:   object.ClassSimple | object.ClassMonsterGenerator,
		NetCode:    0x4567,
		Frame134:   77,
		HealthData: &server.HealthData{Cur: 72, Max: 100},
		UpdateData: unsafe.Pointer(update),
	}
	(&Server{Server: base}).reportUnitHealthDeltaNative4D8760(6, obj)

	want := []byte{byte(netmsg.MSG_REPORT_HEALTH_DELTA), 0x67, 0x45, 0xee, 0xff}
	if !reflect.DeepEqual(packet, want) {
		t.Fatalf("generator health report = % x, want % x", packet, want)
	}
	if update.HealthSamples[6] != 72 {
		t.Fatalf("generator health cache = %d, want 72", update.HealthSamples[6])
	}
}

func TestComplexObjectVisualStateNative519410(t *testing.T) {
	data := []byte{48, 2, 0, 4, 0, 10, 0, 20, 0, 0x75, 9}
	state, ok := decodeComplexObjectVisualState519410(data)
	if !ok {
		t.Fatal("valid complex-object packet was rejected")
	}
	if state.Code != 2 || state.TypeID != 4 || state.X != 10 || state.Y != 20 || state.Direction != 8 || state.Animation != 5 || state.Frame != 9 {
		t.Fatalf("decoded state = %+v", state)
	}
	dr := &client.Drawable{AnimFrameSlave: 3, AnimInd: 4, AnimStart: 5}
	applyComplexObjectVisualState519410(dr, state, 100)
	if dr.Field_72 != 100 || dr.AnimFrameSlave != 9 || dr.Field_78 != 3 || dr.AnimDir != 8 || dr.AnimInd != 5 || dr.AnimStart != 100 {
		t.Fatalf("drawable state was not applied through native fields: %+v", dr)
	}
	if _, ok := decodeSimpleObjectVisualState519410(data[:8]); ok {
		t.Fatal("short simple-object packet was accepted")
	}
	if _, ok := decodeComplexObjectVisualState519410(data[:10]); ok {
		t.Fatal("short complex-object packet was accepted")
	}
}

func TestObjectEnchantVisualStateNative48EA70(t *testing.T) {
	state, ok := decodeObjectEnchantVisualState48EA70([]byte{90, 0x34, 0x92, 0x00, 0x80, 0x00, 0x00})
	if !ok || state.Code != 0x9234 || state.Buffs != 0x8000 {
		t.Fatalf("decoded enchant state = %+v, ok=%t", state, ok)
	}
	if _, ok := decodeObjectEnchantVisualState48EA70(make([]byte, 6)); ok {
		t.Fatal("short enchant packet was accepted")
	}

	dr := &client.Drawable{Buffs: 1 << server.ENCHANT_LIGHT, LightIntensity: 200}
	if !applyObjectEnchantVisualState48EA70(dr, 0, nil, 0, 37.5) {
		t.Fatal("removed light enchant did not restore the thing intensity")
	}
	if dr.Buffs != 0 || dr.LightIntensity != 37.5 {
		t.Fatalf("drawable enchant/light state = %#x/%v", dr.Buffs, dr.LightIntensity)
	}
	dr.Buffs = 1 << server.ENCHANT_LIGHT
	dr.LightIntensity = 200
	if applyObjectEnchantVisualState48EA70(dr, 0, dr, 8, 10) || dr.LightIntensity != 200 {
		t.Fatal("local item-light override did not preserve enchanted intensity")
	}
}

func TestPlayerMapTracksObjectNative519410CircularList(t *testing.T) {
	want := &server.Object{}
	first := &server.MinimapItem{Field4: &server.Object{}}
	second := &server.MinimapItem{Field4: want}
	first.Field8 = second
	second.Field8 = first
	pl := &server.Player{Field4580: first}
	if !playerMapTracksObjectNative519410(pl, want) {
		t.Fatal("tracked object was not found")
	}
	if playerMapTracksObjectNative519410(pl, &server.Object{}) {
		t.Fatal("untracked object was found")
	}
	if got := playerTrackedObjectCountNative519710(pl); got != 2 {
		t.Fatalf("tracked count = %d, want 2", got)
	}
	if playerTrackedObjectCountNative519710(&server.Player{}) != 0 {
		t.Fatal("empty tracked list has a nonzero count")
	}
}

func TestNetTrackedObjectRefreshDueNative519710(t *testing.T) {
	if netTrackedObjectRefreshDueNative519710(100, 80, 0) {
		t.Fatal("zero tracked objects requested a refresh")
	}
	if !netTrackedObjectRefreshDueNative519710(100, 80, 4) {
		t.Fatal("elapsed refresh window was not detected")
	}
	if netTrackedObjectRefreshDueNative519710(95, 80, 4) {
		t.Fatal("the strict GAME.EXE time comparison became inclusive")
	}
	if !netTrackedObjectRefreshDueNative519710(1, 1, 61) {
		t.Fatal("more than sixty tracked objects did not force a refresh")
	}
}
