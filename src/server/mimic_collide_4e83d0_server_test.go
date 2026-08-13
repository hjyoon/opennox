package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMimicCollide4E83D0NativeFieldsArgumentsAndWidths(t *testing.T) {
	mimic := &Object{}
	xBits := uint32(0x7fa12345)
	yBits := uint32(0x80000000)
	other := &Object{
		ObjClass: object.ClassPlayer | object.Class(0x80000000),
		ObjFlags: object.Flags(0x40000000),
		PosVec: types.Ptf(
			math.Float32frombits(xBits),
			math.Float32frombits(yBits),
		),
	}
	allOnes := ^uintptr(0)
	underAttack := &AIStackItem{Args: [4]uintptr{allOnes, allOnes, allOnes, allOnes}}
	fight := &AIStackItem{Args: [4]uintptr{allOnes, allOnes, allOnes, allOnes}}
	collisionWords := [2]uint32{0x3f800000, 0xc0000000}
	collision := unsafe.Pointer(&collisionWords[0])
	resultWord := uint32(0x12345678)
	result := unsafe.Pointer(&resultWord)
	pushes := 0
	frames := []uint32{0xfedcba98, 0x89abcdef}

	got := mimicCollideNative4E83D0(mimic, other, collision, mimicCollideNativeDeps4E83D0{
		isEnemy: func(gotMimic, gotOther *Object) bool {
			if gotMimic != mimic || gotOther != other {
				t.Fatal("native enemy arguments changed")
			}
			return true
		},
		actionScheduled: func(got *Object, action ai.ActionType) bool {
			if got != mimic || action != ai.ACTION_FIGHT {
				t.Fatalf("scheduled args = (%p, %v), want (%p, %v)", got, action, mimic, ai.ACTION_FIGHT)
			}
			return false
		},
		pushAction: func(got *Object, action ai.ActionType) *AIStackItem {
			if got != mimic {
				t.Fatalf("push object = %p, want %p", got, mimic)
			}
			pushes++
			if pushes == 1 {
				if action != ai.DEPENDENCY_UNDER_ATTACK {
					t.Fatalf("first action = %v, want %v", action, ai.DEPENDENCY_UNDER_ATTACK)
				}
				return underAttack
			}
			if action != ai.ACTION_FIGHT {
				t.Fatalf("second action = %v, want %v", action, ai.ACTION_FIGHT)
			}
			return fight
		},
		frame: func() uint32 {
			value := frames[0]
			frames = frames[1:]
			return value
		},
		monsterCollide: func(gotMimic, gotOther *Object, gotCollision unsafe.Pointer) unsafe.Pointer {
			if gotMimic != mimic || gotOther != other || gotCollision != collision {
				t.Fatal("native collision arguments changed")
			}
			return result
		},
	})
	if got != result {
		t.Fatalf("result = %p, want exact %p", got, result)
	}
	if underAttack.Args != [4]uintptr{uintptr(0xfedcba98), allOnes, allOnes, allOnes} {
		t.Fatalf("under-attack args = %#v", underAttack.Args)
	}
	if fight.Args != [4]uintptr{uintptr(xBits), uintptr(yBits), uintptr(0x89abcdef), allOnes} {
		t.Fatalf("fight args = %#v", fight.Args)
	}
	if len(frames) != 0 {
		t.Fatalf("unused frame results = %v", frames)
	}
}

func TestMimicCollide4E83D0NativeDeadUsesNamedFlag(t *testing.T) {
	mimic := &Object{}
	other := &Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagDead}
	resultWord := byte(9)
	result := unsafe.Pointer(&resultWord)
	got := mimicCollideNative4E83D0(mimic, other, nil, mimicCollideNativeDeps4E83D0{
		isEnemy: func(*Object, *Object) bool {
			t.Fatal("enemy called for dead object")
			return false
		},
		actionScheduled: func(*Object, ai.ActionType) bool {
			t.Fatal("scheduled called for dead object")
			return false
		},
		pushAction: func(*Object, ai.ActionType) *AIStackItem {
			t.Fatal("push called for dead object")
			return nil
		},
		frame: func() uint32 {
			t.Fatal("frame called for dead object")
			return 0
		},
		monsterCollide: func(gotMimic, gotOther *Object, gotCollision unsafe.Pointer) unsafe.Pointer {
			if gotMimic != mimic || gotOther != other || gotCollision != nil {
				t.Fatal("final collision arguments changed")
			}
			return result
		},
	})
	if got != result {
		t.Fatalf("result = %p, want exact %p", got, result)
	}
}
