package opennox

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerExecuteAbilityNative4FBB70BindsNativePointersAndWidths(t *testing.T) {
	player := &server.Player{PlayerInd: 7}
	player.SpellLvl[server.AbilityTreadLightly] = 1
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.FlagSelected,
		UpdateData: unsafe.Pointer(update),
	}
	old := &server.ExecAbilityClass{Abil: server.AbilityHarpoon}
	allocated := new(server.ExecAbilityClass)
	head := old
	var cooldown int32
	var cooldownIndices []uint8
	var events []string
	delayCalls := 0

	playerExecuteAbilityNative4FBB70(unit, server.AbilityTreadLightly, playerExecuteAbilityNativeDeps4FBB70{
		gameFlag: func(mask uint32) int32 {
			events = append(events, "game")
			return 0
		},
		isActive: func(got *server.Object, ability server.Ability) int32 {
			if got != unit || ability != server.AbilityTreadLightly {
				t.Fatalf("active args = (%p, %v), want (%p, %v)", got, ability, unit, server.AbilityTreadLightly)
			}
			events = append(events, "active")
			return 0
		},
		isActiveVal: func(*server.Object, server.Ability) int32 {
			t.Fatal("Tread Lightly queried value-active state")
			return 0
		},
		reportText: func(uint8, uint8, *int32) {
			t.Fatal("successful ability reported an error")
		},
		loadCooldown: func(index uint8, ability server.Ability) int32 {
			cooldownIndices = append(cooldownIndices, index)
			if ability != server.AbilityTreadLightly {
				t.Fatalf("cooldown ability = %v", ability)
			}
			return cooldown
		},
		storeCooldown: func(index uint8, ability server.Ability, value int32) {
			cooldownIndices = append(cooldownIndices, index)
			if ability != server.AbilityTreadLightly {
				t.Fatalf("stored cooldown ability = %v", ability)
			}
			cooldown = value
		},
		loadDelay: func(ability server.Ability) int32 {
			if ability != server.AbilityTreadLightly {
				t.Fatalf("delay ability = %v", ability)
			}
			delayCalls++
			if delayCalls == 1 {
				return math.MinInt32 + 9
			}
			return 1
		},
		reportState: func(got *server.Object, ability server.Ability, state uint8) {
			if got != unit || ability != server.AbilityTreadLightly || state != 0 {
				t.Fatalf("state report = (%p, %v, %d)", got, ability, state)
			}
			events = append(events, "state")
		},
		loadDuration: func(ability server.Ability) int32 {
			if ability != server.AbilityTreadLightly {
				t.Fatalf("duration ability = %v", ability)
			}
			return 9
		},
		allocExec: func() *server.ExecAbilityClass {
			return allocated
		},
		loadFrame: func() uint32 {
			return math.MaxUint32 - 4
		},
		loadExecHead: func() *server.ExecAbilityClass {
			return head
		},
		storeExecHead: func(exec *server.ExecAbilityClass) {
			head = exec
		},
		invoke: func(got *server.Object, ability server.Ability) {
			if got != unit || ability != server.AbilityTreadLightly {
				t.Fatalf("invoke args = (%p, %v)", got, ability)
			}
			events = append(events, "invoke")
		},
		loadSound: func(ability server.Ability, slot int32) int32 {
			if ability != server.AbilityTreadLightly || slot != 0 {
				t.Fatalf("sound args = (%v, %d)", ability, slot)
			}
			return 0x12345678
		},
		audio: func(soundID int32, got *server.Object, kind int32, code uint32) {
			if soundID != 0x12345678 || got != unit || kind != 0 || code != 0 {
				t.Fatalf("audio args = (%#x, %p, %d, %#x)", uint32(soundID), got, kind, code)
			}
			events = append(events, "audio")
		},
	})

	if got, want := cooldownIndices, []uint8{7, 7}; !reflect.DeepEqual(got, want) {
		t.Fatalf("cooldown indices = %v, want %v", got, want)
	}
	if got, want := uint32(cooldown), uint32(0x80000009); got != want {
		t.Fatalf("cooldown = %#08x, want %#08x", got, want)
	}
	if delayCalls != 2 {
		t.Fatalf("delay calls = %d, want 2", delayCalls)
	}
	if allocated.Unit != unit || allocated.Abil != server.AbilityTreadLightly || allocated.Frame != 4 || allocated.Active != 1 {
		t.Fatalf("execution record = %+v", allocated)
	}
	if allocated.Next != old || allocated.Prev != nil || old.Prev != allocated || head != allocated {
		t.Fatalf("execution links = new(%p,%p) old.prev=%p head=%p", allocated.Next, allocated.Prev, old.Prev, head)
	}
	if got, want := events[len(events)-3:], []string{"state", "invoke", "audio"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("terminal events = %v, want %v; all %v", got, want, events)
	}
	if got := int32(sound.SoundPermanentFizzle); got != playerExecuteAbilityFizzleSound4FBB70 {
		t.Fatalf("PermanentFizzle = %#x, oracle %#x", uint32(got), uint32(playerExecuteAbilityFizzleSound4FBB70))
	}
}
