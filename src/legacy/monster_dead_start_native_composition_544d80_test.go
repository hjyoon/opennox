package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type deadStartCompositionRow544D80 struct {
	Input struct {
		Name         string   `json:"name"`
		Field131     uint32   `json:"field131"`
		Subclass     uint32   `json:"subclass"`
		Flags        uint32   `json:"flags"`
		Definition   int      `json:"definition"`
		Callback     int      `json:"callback"`
		Zombie       []uint32 `json:"zombie"`
		RandomReturn uint32   `json:"random_return"`
		Mutation     struct {
			Stage string `json:"stage"`
		} `json:"mutation"`
	} `json:"input"`
	Result struct {
		ReturnByte uint8 `json:"return_byte"`
		Events     []struct {
			Kind string  `json:"kind"`
			Args []int32 `json:"args"`
		} `json:"events"`
		State struct {
			Flags     uint32    `json:"flags"`
			Motion    [6]uint32 `json:"motion"`
			Durations [2]uint32 `json:"durations"`
		} `json:"state"`
	} `json:"result"`
}

func TestMonsterDeadStartNativeComposition544D80(t *testing.T) {
	const child = "NOX_MONSTER_DEAD_START_COMPOSITION_CHILD"
	if os.Getenv(child) == "" {
		for _, kind := range []string{"ordinary", "soul", "zombie", "vile"} {
			for _, low := range []int{0, 128, 255} {
				mode := fmt.Sprintf("%s-%d", kind, low)
				t.Run(mode, func(t *testing.T) {
					bin, err := os.Executable()
					if err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(bin, "-test.run=^TestMonsterDeadStartNativeComposition544D80$", "-test.v")
					cmd.Env = append(os.Environ(), child+"="+mode)
					out, err := cmd.CombinedOutput()
					t.Logf("%s", out)
					if err != nil {
						t.Fatalf("actualC DeadStart composition: %v", err)
					}
				})
			}
		}
		return
	}
	mode := os.Getenv(child)
	parts := strings.Split(mode, "-")
	if len(parts) != 2 {
		t.Fatal("invalid composition mode")
	}
	low, err := strconv.ParseUint(parts[1], 10, 8)
	if err != nil {
		t.Fatal(err)
	}
	kind := parts[0]
	zombie := kind == "zombie" || kind == "vile"
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	view := s.TypePredicateRuntime534A10()
	view.Write(1, 100)
	view.Write(2, 200)
	unit.TypeInd = 300
	if kind == "zombie" {
		unit.TypeInd = 100
	}
	if kind == "vile" {
		unit.TypeInd = 200
	}
	unit.Field131 = 0
	unit.ObjSubClass = 0
	if kind == "soul" {
		unit.Field131 = 14
		unit.ObjSubClass = 0x10000
	}
	if s.Types.ByID("ReleasedSoul") != nil {
		t.Fatal("declared unavailable soul allocation boundary")
	}
	def, freeDef := alloc.New(server.MonsterDef{})
	defer freeDef()
	*def = server.MonsterDef{}
	update.MonsterDef = def
	update.Field123 = 0xcccc1111
	unit.ObjFlags = object.Flags(low)
	unit.VelVec, unit.ForceVec, unit.Pos24 = types.Ptf(1, 2), types.Ptf(3, 4), types.Ptf(5, 6)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("ZombieDeadDuration: [7.900000095367431640625, 12.8999996185302734375]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	oldData := datapath.Data()
	datapath.SetData(dir)
	defer datapath.SetData(oldData)
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	if math.Float32bits(float32(s.Balance.FloatInd("ZombieDeadDuration", 0))) != 0x40fccccd || math.Float32bits(float32(s.Balance.FloatInd("ZombieDeadDuration", 1))) != 0x414e6666 {
		t.Fatal("real balance load differs from frozen binary32 inputs")
	}
	path := "../server/testdata/monster_dead_start_native_original_544d80.json"
	if p := os.Getenv("NOX_MONSTER_DEAD_START_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA     string                          `json:"original_sha256"`
		Control uint16                          `json:"control_word"`
		Rows    []deadStartCompositionRow544D80 `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Control != 0xe3f || len(f.Rows) != 630 {
		t.Fatal("composition original identity/profile/count")
	}
	seed := 0
	if kind == "vile" {
		seed = 511
	}
	predict := prand.New(seed)
	wantedRandom := uint32(0)
	if zombie {
		var bounds []int32
		for _, row := range f.Rows {
			if strings.HasPrefix(row.Input.Name, "native-duration-alias-") {
				for _, e := range row.Result.Events {
					if e.Kind == "random" {
						bounds = e.Args
					}
				}
				break
			}
		}
		if len(bounds) != 2 || bounds[0] != 7 || bounds[1] != 12 {
			t.Fatal("frozen original RNG bounds")
		}
		wantedRandom = uint32(predict.IntClamp(int(bounds[0]), int(bounds[1])))
	}
	var ref *deadStartCompositionRow544D80
	for i := range f.Rows {
		row := &f.Rows[i]
		if row.Input.Field131 != unit.Field131 || row.Input.Subclass != uint32(unit.ObjSubClass) || row.Input.Flags != uint32(unit.ObjFlags) || row.Input.Callback != 0 || row.Input.Definition != 0 || row.Input.Mutation.Stage != "" {
			continue
		}
		if zombie {
			if !strings.HasPrefix(row.Input.Name, "native-duration-alias-") || row.Input.RandomReturn != wantedRandom || len(row.Input.Zombie) != 1 || row.Input.Zombie[0] != 1 {
				continue
			}
		} else {
			if !strings.HasPrefix(row.Input.Name, "branch-flags-") || len(row.Input.Zombie) != 2 || row.Input.Zombie[0] != 0 || row.Input.Zombie[1] != 0 {
				continue
			}
		}
		if ref != nil {
			t.Fatal("ambiguous frozen original composition")
		}
		ref = row
	}
	if ref == nil {
		t.Fatal("missing frozen original composition")
	}
	if uint32(zombieEntryCProbe534A10(1, unit)) != map[bool]uint32{false: 0, true: 1}[zombie] {
		t.Fatal("real C IsZombie predicate")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(unsafe.Pointer(def)) <= 0xffffffff) {
		t.Fatal("native C-owned object/update/definition must exceed4GiB")
	}
	beforeUnit, beforeUpdate, beforeDef, beforeMissile := *unit, *update, *def, *missile
	beforeChanged := s.AI.StackChanged
	wantUnit, wantUpdate := beforeUnit, beforeUpdate
	wantUnit.ObjFlags = object.Flags(ref.Result.State.Flags)
	motion := ref.Result.State.Motion
	wantUnit.VelVec = types.Ptf(math.Float32frombits(motion[0]), math.Float32frombits(motion[1]))
	wantUnit.ForceVec = types.Ptf(math.Float32frombits(motion[2]), math.Float32frombits(motion[3]))
	wantUnit.Pos24 = types.Ptf(math.Float32frombits(motion[4]), math.Float32frombits(motion[5]))
	wantUpdate.Field123 = ref.Result.State.Durations[0]
	rawUnit := func(v *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	rawUpdate := func(v *server.MonsterUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))
	}
	rawDef := func(v *server.MonsterDef) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))
	}
	matches := func() bool {
		return bytes.Equal(rawUnit(unit), rawUnit(&wantUnit)) && bytes.Equal(rawUpdate(update), rawUpdate(&wantUpdate)) && bytes.Equal(rawDef(def), rawDef(&beforeDef)) && bytes.Equal(rawUnit(missile), rawUnit(&beforeMissile)) && s.AI.StackChanged == beforeChanged && s.Rand.Logic.Index() == predict.Index()
	}
	a, ok := server.GetAIAction(ai.ACTION_DEAD).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary native dead dispatcher")
	}
	s.Rand.Logic.Reset(seed)
	a.Start(unit)
	ordinaryMatch := matches()
	t.Logf("DEAD START COMPOSITION ORDINARY COMPLETED: mode=%s original-match=%t duration=%d want=%d rng-index=%d want-index=%d unit=%p update=%p def=%p", mode, ordinaryMatch, update.Field123, wantUpdate.Field123, s.Rand.Logic.Index(), predict.Index(), unit, update, def)
	*unit, *update, *def, *missile = beforeUnit, beforeUpdate, beforeDef, beforeMissile
	s.AI.StackChanged = beforeChanged
	s.Rand.Logic.Reset(seed)
	got := uint8(ccall.CallIntPtr(monsterDeadStartPublicC544D80(), unit.CObj()))
	if !ordinaryMatch {
		t.Error("ordinary production path differs from frozen original startup context")
	}
	if got != ref.Result.ReturnByte || !matches() {
		t.Fatalf("actual C DeadStart differs:char=%d want%d duration=%d want%d rng-index=%d want%d", got, ref.Result.ReturnByte, update.Field123, wantUpdate.Field123, s.Rand.Logic.Index(), predict.Index())
	}
	t.Logf("DEAD START COMPOSITION NATIVE C PASS: mode=%s original-char=%d duration=%d real-balance=true real-RNG=true real-IsZombie=true native-soul-allocation-boundary=%t full-records=true", mode, got, update.Field123, kind == "soul")
}
