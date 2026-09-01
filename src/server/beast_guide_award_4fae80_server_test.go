package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
)

func defaultBeastGuideAwardNativeDeps4FAE80() beastGuideAwardNativeDeps4FAE80 {
	return beastGuideAwardNativeDeps4FAE80{
		loadString:       func(string, string, int) string { return "" },
		sendLineMessage:  func(*Object, string) {},
		awardProtection:  func(uint32, int32, int32) {},
		audio:            func(uint32, *Object, int32, uint32) {},
		rewardNotify:     func(*Object, int32, *Object, int32) {},
		relatedGuides:    func(int32) []int32 { return nil },
		firstPlayer:      func() *Player { return nil },
		nextPlayer:       func(*Player) *Player { return nil },
		reportGuideAward: func(*Object, int32, int32, int32) {},
	}
}

func TestBeastGuideAwardNative4FAE80PreservesPointersAndScalars(t *testing.T) {
	player := &Player{Prot4640: 0x89abcdef}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: 0xf4, UpdateData: unsafe.Pointer(update)}
	player.PlayerUnit = unit
	other := &Object{ObjClass: 4}
	otherPlayer := &Player{PlayerUnit: other}

	type protectionCall struct {
		token uint32
		guide int32
		level int32
	}
	type notification struct {
		recipient *Object
		kind      int32
		source    *Object
		guide     int32
	}
	var (
		protections   []protectionCall
		notifications []notification
		audioUnit     *Object
		reportUnit    *Object
		reportArgs    [3]int32
	)
	deps := defaultBeastGuideAwardNativeDeps4FAE80()
	deps.awardProtection = func(token uint32, guide, level int32) {
		protections = append(protections, protectionCall{token, guide, level})
	}
	deps.audio = func(id uint32, gotUnit *Object, kind int32, code uint32) {
		if id != beastGuideAwardSound4FAE80 || kind != 0 || code != 0 {
			t.Fatalf("audio scalars = %d/%d/%#x", id, kind, code)
		}
		audioUnit = gotUnit
	}
	deps.rewardNotify = func(recipient *Object, kind int32, source *Object, guide int32) {
		notifications = append(notifications, notification{recipient, kind, source, guide})
	}
	deps.relatedGuides = func(guide int32) []int32 {
		if guide != 24 {
			t.Fatalf("relations guide = %d, want 24", guide)
		}
		return []int32{8}
	}
	deps.firstPlayer = func() *Player { return player }
	deps.nextPlayer = func(got *Player) *Player {
		switch got {
		case player:
			return otherPlayer
		case otherPlayer:
			return nil
		default:
			t.Fatalf("unexpected player pointer %p", got)
			return nil
		}
	}
	deps.reportGuideAward = func(gotUnit *Object, guide, notify, shop int32) {
		reportUnit = gotUnit
		reportArgs = [3]int32{guide, notify, shop}
	}

	const notify = int32(math.MinInt32 + 0x1234)
	if got := beastGuideAwardNative4FAE80(unit, 24, notify, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if player.BeastScrollLvl[24] != 1 || player.BeastScrollLvl[8] != 1 {
		t.Fatalf("guide levels = %d/%d, want 1/1", player.BeastScrollLvl[24], player.BeastScrollLvl[8])
	}
	wantProtections := []protectionCall{
		{0x89abcdef, 24, 1},
		{0x89abcdef, 8, 1},
	}
	if !reflect.DeepEqual(protections, wantProtections) {
		t.Fatalf("protections = %#v, want %#v", protections, wantProtections)
	}
	if audioUnit != unit {
		t.Fatalf("audio unit = %p, want %p", audioUnit, unit)
	}
	wantNotifications := []notification{
		{unit, 1, unit, 24},
		{other, 1, unit, 24},
	}
	if !reflect.DeepEqual(notifications, wantNotifications) {
		t.Fatalf("notifications = %#v, want %#v", notifications, wantNotifications)
	}
	if reportUnit != unit || reportArgs != [3]int32{24, notify, 0} {
		t.Fatalf("report = %p/%v, want %p/[24 %d 0]", reportUnit, reportArgs, unit, notify)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"player":       uintptr(unsafe.Pointer(player)),
			"update":       uintptr(unsafe.Pointer(update)),
			"unit":         uintptr(unsafe.Pointer(unit)),
			"other":        uintptr(unsafe.Pointer(other)),
			"other player": uintptr(unsafe.Pointer(otherPlayer)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(other)
	runtime.KeepAlive(otherPlayer)
}

func TestBeastGuideAwardReportNative4D8000BuildsExactPacket(t *testing.T) {
	s := new(Server)
	player := &Player{PlayerInd: 0xfe}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: 4, UpdateData: unsafe.Pointer(update)}

	var (
		gotRecipient int
		gotPacket    []byte
		gotRelated   *Object
		gotRemove    int
		gotSequence  int
	)
	s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
		gotRecipient = index
		gotPacket = append([]byte(nil), packet...)
		gotRelated = related
		gotRemove = remove
		gotSequence = sequence
		return math.MinInt32
	}

	if got := beastGuideAwardReportNative4D8000(s, unit, 0x123, 0x22, math.MinInt32); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, math.MinInt32)
	}
	want := []byte{byte(netmsg.MSG_REPORT_GUIDE_AWARD), 0x23, 0xa2}
	if gotRecipient != 0xfe || !reflect.DeepEqual(gotPacket, want) || gotRelated != nil || gotRemove != 1 || gotSequence != 1 {
		t.Fatalf("packet = recipient %d bytes %v related %p remove %d sequence %d", gotRecipient, gotPacket, gotRelated, gotRemove, gotSequence)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}

func TestBeastGuideAwardReportNative4D8000NonPlayerReturnsLowPointer(t *testing.T) {
	s := new(Server)
	unit := &Object{ObjClass: 0xf0}
	s.NetSendPacketXxx = func(int, []byte, *Object, int, int) int {
		t.Fatal("non-player report sent a packet")
		return 0
	}
	want := int32(uintptr(unsafe.Pointer(unit)))
	if got := beastGuideAwardReportNative4D8000(s, unit, 1, 1, 0); got != want {
		t.Fatalf("result = %#x, want low pointer %#x", uint32(got), uint32(want))
	}
	runtime.KeepAlive(unit)
}
