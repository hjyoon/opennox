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
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

// Complement the sealed Push-service comparisons with the actual public C
// entry, native server clock, production prand and real stack-capacity logic.
func TestMonsterReactiveProduction545DA0(t *testing.T) {
	const child = "NOX_REACTIVE_PRODUCTION_545DA0"
	mode := os.Getenv(child)
	if mode == "" {
		for _, root := range []string{"Sound", "Look"} {
			for _, seed := range []int{0, 4095} {
				for slots := 0; slots <= 4; slots++ {
					mode := fmt.Sprintf("%s:%d:%d", root, seed, slots)
					t.Run(mode, func(t *testing.T) {
						bin, err := os.Executable()
						if err != nil {
							t.Fatal(err)
						}
						cmd := exec.Command(bin, "-test.run=^TestMonsterReactiveProduction545DA0$", "-test.v")
						cmd.Env = append(os.Environ(), child+"="+mode)
						out, err := cmd.CombinedOutput()
						t.Logf("%s", out)
						if err != nil {
							t.Fatal("actual public C production composition", err)
						}
					})
				}
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 3 {
		t.Fatal("production mode")
	}
	seed, e1 := strconv.Atoi(parts[1])
	slots, e2 := strconv.Atoi(parts[2])
	if e1 != nil || e2 != nil || (seed != 0 && seed != 4095) || slots < 0 || slots > 4 {
		t.Fatal("production seed/capacity")
	}
	s, unit, update, enemy := monsterMainBlockLegacyFixture547210(t)
	if monsterReactiveRuntimeOverride545DA0 != nil {
		t.Fatal("production services must remain intact")
	}
	s.SetFrame(100)
	s.SetTickRate(30)
	s.Rand.Logic = prand.New(seed)
	s.AI.StackChanged = false
	unit.Pos132.X = math.Float32frombits(0x7f812345)
	unit.Pos132.Y = math.Float32frombits(0x80000000)
	update.StatusFlags = object.MonStatusInjured
	update.Field97, update.Field101 = 7, 99
	update.Field99X, update.Field99Y = 12.5, -6.25
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
	actions := []ai.ActionType{ai.ACTION_FACE_LOCATION}
	if parts[0] == "Sound" {
		actions = []ai.ActionType{ai.DEPENDENCY_NO_INTERESTING_SOUND, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.ACTION_WAIT, ai.ACTION_FACE_LOCATION}
		wantUpdate.Field97 = 0
	} else if parts[0] != "Look" {
		t.Fatal("production root")
	}
	accepted := min(slots, len(actions))
	if accepted > 0 {
		// The real Push service also performs the established action reset;
		// these writes are outside the reactive roots' declared Push service.
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
		case ai.ACTION_FACE_LOCATION:
			if parts[0] == "Sound" {
				item.Args[0], item.Args[1] = uintptr(math.Float32bits(float32(12.5))), uintptr(math.Float32bits(float32(-6.25)))
			} else {
				item.Args[0], item.Args[1] = 0x7f812345, 0x80000000
			}
		}
	}
	if result := monsterReactiveOriginalC545DA0(parts[0], unit); result != 1 {
		t.Fatal("original successful DWORD even with full stack", result)
	}
	if s.AI.StackChanged != (accepted > 0) || s.Rand.Logic.Index() != control.Index() {
		t.Fatal("production stack-change/RNG index", s.AI.StackChanged, s.Rand.Logic.Index(), control.Index())
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
		if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
			got, want := unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("whole native production records/padding differ at byte%d got%02x want%02x", i, got[i], want[i])
				}
			}
		}
	}
	t.Logf("REACTIVE PRODUCTION C PASS: %s unit=%p update=%p actual-frame-FPS-prand-Push=true", mode, unit, update)
}
