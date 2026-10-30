package legacy

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Compose the sealed original Confused rows with real capability checks,
// logic RNG and native AI-stack pushes through both ordinary and public C
// dispatch. The fixture's declared Push service is replaced with the real
// implementation; its original action/argument result remains the reference.
func TestMonsterConfusedPublicCComposition545140(t *testing.T) {
	var fixture struct {
		Rows []struct {
			Input struct {
				Name               string
				Direction          uint16
				Position           [2]uint32
				FirstRandom        int `json:"first_random"`
				DecisionRandom     int `json:"decision_random"`
				Melee, Shoot, Push int
			}
			Result struct {
				State struct{ Items [2][5]uint32 }
			}
		}
	}
	blob, err := os.ReadFile("../server/testdata/monster_confused_native_original_545140.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	seeds := make(map[[2]int]int)
	for _, first := range []int{0, 14} {
		for _, second := range []int{49, 50} {
			key := [2]int{first, second}
			found := false
			for seed := 0; seed < 2000000; seed++ {
				rng := prand.New(seed)
				// The stock RNG is a fixed cyclic table. Preserve the exact
				// first draw and select a natural second draw in the original
				// 49/50 threshold branch; the sealed component covers the
				// exact boundary values independently.
				if rng.IntClamp(0, 100) == first && (rng.IntClamp(0, 100) < 50) == (second < 50) {
					seeds[key], found = seed, true
					break
				}
			}
			if !found {
				t.Fatal("cannot select natural logic RNG threshold sequence", key)
			}
		}
	}
	count := 0
	for _, row := range fixture.Rows {
		input := row.Input
		if !strings.HasPrefix(input.Name, "capability-branch-") || input.Push == 0 ||
			(input.Direction != 1 && input.Direction != 255) {
			continue
		}
		count++
		t.Run(input.Name, func(t *testing.T) {
			s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
			def, freeDef := alloc.New(server.MonsterDef{})
			t.Cleanup(freeDef)
			unit.ObjSubClass = object.SubClass(0x80)
			unit.Direction1, unit.Direction2 = server.Dir16(input.Direction), 0xbabe
			unit.PosVec = types.Ptf(math.Float32frombits(input.Position[0]), math.Float32frombits(input.Position[1]))
			update.MonsterDef = def
			update.StatusFlags = 0
			if input.Melee != 0 {
				def.MeleeAttackRange112 = 1
			}
			if input.Shoot != 0 {
				def.MissileName148[0] = 'X'
			}
			update.AIStackInd = 0
			if input.Push < 0 {
				update.AIStackInd = int8(len(update.AIStack) - 1)
			}
			update.AIStack[update.AIStackInd] = server.AIStackItem{Action: uint32(ai.ACTION_CONFUSED)}
			update.AIStack[1].Action = 0xa5a50001
			for i := range update.AIStack[1].Args {
				value := uintptr(0x5a000101 + i)
				if unsafe.Sizeof(uintptr(0)) == 8 {
					high := uint64(0xface0000)
					value |= uintptr(high << 32)
				}
				update.AIStack[1].Args[i] = value
			}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, ptr := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(update), unsafe.Pointer(def)} {
					if uintptr(ptr) <= 0xffffffff {
						t.Fatal("native composition must exceed4GiB")
					}
				}
			}
			beforeUnit, beforeUpdate, beforeDef, beforeMissile := *unit, *update, *def, *missile
			seed := seeds[[2]int{input.FirstRandom, input.DecisionRandom}]
			expectRNG := prand.New(seed)
			expectRNG.IntClamp(0, 100)
			if input.Melee != 0 && input.Shoot != 0 {
				expectRNG.IntClamp(0, 100)
			}
			expectNext := expectRNG.IntClamp(0, 100)
			check := func() {
				t.Helper()
				want := row.Result.State.Items[1]
				item := &update.AIStack[1]
				if item.Action != want[0] {
					t.Fatalf("original action=%d got=%d", want[0], item.Action)
				}
				for i, value := range item.Args {
					expected := beforeUpdate.AIStack[1].Args[i]
					if input.Push == 1 && item.Action == uint32(ai.ACTION_MISSILE_ATTACK) && i < 3 {
						expected = uintptr(want[i+1])
					}
					if value != expected {
						t.Fatalf("original argument%d=%#x got=%#x", i, expected, value)
					}
				}
				expectedInd := beforeUpdate.AIStackInd
				pushed := input.Push == 1 && (input.Melee != 0 || input.Shoot != 0)
				if pushed {
					expectedInd++
				}
				if update.AIStackInd != expectedInd || s.AI.StackChanged != pushed {
					t.Fatal("real native stack push/capacity state differs")
				}
				if got := s.Rand.Logic.IntClamp(0, 100); got != expectNext {
					t.Fatalf("original RNG continuation=%d got=%d", expectNext, got)
				}
			}
			s.Rand.Logic, s.AI.StackChanged = prand.New(seed), false
			action, ok := server.GetAIAction(ai.ACTION_CONFUSED).(cgoAIAction)
			if !ok {
				t.Fatal("ordinary Confused dispatcher")
			}
			action.Update(unit)
			check()
			wantUnit, wantUpdate, wantDef := *unit, *update, *def
			*unit, *update, *def = beforeUnit, beforeUpdate, beforeDef
			s.Rand.Logic, s.AI.StackChanged = prand.New(seed), false
			ccall.CallVoidPtr(monsterConfusedPublicC545140(), unit.CObj())
			check()
			for _, pair := range []struct {
				got, want unsafe.Pointer
				size      uintptr
			}{
				{unit.CObj(), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
				{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
				{unsafe.Pointer(def), unsafe.Pointer(&wantDef), unsafe.Sizeof(wantDef)},
				{missile.CObj(), unsafe.Pointer(&beforeMissile), unsafe.Sizeof(beforeMissile)},
			} {
				if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
					t.Fatal("whole native unit/update/definition/missile differs")
				}
			}
			t.Logf("CONFUSED REAL C COMPOSITION PASS: row=%s original-action-arguments=true real-capability-RNG-push=true full-records=true", input.Name)
		})
	}
	if count != 64 {
		t.Fatalf("original composition rows=%d want64", count)
	}
}
