package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterSearchEdibleNativeProduction544A00(t *testing.T) {
	const child = "OPENNOX_EDIBLE_PRODUCTION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		for entry := 0; entry < 3; entry++ {
			for kind := 0; kind < 10; kind++ {
				indices := []int{0}
				if entry == 0 {
					indices = []int{0, 11, 18}
				}
				for _, index := range indices {
					mode := fmt.Sprintf("%d:%d:%d", entry, kind, index)
					t.Run(mode, func(t *testing.T) {
						cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestMonsterSearchEdibleNativeProduction544A00$")
						cmd.Env = append(os.Environ(), child+"="+mode)
						out, err := cmd.CombinedOutput()
						t.Logf("%s", out)
						if err != nil || !bytes.Contains(out, []byte("EDIBLE_REAL_PRODUCTION_PASS")) {
							t.Fatal("real C map/visibility/flags/Push composition", err)
						}
					})
				}
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 3 {
		t.Fatal("mode")
	}
	entry, e1 := strconv.Atoi(parts[0])
	kind, e2 := strconv.Atoi(parts[1])
	index, e3 := strconv.Atoi(parts[2])
	if e1 != nil || e2 != nil || e3 != nil || entry < 0 || entry > 2 || kind < 0 || kind > 9 || (index != 0 && index != 11 && index != 18) {
		t.Fatal("mode values")
	}
	TestMonsterSearchEdibleNativeLayout544A00(t)
	if edibleOriginalABI544A00(entry) != 64 {
		t.Fatal("typed native ABI required")
	}
	if edibleNativeRuntimeOverride544A00 != nil {
		t.Fatal("production must use real engine bindings")
	}
	oldScratch := edibleNativeScratch544A00
	defer func() { edibleNativeScratch544A00 = oldScratch }()
	oldFlags := noxflags.GetGame()
	defer func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) }()
	s, unit, update, owner := monsterMainBlockLegacyFixture547210(t)
	food, freeFood := alloc.New(server.Object{})
	defer freeFood()
	*food = server.Object{ObjClass: object.ClassFood, ObjFlags: object.FlagActive | object.FlagEnabled, PosVec: types.Ptf(320, 300), NewPos: types.Ptf(320, 300)}
	food.Shape.Kind = server.ShapeKindCenter
	flags := noxflags.GameFlag(0)
	found := true
	switch kind {
	case 0:
	case 1:
		flags = noxflags.GameModeQuest
	case 2:
		food.ObjSubClass = object.SubClass(0x80)
		unit.Poison540 = 255
	case 3:
		food.ObjSubClass = object.SubClass(0x80)
		found = false
	case 4:
		food.ObjSubClass = object.SubClass(0x18)
		flags = noxflags.GameOnline
		unit.ObjSubClass = object.SubClass(0x10)
	case 5:
		food.ObjSubClass = object.SubClass(0x18)
		unit.ObjSubClass = object.SubClass(0x10)
		found = false
	case 6:
		food.ObjSubClass = object.SubClass(0x18)
		flags = noxflags.GameOnline
		found = false
	case 7:
		food.ObjSubClass = object.SubClass(4)
		found = false
	case 8:
		food.ObjClass = object.ClassMonster
		found = false
	case 9:
		food.ObjSubClass = object.SubClass(8)
		flags = noxflags.GameOnline
		unit.ObjSubClass = object.SubClass(0x10)
		found = false
	}
	noxflags.ResetGame()
	noxflags.SetGame(flags)
	s.Map.AddObjectToIndex(food)
	defer s.Map.RemoveObjectFromIndex(food)
	unit.HealthData.Cur, unit.HealthData.Max = 10, 100
	update.ResumeLevel, update.CurrentEnemy = 0.5, nil
	update.SoundSet122 = unsafe.Pointer(owner)
	update.AIStackInd = int8(index)
	for i := range update.AIStack {
		update.AIStack[i].Action = uint32(ai.ACTION_RETREAT)
		update.AIStack[i].Field5 = 0xa5000000 + uint32(i)
		for j := range update.AIStack[i].Args {
			update.AIStack[i].Args[j] = uintptr(unsafe.Pointer(owner)) + uintptr(i*8+j)
		}
	}
	records := []struct {
		ptr  unsafe.Pointer
		size uintptr
	}{{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, {unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(owner), unsafe.Sizeof(*owner)}, {unsafe.Pointer(food), unsafe.Sizeof(*food)}, {unsafe.Pointer(unit.HealthData), unsafe.Sizeof(*unit.HealthData)}}
	for _, r := range records {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(r.ptr) <= 0xffffffff {
			t.Fatal("native production pointer must exceed4GiB")
		}
	}
	snapshot := func() [][]byte {
		out := make([][]byte, len(records))
		for i, r := range records {
			out[i] = bytes.Clone(unsafe.Slice((*byte)(r.ptr), int(r.size)))
		}
		return out
	}
	// A search visits at level zero; its accepted visibility check visits
	// at level one. A direct raw callback's visibility runs at level zero.
	// Check those exact service stores while preserving every other byte.
	nextVisit := func(in [][]byte, search bool) [][]byte {
		out := make([][]byte, len(in))
		for i := range in {
			out[i] = bytes.Clone(in[i])
		}
		offset := int(unsafe.Offsetof(food.Field62))
		if search || found {
			value := binary.LittleEndian.Uint32(out[3][offset:])
			binary.LittleEndian.PutUint32(out[3][offset:], value+1)
		}
		if search && found {
			value := binary.LittleEndian.Uint32(out[3][offset+4:])
			binary.LittleEndian.PutUint32(out[3][offset+4:], value+1)
		}
		return out
	}
	before := snapshot()
	restore := func() {
		for i, r := range records {
			copy(unsafe.Slice((*byte)(r.ptr), int(r.size)), before[i])
		}
	}
	check := func(want [][]byte) {
		t.Helper()
		for i, r := range records {
			if !bytes.Equal(want[i], unsafe.Slice((*byte)(r.ptr), int(r.size))) {
				for offset, got := range unsafe.Slice((*byte)(r.ptr), int(r.size)) {
					if got != want[i][offset] {
						t.Fatalf("whole native production record %d offset=%#x got=%02x want=%02x food-stamps=%v", i, offset, got, want[i][offset], food.Field62)
					}
				}
			}
		}
	}
	expected := food
	if !found {
		expected = nil
	}
	if got := s.MonsterSearchEdible544A00(unit, 250); got != expected {
		t.Fatal("ordinary real native food control", got, expected)
	}
	check(nextVisit(before, true))
	before = snapshot()
	frame, logic, other, changed := s.Frame(), s.Rand.Logic.Index(), s.Rand.Other.Index(), s.AI.StackChanged
	edibleNativeScratch544A00 = server.MonsterEdibleSearchState544A00{BestDistance: 10000000, BestObject: owner}
	result := edibleOriginalC544A00(entry, unit, food, 250)
	if !bytes.Equal(nextVisit(before, entry != 2)[3], unsafe.Slice((*byte)(unsafe.Pointer(food)), int(unsafe.Sizeof(*food)))) {
		t.Fatalf("real search/visibility must change only predicted visit stamps: mode=%s food-stamps=%v", mode, food.Field62)
	}
	if entry == 0 {
		actions := []ai.ActionType{ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_NO_VISIBLE_FOOD, ai.ACTION_ROAM}
		if found {
			actions = []ai.ActionType{ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO}
		}
		if int(update.AIStackInd) != index+len(actions) || result != unsafe.Pointer(&update.AIStack[index+len(actions)]) {
			t.Fatal("full native Push return/index", result, update.AIStackInd)
		}
		for i, a := range actions {
			if update.AIStack[index+1+i].Type() != a {
				t.Fatal("Push order", i)
			}
		}
		if found {
			if update.AIStack[index+3].ArgObj(2) != food || update.AIStack[index+4].ArgObj(0) != food || update.AIStack[index+5].ArgObj(2) != food {
				t.Fatal("full native food payload")
			}
		} else {
			last := &update.AIStack[index+4]
			if last.Args[0] != 0 || last.Args[2]&0xff != 0x80 {
				t.Fatal("ROAM partial-byte payload")
			}
		}
		after, afterChanged := snapshot(), s.AI.StackChanged
		restore()
		s.AI.StackChanged = changed
		s.MonsterActionRetreat545440(unit)
		check(nextVisit(after, true))
		if s.AI.StackChanged != afterChanged {
			t.Fatal("ordinary/raw stack-change mismatch")
		}
	} else {
		check(nextVisit(before, entry == 1))
		if entry == 1 && result != unsafe.Pointer(expected) {
			t.Fatal("full native selected return", result, expected)
		}
		want := expected
		if entry == 2 && !found {
			want = owner
		}
		if edibleNativeScratch544A00.BestObject != want {
			t.Fatal("native scratch selection identity", edibleNativeScratch544A00.BestObject, want)
		}
		if s.AI.StackChanged != changed {
			t.Fatal("search/callback changed stack state")
		}
	}
	if s.Frame() != frame || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || noxflags.GetGame() != flags {
		t.Fatal("frame/RNG/game flags changed")
	}
	t.Logf("EDIBLE_REAL_PRODUCTION_PASS mode=%s unit=%p food=%p update=%p all-five-whole-records=true flags-poison-map-visibility-real=true", mode, unit, food, update)
}
