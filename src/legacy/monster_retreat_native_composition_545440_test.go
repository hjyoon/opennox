package legacy

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Compose the original TIME/FLEE and empty-food action sequences with real
// RNG, frame/FPS, empty spatial food search, native stack capacity and reused
// argument slots. The independent original fixture establishes those payloads;
// complete ordinary and public-C records must also agree here.
func TestMonsterRetreatPublicCComposition545440(t *testing.T) {
	const child = "NOX_MONSTER_RETREAT_COMPOSITION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		for _, kind := range []string{"flee", "food"} {
			for slots := 0; slots <= 5; slots++ {
				mode := fmt.Sprintf("%s:%d", kind, slots)
				t.Run(mode, func(t *testing.T) {
					bin, err := os.Executable()
					if err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(bin, "-test.run=^TestMonsterRetreatPublicCComposition545440$", "-test.v")
					cmd.Env = append(os.Environ(), child+"="+mode)
					out, err := cmd.CombinedOutput()
					t.Logf("%s", out)
					if err != nil {
						t.Fatalf("actual native Retreat composition: %v", err)
					}
				})
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 2 {
		t.Fatal("composition mode")
	}
	slots, err := strconv.Atoi(parts[1])
	if err != nil || slots < 0 || slots > 5 {
		t.Fatal("composition capacity")
	}
	s, unit, update, enemy := monsterMainBlockLegacyFixture547210(t)
	s.SetFrame(0xfffffffd)
	s.SetTickRate(31)
	unit.HealthData.Cur = 10
	update.ResumeLevel, update.StatusFlags = 0.5, 0
	for i := range update.AIStack {
		update.AIStack[i] = server.AIStackItem{}
		for j := range update.AIStack[i].Args {
			value := uintptr(0x5a000000 + i*256 + j + 1)
			if unsafe.Sizeof(uintptr(0)) == 8 {
				high := uint64(0xface0000)
				value |= uintptr(high << 32)
			}
			update.AIStack[i].Args[j] = value
		}
	}
	update.AIStackInd = int8(len(update.AIStack) - 1 - slots)
	update.AIStackHead().Action = uint32(ai.ACTION_RETREAT)
	update.CurrentEnemy = nil
	wantActions := []ai.ActionType{ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_NO_VISIBLE_FOOD, ai.ACTION_ROAM}
	if parts[0] == "flee" {
		update.CurrentEnemy = enemy
		wantActions = []ai.ActionType{ai.DEPENDENCY_TIME, ai.ACTION_FLEE}
	} else if parts[0] != "food" {
		t.Fatal("composition branch")
	}
	beforeUnit, beforeUpdate, beforeHealth, beforeEnemy := *unit, *update, *unit.HealthData, *enemy
	const seed = 545440
	control := prand.New(seed)
	deadline := uint32(0xfffffffd)
	if parts[0] == "flee" && slots > 0 {
		deadline += uint32(control.IntClamp(124, 186))
	}
	expectedNext := control.IntClamp(0, math.MaxInt32)
	check := func() {
		t.Helper()
		accepted := min(slots, len(wantActions))
		if int(update.AIStackInd) != int(beforeUpdate.AIStackInd)+accepted || s.AI.StackChanged != (accepted > 0) {
			t.Fatal("real stack capacity/change differs")
		}
		for n, action := range wantActions[:accepted] {
			i := int(beforeUpdate.AIStackInd) + 1 + n
			item := &update.AIStack[i]
			if item.Type() != action {
				t.Fatalf("original action=%s got=%s", action, item.Type())
			}
			want := beforeUpdate.AIStack[i].Args
			switch action {
			case ai.DEPENDENCY_TIME:
				want[0] = uintptr(deadline)
			case ai.ACTION_FLEE:
				want[0], want[1], want[2] = uintptr(math.Float32bits(enemy.PosVec.X)), uintptr(math.Float32bits(enemy.PosVec.Y)), 0
			case ai.ACTION_ROAM:
				want[0] = 0
				want[2] = want[2]&^uintptr(255) | 128
			}
			if item.Args != want {
				t.Fatalf("original native arguments action=%s got=%#v want=%#v", action, item.Args, want)
			}
		}
		if got := s.Rand.Logic.IntClamp(0, math.MaxInt32); got != expectedNext {
			t.Fatalf("RNG consumed outside accepted TIME: %d want%d", got, expectedNext)
		}
		if s.Frame() != 0xfffffffd {
			t.Fatal("frame mutated")
		}
	}
	s.Rand.Logic, s.AI.StackChanged = prand.New(seed), false
	action, ok := server.GetAIAction(ai.ACTION_RETREAT).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary Retreat dispatcher")
	}
	action.Update(unit)
	check()
	wantUnit, wantUpdate, wantHealth, wantEnemy := *unit, *update, *unit.HealthData, *enemy
	t.Logf("RETREAT REAL CONTROL COMPLETED: mode=%s unit=%p update=%p health=%p real-RNG-search-push=true", mode, unit, update, unit.HealthData)
	*unit, *update, *unit.HealthData, *enemy = beforeUnit, beforeUpdate, beforeHealth, beforeEnemy
	s.Rand.Logic, s.AI.StackChanged = prand.New(seed), false
	ccall.CallVoidPtr(monsterRetreatPublicC545440(), unit.CObj())
	check()
	for _, pair := range []struct {
		got, want unsafe.Pointer
		size      uintptr
	}{
		{unit.CObj(), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
		{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
		{unsafe.Pointer(unit.HealthData), unsafe.Pointer(&wantHealth), unsafe.Sizeof(wantHealth)},
		{enemy.CObj(), unsafe.Pointer(&wantEnemy), unsafe.Sizeof(wantEnemy)},
	} {
		if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
			t.Fatal("whole native unit/update/health/enemy differs")
		}
	}
	t.Logf("RETREAT REAL C COMPOSITION PASS: mode=%s full-records=true real-RNG-search-push=true", mode)
}
