package legacy

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// DelayedDelete is the declared external boundary. The real existing native
// BurnDelete body runs with no owner, disabled coop rewards and unavailable
// scorch type. This does not claim a real deletion queue, XP or scorch spawn.
type dyingEndNativeServices544CA0 struct {
	Server
	deleted []*server.Object
}

func (s *dyingEndNativeServices544CA0) DelayedDelete(unit *server.Object) {
	s.deleted = append(s.deleted, unit)
}

func TestMonsterDyingEndNativeComposition544CA0(t *testing.T) {
	const child = "NOX_MONSTER_DYING_END_COMPOSITION_CHILD"
	if os.Getenv(child) == "" {
		for _, mode := range []string{"ordinary-wait", "ordinary-fire", "zombie-wait", "zombie-fire", "vile-wait", "vile-fire"} {
			t.Run(mode, func(t *testing.T) {
				bin, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(bin, "-test.run=^TestMonsterDyingEndNativeComposition544CA0$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actualC dying-end composition: %v", err)
				}
			})
		}
		return
	}
	mode := os.Getenv(child)
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	view := s.TypePredicateRuntime534A10()
	view.Write(1, 100)
	view.Write(2, 200)
	unit.TypeInd = 300
	query := uint32(0)
	switch mode {
	case "ordinary-wait", "ordinary-fire":
	case "zombie-wait", "zombie-fire":
		unit.TypeInd = 100
		query = 1
	case "vile-wait", "vile-fire":
		unit.TypeInd = 200
		query = 1
	default:
		t.Fatal("unknown frozen composition")
	}
	unit.ObjSubClass = object.SubClass(0xa5)
	update.StatusFlags = 0
	if mode == "ordinary-fire" || mode == "zombie-fire" || mode == "vile-fire" {
		update.StatusFlags = object.MonStatusOnFire
	}
	if uint32(zombieEntryCProbe534A10(1, unit)) != query {
		t.Fatal("real native C IsZombie predicate")
	}
	if s.Types.ByID("ScorchMarkFloorMediumA") != nil || unit.ObjOwner != nil || unit.Obj130 != nil {
		t.Fatal("frozen no-owner, unavailable scorch and attribution boundaries")
	}
	path := "../server/testdata/monster_dying_end_native_original_544ca0.json"
	if p := os.Getenv("NOX_MONSTER_DYING_END_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Rows []struct {
			Input struct {
				Name   string `json:"name"`
				Zombie uint32 `json:"zombie"`
				Flags  uint32 `json:"flags"`
			} `json:"input"`
			Result struct {
				Events []struct {
					Kind string `json:"kind"`
				} `json:"events"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	found, wantBurn := false, false
	for _, row := range f.Rows {
		if len(row.Input.Name) >= 16 && row.Input.Name[:16] == "predicate-flags-" && row.Input.Zombie == query && row.Input.Flags == uint32(update.StatusFlags) {
			if found {
				t.Fatal("ambiguous frozen original context")
			}
			found = true
			if len(row.Result.Events) == 2 && row.Result.Events[1].Kind == "burn" {
				wantBurn = true
			} else if len(row.Result.Events) != 1 || row.Result.Events[0].Kind != "zombie" {
				t.Fatal("original service sequence")
			}
		}
	}
	if !found {
		t.Fatal("missing frozen original context")
	}
	native := &dyingEndNativeServices544CA0{Server: GetServer()}
	previous := GetServer
	GetServer = func() Server { return native }
	defer func() { GetServer = previous }()
	a, ok := server.GetAIAction(ai.ACTION_DYING).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary native dying dispatcher")
	}
	beforeUnit, beforeUpdate, beforeMissile := *unit, *update, *missile
	a.End(unit)
	checkDeletion := func() {
		if wantBurn {
			if len(native.deleted) != 1 || native.deleted[0] != unit {
				t.Fatal("original burn must reach native deletion boundary once", native.deleted)
			}
		} else if len(native.deleted) != 0 {
			t.Fatal("original wait unexpectedly deleted", native.deleted)
		}
	}
	checkDeletion()
	rawUnit := func(v *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	rawUpdate := func(v *server.MonsterUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))
	}
	checkRecords := func() {
		if !bytes.Equal(rawUnit(unit), rawUnit(&beforeUnit)) || !bytes.Equal(rawUpdate(update), rawUpdate(&beforeUpdate)) || !bytes.Equal(rawUnit(missile), rawUnit(&beforeMissile)) {
			t.Fatal("full native object/update/missile records changed")
		}
	}
	checkRecords()
	t.Logf("DYING END COMPOSITION ORDINARY COMPLETED: mode=%s original-burn=%t unit=%p update=%p", mode, wantBurn, unit, update)
	native.deleted = nil
	ccall.CallVoidPtr(monsterDyingEndPublicC544CA0(), unit.CObj())
	checkDeletion()
	checkRecords()
	t.Logf("DYING END COMPOSITION NATIVE C PASS: mode=%s original-burn=%t real-IsZombie=true real-native-BurnDelete=%t external-delete-boundary=true full-records=true", mode, wantBurn, wantBurn)
}
