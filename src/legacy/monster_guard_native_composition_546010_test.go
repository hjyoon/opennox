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

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Pair the independent original action/payload contracts with real native
// type predicates, RNG, frame/FPS, ordinary dispatch and stack capacity.
// No production service is replaced by a successful callback.
func TestMonsterGuardPublicCComposition546010(t *testing.T) {
	const child = "NOX_MONSTER_GUARD_COMPOSITION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		for _, kind := range []string{"fight", "plant", "moderate", "move", "mimic-angle", "seen-player", "previous-location", "recent-sound"} {
			for slots := 0; slots <= 5; slots++ {
				mode := fmt.Sprintf("%s:%d", kind, slots)
				t.Run(mode, func(t *testing.T) {
					bin, err := os.Executable()
					if err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(bin, "-test.run=^TestMonsterGuardPublicCComposition546010$", "-test.v")
					cmd.Env = append(os.Environ(), child+"="+mode)
					out, err := cmd.CombinedOutput()
					t.Logf("%s", out)
					if err != nil {
						t.Fatalf("actual native Guard composition: %v", err)
					}
				})
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 2 {
		t.Fatal("Guard composition mode")
	}
	slots, err := strconv.Atoi(parts[1])
	if err != nil || slots < 0 || slots > 5 {
		t.Fatal("Guard composition capacity")
	}
	s, unit, update, enemy := monsterMainBlockLegacyFixture547210(t)
	s.Types.Free()
	for _, name := range []string{"GuardNativeOrdinary", "CarnivorousPlant", "Mimic"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	unit.TypeInd = uint16(s.Types.IndByID("GuardNativeOrdinary"))
	unit.NetCode, unit.Buffs = 0, 1<<29
	unit.PrevPos = unit.PosVec
	s.SetFrame(1408)
	s.SetTickRate(31)
	update.Aggression, update.SightRange, update.StatusFlags = 0.5, 500, 0
	update.CurrentEnemy = nil
	for i := range update.AIStack {
		update.AIStack[i] = server.AIStackItem{Field5: 0x13572468}
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
	head := update.AIStackHead()
	head.Action = uint32(ai.ACTION_GUARD)
	head.SetArgs(unit.PosVec, uint32(0))
	wantActions := []ai.ActionType{}
	first, freeFirst := alloc.New(server.Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagEnabled, PosVec: types.Ptf(340, 300)})
	t.Cleanup(freeFirst)
	switch parts[0] {
	case "fight":
		update.Aggression, update.CurrentEnemy = 1, enemy
		wantActions = []ai.ActionType{ai.ACTION_FIGHT}
	case "plant":
		unit.TypeInd = uint16(s.Types.IndByID("CarnivorousPlant"))
		update.CurrentEnemy = enemy
		wantActions = []ai.ActionType{ai.DEPENDENCY_ENEMY_CLOSER_THAN, ai.ACTION_FIGHT}
	case "moderate":
		update.CurrentEnemy = enemy
		wantActions = []ai.ActionType{ai.DEPENDENCY_UNDER_ATTACK, ai.DEPENDENCY_LOCATION_CLOSER_THAN, ai.DEPENDENCY_OR, ai.ACTION_FIGHT}
	case "move":
		update.Aggression = 1
		head.SetArgs(types.Ptf(400, 300), uint32(0))
		wantActions = []ai.ActionType{ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_NO_INTERESTING_SOUND, ai.DEPENDENCY_NOT_UNDER_ATTACK, ai.ACTION_MOVE_TO}
	case "mimic-angle":
		unit.TypeInd = uint16(s.Types.IndByID("Mimic"))
		head.SetArgs(unit.PosVec, uint32(128))
		wantActions = []ai.ActionType{ai.ACTION_FACE_ANGLE}
	case "seen-player":
		enemy.ObjClass, enemy.PosVec = object.ClassPlayer, types.Ptf(260, 300)
		update.Field282_1 = 2
		update.SeenEnemies[0], update.SeenEnemies[1] = first, enemy
		wantActions = []ai.ActionType{ai.ACTION_FACE_OBJECT}
	case "previous-location":
		s.SetFrame(1409)
		update.Field137 = 0
		unit.PrevPos = types.Ptf(310, 320)
		wantActions = []ai.ActionType{ai.ACTION_FACE_LOCATION}
	case "recent-sound":
		s.SetFrame(0xfffffffd)
		update.Field97, update.Field101 = 7, 0xfffffffc
		update.Field99X, update.Field99Y = 12.5, -8.25
		wantActions = []ai.ActionType{ai.DEPENDENCY_NO_INTERESTING_SOUND, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.ACTION_WAIT, ai.ACTION_FACE_LOCATION}
	default:
		t.Fatal("Guard composition branch")
	}
	if s.IsMimic(unit) != (parts[0] == "mimic-angle") || s.IsPlant(unit) != (parts[0] == "plant") {
		t.Fatal("real native type predicates disagree")
	}
	beforeUnit, beforeUpdate, beforeHealth, beforeEnemy, beforeFirst := *unit, *update, *unit.HealthData, *enemy, *first
	const seed = 546010
	control := prand.New(seed)
	deadline := s.Frame()
	if parts[0] == "recent-sound" && slots >= 3 {
		deadline += uint32(control.IntClamp(31, 62))
	}
	expectedNext := control.IntClamp(0, math.MaxInt32)
	check := func() {
		t.Helper()
		accepted := min(slots, len(wantActions))
		if int(update.AIStackInd) != int(beforeUpdate.AIStackInd)+accepted || s.AI.StackChanged != (accepted > 0) {
			t.Fatalf("real native Guard capacity/change differs: index%d expected%d changed%t", update.AIStackInd, int(beforeUpdate.AIStackInd)+accepted, s.AI.StackChanged)
		}
		for n, action := range wantActions[:accepted] {
			i := int(beforeUpdate.AIStackInd) + 1 + n
			item := &update.AIStack[i]
			if item.Type() != action {
				t.Fatalf("Guard action got%s original%s", item.Type(), action)
			}
			want := beforeUpdate.AIStack[i].Args
			switch action {
			case ai.ACTION_FIGHT:
				want[0], want[1], want[2] = uintptr(math.Float32bits(beforeEnemy.PosVec.X)), uintptr(math.Float32bits(beforeEnemy.PosVec.Y)), uintptr(s.Frame())
			case ai.DEPENDENCY_ENEMY_CLOSER_THAN:
				want[0] = uintptr(math.Float32bits(524.9999389648438))
			case ai.DEPENDENCY_UNDER_ATTACK:
				want[0] = 0
			case ai.DEPENDENCY_LOCATION_CLOSER_THAN:
				want[0], want[2], want[3] = uintptr(math.Float32bits(750)), uintptr(math.Float32bits(300)), uintptr(math.Float32bits(300))
			case ai.ACTION_MOVE_TO:
				want[0], want[1], want[2] = uintptr(math.Float32bits(400)), uintptr(math.Float32bits(300)), 0
			case ai.ACTION_FACE_ANGLE:
				want[0] = 128
			case ai.ACTION_FACE_OBJECT:
				want[0] = uintptr(enemy.CObj())
			case ai.ACTION_WAIT:
				want[0] = uintptr(deadline)
			case ai.ACTION_FACE_LOCATION:
				position := beforeUnit.PrevPos
				if parts[0] == "recent-sound" {
					position = types.Ptf(beforeUpdate.Field99X, beforeUpdate.Field99Y)
				}
				want[0], want[1] = uintptr(math.Float32bits(position.X)), uintptr(math.Float32bits(position.Y))
			}
			if item.Args != want {
				t.Fatalf("original Guard payload action%s got%#v want%#v", action, item.Args, want)
			}
		}
		if parts[0] == "recent-sound" && update.Field97 != 0 {
			t.Fatal("original recent-sound marker was not consumed")
		}
		if got := s.Rand.Logic.IntClamp(0, math.MaxInt32); got != expectedNext {
			t.Fatalf("Guard RNG gate got%d want%d", got, expectedNext)
		}
	}
	s.Rand.Logic, s.AI.StackChanged = prand.New(seed), false
	action, ok := server.GetAIAction(ai.ACTION_GUARD).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary Guard dispatcher")
	}
	action.Update(unit)
	check()
	wantUnit, wantUpdate, wantHealth, wantEnemy, wantFirst := *unit, *update, *unit.HealthData, *enemy, *first
	t.Logf("GUARD REAL CONTROL COMPLETED: mode=%s unit=%p update=%p real-RNG-types-push=true", mode, unit, update)
	*unit, *update, *unit.HealthData, *enemy, *first = beforeUnit, beforeUpdate, beforeHealth, beforeEnemy, beforeFirst
	s.Rand.Logic, s.AI.StackChanged = prand.New(seed), false
	ccall.CallVoidPtr(monsterGuardPublicC546010(), unit.CObj())
	check()
	for _, pair := range []struct {
		got, want unsafe.Pointer
		size      uintptr
	}{
		{unit.CObj(), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
		{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
		{unsafe.Pointer(unit.HealthData), unsafe.Pointer(&wantHealth), unsafe.Sizeof(wantHealth)},
		{enemy.CObj(), unsafe.Pointer(&wantEnemy), unsafe.Sizeof(wantEnemy)},
		{first.CObj(), unsafe.Pointer(&wantFirst), unsafe.Sizeof(wantFirst)},
	} {
		if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
			t.Fatal("whole native Guard composition records differ")
		}
	}
	t.Logf("GUARD REAL C COMPOSITION PASS: mode=%s full-records=true real-RNG-types-push=true", mode)
}
