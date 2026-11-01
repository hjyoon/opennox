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
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

// Tile/path remain declared terrain services. Clocks, RNG, Push and the
// empty native wall-map trace use production code on both public entries.
func TestMonsterInterestingProduction5466F0(t *testing.T) {
	const child = "OPENNOX_INTERESTING_PRODUCTION_5466F0"
	mode := os.Getenv(child)
	if mode == "" {
		for _, path := range []string{"C", "Go"} {
			for _, branch := range []string{"direct", "blocked", "no-path", "lava"} {
				for _, seed := range []int{0, 4095} {
					for slots := 0; slots <= 5; slots++ {
						mode := fmt.Sprintf("%s:%s:%d:%d", path, branch, seed, slots)
						t.Run(mode, func(t *testing.T) {
							cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterInterestingProduction5466F0$", "-test.v")
							cmd.Env = append(os.Environ(), child+"="+mode)
							out, err := cmd.CombinedOutput()
							t.Logf("%s", out)
							if err != nil {
								t.Fatal("native production composition", err)
							}
						})
					}
				}
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 4 {
		t.Fatal("production mode")
	}
	seed, e1 := strconv.Atoi(parts[2])
	slots, e2 := strconv.Atoi(parts[3])
	if e1 != nil || e2 != nil || (seed != 0 && seed != 4095) || slots < 0 || slots > 5 {
		t.Fatal("production seed/capacity")
	}
	s, unit, update, enemy := monsterMainBlockLegacyFixture547210(t)
	if monsterInterestingSoundRuntimeOverride5466F0 != nil || monsterReactiveRuntimeOverride545DA0 != nil {
		t.Fatal("clean production services required")
	}
	s.SetFrame(100)
	s.SetTickRate(30)
	s.Rand.Logic = prand.New(seed)
	s.AI.StackChanged = false
	unit.PosVec = types.Ptf(300, 300)
	update.Field97, update.Field101 = 7, 99
	update.Field99X, update.Field99Y = 312.5, 293.75
	if parts[1] == "blocked" {
		unit.PosVec = types.Ptf(-46, -46)
	}
	for i := range update.AIStack {
		update.AIStack[i] = server.AIStackItem{Action: uint32(ai.ACTION_GUARD)}
		for j := range update.AIStack[i].Args {
			update.AIStack[i].Args[j] = uintptr(0xa5a50000) + uintptr(i*16+j)
		}
	}
	update.AIStackInd = int8(len(update.AIStack) - 1 - slots)
	wantUpdate := *update
	wantUnit, wantEnemy, wantHealth := *unit, *enemy, *unit.HealthData
	control := prand.New(seed)
	actions := []ai.ActionType{}
	switch parts[1] {
	case "direct":
		actions = []ai.ActionType{ai.DEPENDENCY_NO_INTERESTING_SOUND, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.ACTION_WAIT, ai.ACTION_FACE_LOCATION}
	case "blocked":
		actions = []ai.ActionType{ai.DEPENDENCY_NO_INTERESTING_SOUND, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_NOT_FRUSTRATED, ai.DEPENDENCY_LOCATION_IS_SAFE, ai.ACTION_MOVE_TO}
	case "no-path", "lava":
	default:
		t.Fatal("production branch")
	}
	wantUpdate.Field97 = 0
	accepted := min(slots, len(actions))
	if accepted > 0 {
		wantUpdate.Field2, wantUpdate.Field67, wantUpdate.Field74, wantUpdate.Field91 = 0, 0, 0, 0
		wantUpdate.Field120_1, wantUpdate.Field120_2, wantUpdate.Field120_3 = 0, 0, 0
		wantUpdate.Field124, wantUpdate.Field137 = 100, 100
	}
	for _, action := range actions[:accepted] {
		wantUpdate.AIStackInd++
		item := &wantUpdate.AIStack[wantUpdate.AIStackInd]
		item.Action = uint32(action)
		item.Field5 = 0
		switch action {
		case ai.ACTION_WAIT:
			item.Args[0] = uintptr(uint32(100) + uint32(control.IntClamp(30, 60)))
		case ai.ACTION_FACE_LOCATION, ai.DEPENDENCY_LOCATION_IS_SAFE, ai.ACTION_MOVE_TO:
			item.Args[0], item.Args[1] = uintptr(math.Float32bits(float32(312.5))), uintptr(math.Float32bits(float32(293.75)))
			if action == ai.ACTION_MOVE_TO {
				item.Args[2] = 0
			}
		}
	}
	tileCalls, pathCalls := 0, 0
	tile := func(point *types.Pointf) int {
		if unsafe.Pointer(point) != unsafe.Pointer(&update.Field99X) {
			t.Fatal("live terrain coordinate pointer")
		}
		tileCalls++
		if parts[1] == "lava" {
			return 6
		}
		return 0
	}
	path := func(obj *server.Object, point *types.Pointf) bool {
		if obj != unit || unsafe.Pointer(point) != unsafe.Pointer(&update.Field99X) {
			t.Fatal("native path identities")
		}
		pathCalls++
		return parts[1] != "no-path"
	}
	var result int
	if parts[0] == "C" {
		world := monsterInterestingSoundOriginalRuntime5466F0{Reactive: monsterReactiveRuntime545DA0(), Tile: tile, Path: path,
			Trace: func(from, to types.Pointf) bool { return s.MapTraceRayAt(from, to, nil, nil, 9) }}
		monsterInterestingSoundRuntimeOverride5466F0 = &world
		defer func() { monsterInterestingSoundRuntimeOverride5466F0 = nil }()
		result = monsterInterestingOriginalC5466F0(unit)
	} else if parts[0] == "Go" {
		// Public ordinary wrapper uses a by-value tile API; path still receives
		// the exact live MUD pointer, and trace remains the real native service.
		result = s.MonsterInterestingSound5466F0(unit, func(point types.Pointf) int {
			if point != (types.Pointf{X: update.Field99X, Y: update.Field99Y}) {
				t.Fatal("ordinary tile coordinates")
			}
			tileCalls++
			if parts[1] == "lava" {
				return 6
			}
			return 0
		}, path)
	} else {
		t.Fatal("production entry")
	}
	wantPath := 1
	if parts[1] == "lava" {
		wantPath = 0
	}
	if result != 1 || tileCalls != 1 || pathCalls != wantPath {
		t.Fatal("terrain service gates/result", result, tileCalls, pathCalls)
	}
	if s.AI.StackChanged != (accepted > 0) || s.Rand.Logic.Index() != control.Index() {
		t.Fatal("production stack/RNG state", s.AI.StackChanged, s.Rand.Logic.Index(), control.Index())
	}
	for _, pair := range []struct {
		got, want unsafe.Pointer
		size      uintptr
	}{
		{unit.CObj(), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
		{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
		{enemy.CObj(), unsafe.Pointer(&wantEnemy), unsafe.Sizeof(wantEnemy)},
		{unsafe.Pointer(unit.HealthData), unsafe.Pointer(&wantHealth), unsafe.Sizeof(wantHealth)},
	} {
		got, want := unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))
		if !bytes.Equal(got, want) {
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("whole native production records/padding byte%d got%02x want%02x", i, got[i], want[i])
				}
			}
		}
	}
	t.Logf("INTERESTING PRODUCTION PASS: %s actual-clocks-RNG-Push-empty-wall-trace=true tile-path=declared", mode)
}
