package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestSoloMonsterKillReward4EE500NativeLayout(t *testing.T) {
	wantSize := uintptr(780)
	wantClass := uintptr(8)
	wantExperience := uintptr(28)
	wantOwner := uintptr(508)
	wantAttribution := uintptr(520)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantClass = 12
		wantExperience = 32
		wantOwner = 552
		wantAttribution = 576
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.Experience", unsafe.Offsetof(Object{}.Experience), wantExperience},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.Obj130", unsafe.Offsetof(Object{}.Obj130), wantAttribution},
	} {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSoloMonsterKillRewardNative4EE500BindsFieldsAndLiveOwner(t *testing.T) {
	player := &Object{ObjClass: object.ClassPlayer}
	second := &Object{ObjClass: object.ClassMonster}
	monster := &Object{ObjClass: object.ClassMonster, ObjOwner: player}
	killed := &Object{Obj130: monster, Experience: math.Float32frombits(0x41480000)}

	monitoredCalls := 0
	giveCalls := 0
	stringCalls := 0
	sendCalls := 0
	soloMonsterKillRewardNative4EE500(killed, soloMonsterKillRewardNativeDeps4EE500{
		gameFlag: func(flag uint32) int32 {
			if flag != uint32(noxflags.GameModeCoop) {
				t.Fatalf("game flag = %08x, want %08x", flag, uint32(noxflags.GameModeCoop))
			}
			return -1
		},
		findParent: (*Object).FindOwnerChainPlayer,
		isMonitored: func(gotPlayer, gotMonster *Object) bool {
			monitoredCalls++
			if gotPlayer != player || gotMonster != monster {
				t.Fatalf("monitored args = %p/%p, want %p/%p", gotPlayer, gotMonster, player, monster)
			}
			monster.ObjOwner = second
			return true
		},
		giveXP: func(gotPlayer *Object, experience float32) float64 {
			giveCalls++
			if gotPlayer != player || math.Float32bits(experience) != 0x41480000 {
				t.Fatalf("give args = %p/%08x, want %p/41480000", gotPlayer, math.Float32bits(experience), player)
			}
			return float64(uint64(1)<<32) + 9.75
		},
		loadString: func(key, path string, line int) string {
			stringCalls++
			if key != soloMonsterKillRewardMessageKey4EE500 ||
				path != soloMonsterKillRewardMessagePath4EE500 ||
				line != soloMonsterKillRewardMessageLine4EE500 {
				t.Fatalf("string provenance = %q/%q/%d", key, path, line)
			}
			return "points=%d"
		},
		sendLineMessage: func(gotPlayer *Object, message string, points uint32) {
			sendCalls++
			if gotPlayer != player || message != "points=%d" || points != 9 {
				t.Fatalf("send args = %p/%q/%d, want %p/points=%%d/9", gotPlayer, message, points, player)
			}
		},
	})
	if monitoredCalls != 1 || giveCalls != 1 || stringCalls != 1 || sendCalls != 1 {
		t.Fatalf("calls = monitored:%d give:%d string:%d send:%d, want all 1", monitoredCalls, giveCalls, stringCalls, sendCalls)
	}
}

func TestSoloMonsterKillRewardNative4EE500PreservesNilParentFault(t *testing.T) {
	killed := &Object{Obj130: &Object{}}
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil parent class panic")
		}
	}()
	soloMonsterKillRewardNative4EE500(killed, soloMonsterKillRewardNativeDeps4EE500{
		gameFlag:   func(uint32) int32 { return 1 },
		findParent: func(*Object) *Object { return nil },
		isMonitored: func(*Object, *Object) bool {
			t.Fatal("monitored called after nil parent")
			return false
		},
		giveXP: func(*Object, float32) float64 {
			t.Fatal("give XP called after nil parent")
			return 0
		},
		loadString: func(string, string, int) string {
			t.Fatal("string loaded after nil parent")
			return ""
		},
		sendLineMessage: func(*Object, string, uint32) {
			t.Fatal("message sent after nil parent")
		},
	})
}

func TestSoloMonsterKillReward4EE500ServerBindingUsesCoopGate(t *testing.T) {
	oldFlags := noxflags.GetGame()
	defer noxflags.SetGame(oldFlags)
	noxflags.SetGame(oldFlags &^ noxflags.GameModeCoop)

	killed := &Object{Obj130: &Object{}}
	s := &Server{}
	s.SoloMonsterKillReward4EE500(killed, SoloMonsterKillRewardRuntime4EE500{
		GiveXP: func(*Object, float32) float64 {
			t.Fatal("GiveXP called outside cooperative mode")
			return 0
		},
		SendLineMessage: func(*Object, string, uint32) {
			t.Fatal("SendLineMessage called outside cooperative mode")
		},
	})
}
