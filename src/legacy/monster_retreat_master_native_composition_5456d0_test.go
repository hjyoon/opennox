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

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterRetreatMasterNativeProduction5456D0(t *testing.T) {
	const child = "OPENNOX_RETREAT_MASTER_PRODUCTION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		for _, kind := range []string{"far", "near", "edge", "unspilled-edge", "beyond", "no-owner", "healthy", "anti-magic", "nan-range"} {
			for _, slots := range []int{0, 1, 4} {
				mode := fmt.Sprintf("%s:%d", kind, slots)
				t.Run(mode, func(t *testing.T) {
					cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestMonsterRetreatMasterNativeProduction5456D0$")
					cmd.Env = append(os.Environ(), child+"="+mode)
					out, err := cmd.CombinedOutput()
					t.Logf("%s", out)
					if err != nil || !bytes.Contains(out, []byte("RETREAT_MASTER_REAL_PRODUCTION_PASS")) {
						t.Fatal("actual native C parent and production Push/Pop composition", err)
					}
				})
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 2 {
		t.Fatal("production mode")
	}
	slots, err := strconv.Atoi(parts[1])
	if err != nil || slots != 0 && slots != 1 && slots != 4 {
		t.Fatal("production capacity")
	}
	if retreatMasterRuntimeOverride5456D0 != nil {
		t.Fatal("production path must use real Push/Pop bindings")
	}
	s, unit, update, owner := monsterMainBlockLegacyFixture547210(t)
	unit.PosVec, owner.PosVec = types.Ptf(0, 0), types.Ptf(200, 0)
	unit.ObjOwner = owner
	unit.HealthData.Cur, unit.HealthData.Max = 10, 100
	unit.Buffs = 0
	update.ResumeLevel, update.Field329, update.StatusFlags = 0.5, 20, 0
	for i := range update.AIStack {
		update.AIStack[i] = server.AIStackItem{}
		for arg := range update.AIStack[i].Args {
			value := uintptr(0x5a000000 + i*256 + arg + 1)
			if unsafe.Sizeof(value) == 8 {
				high := uint64(0xface1234)
				value |= uintptr(high << 32)
			}
			update.AIStack[i].Args[arg] = value
		}
	}
	update.AIStackInd = int8(23 - slots)
	update.AIStackHead().Action = uint32(ai.ACTION_RETREAT_TO_MASTER)
	wantMove, wantPop := true, false
	switch parts[0] {
	case "far":
	case "near":
		owner.PosVec, wantMove = types.Ptf(10, 0), false
	case "edge":
		owner.PosVec, wantMove = types.Ptf(50, 0), false
	case "unspilled-edge":
		owner.PosVec, wantMove = types.Ptf(50, 6e-7), false
	case "beyond":
		owner.PosVec = types.Ptf(50, 1e-6)
	case "no-owner":
		unit.ObjOwner, wantMove, wantPop = nil, false, true
	case "healthy":
		unit.HealthData.Cur, wantMove, wantPop = 75, false, true
	case "anti-magic":
		unit.HealthData.Cur, update.StatusFlags, unit.Buffs = 75, 0x20, 0x20000000
	case "nan-range":
		update.Field329 = math.Float32frombits(0x7fc12345)
	default:
		t.Fatal("unknown production branch")
	}
	beforeUnit, beforeUpdate, beforeOwner, beforeHealth := *unit, *update, *owner, *unit.HealthData
	logic, other, frame := s.Rand.Logic.Index(), s.Rand.Other.Index(), s.Frame()
	s.AI.StackChanged = false
	if !s.MonsterActionRetreatToMaster5456D0(unit) {
		t.Fatal("ordinary dispatcher rejected active retreat-to-master")
	}
	wantUnit, wantUpdate, wantOwner, wantHealth := *unit, *update, *owner, *unit.HealthData
	wantChanged := s.AI.StackChanged
	wantHead := beforeUpdate.AIStackInd
	if wantPop {
		wantHead--
	} else if wantMove {
		wantHead += int8(min(slots, 2))
	}
	if update.AIStackInd != wantHead {
		t.Fatal("original action admission/capacity", update.AIStackInd, wantHead)
	}
	if wantMove {
		for index, action := range []ai.ActionType{ai.DEPENDENCY_OBJECT_FARTHER_THAN, ai.ACTION_MOVE_TO}[:min(slots, 2)] {
			item := &update.AIStack[int(beforeUpdate.AIStackInd)+1+index]
			if item.Type() != action || item.ArgObj(2) != owner {
				t.Fatal("original action order or full native owner pointer", item.Type(), action, item.ArgObj(2), owner)
			}
		}
	}
	*unit, *update, *owner, *unit.HealthData = beforeUnit, beforeUpdate, beforeOwner, beforeHealth
	s.AI.StackChanged = false
	retreatMasterOriginalC5456D0(unit)
	for _, pair := range []struct {
		got, want unsafe.Pointer
		size      uintptr
	}{
		{unsafe.Pointer(unit), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
		{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
		{unsafe.Pointer(owner), unsafe.Pointer(&wantOwner), unsafe.Sizeof(wantOwner)},
		{unsafe.Pointer(unit.HealthData), unsafe.Pointer(&wantHealth), unsafe.Sizeof(wantHealth)},
	} {
		if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
			t.Fatal("whole production object/MUD/owner/health differs between raw C and ordinary native dispatcher")
		}
	}
	if s.AI.StackChanged != wantChanged || s.Frame() != frame || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
		t.Fatal("production parent changed AI state, frame or RNG outside original operations")
	}
	t.Logf("RETREAT_MASTER_REAL_PRODUCTION_PASS mode=%s unit=%p owner=%p cachedMUD=%p full-records=true real-PushPop=true", mode, unit, owner, update)
}
