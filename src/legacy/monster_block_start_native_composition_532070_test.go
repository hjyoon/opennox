package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterBlockStartPublicCNativeComposition532070(t *testing.T) {
	const child = "NOX_MONSTER_BLOCK_START_NATIVE_COMPOSE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"wait", "refresh", "npc-expired", "monster-expired", "push-failed", "npc-wrap", "monster-wrap"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestMonsterBlockStartPublicCNativeComposition532070$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual block-start %s C composition: %v", mode, err)
				}
			})
		}
		return
	}
	path := "../server/testdata/monster_block_start_native_original_532070.json"
	if override := os.Getenv("NOX_MONSTER_BLOCK_START_ORIGINAL_FIXTURE"); override != "" {
		path = override
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		ControlWord uint16 `json:"control_word"`
		Rows        []struct {
			Input struct {
				Shield    int    `json:"shield"`
				Frame     uint32 `json:"frame"`
				FPS       uint32 `json:"fps"`
				Deadline  uint32 `json:"deadline"`
				Subclass  uint32 `json:"subclass"`
				HeadIndex int8   `json:"head_index"`
				PushNil   bool   `json:"push_nil"`
				PushLow   uint8  `json:"push_low"`
				Mutation  struct {
					Stage string `json:"stage"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				ReturnByte uint8 `json:"return_byte"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.ControlWord != 0xe3f || len(fixture.Rows) != 485 {
		t.Fatal("original block fixture identity/profile/count")
	}
	key := func(shield int, frame, fps, deadline, subclass uint32, pushNil bool, low uint8) string {
		return fmt.Sprintf("%d/%d/%d/%d/%d/%t/%d", shield, frame, fps, deadline, subclass, pushNil, low)
	}
	wantBytes := make(map[string]uint8)
	for _, row := range fixture.Rows {
		c := row.Input
		if c.HeadIndex != 1 || c.Mutation.Stage != "" {
			continue
		}
		k := key(c.Shield, c.Frame, c.FPS, c.Deadline, c.Subclass, c.PushNil, c.PushLow)
		if old, exists := wantBytes[k]; exists && old != row.Result.ReturnByte {
			t.Fatal("inconsistent original char alias")
		}
		wantBytes[k] = row.Result.ReturnByte
	}
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	s.SetFrame(100)
	s.SetTickRate(30)
	update.AIStack = [24]server.AIStackItem{}
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
	update.AIStack[1].Action = uint32(ai.ACTION_BLOCK_ATTACK)
	deadline, shield := uint32(99), 0
	unit.ObjSubClass = 0
	pushNil := false
	switch mode {
	case "wait":
		deadline = 101
	case "refresh":
		shield = 1
	case "npc-expired":
		unit.ObjSubClass = object.SubClass(object.MonsterNPC)
	case "monster-expired":
	case "push-failed":
		update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
		pushNil = true
	case "npc-wrap":
		s.SetFrame(0xffffffff)
		shield = 1
		unit.ObjSubClass = object.SubClass(object.MonsterNPC)
	case "monster-wrap":
		s.SetFrame(0xffffffff)
		shield = 1
	default:
		t.Fatal("unknown block composition mode")
	}
	update.AIStack[1].Args[0] = uintptr(deadline)
	if shield != 0 {
		s.Map.AddObjectToIndex(missile)
	}
	if got := Nox_xxx_monsterTestBlockShield_533E70(unit); got != shield {
		t.Fatalf("actual missile shield query=%d want%d", got, shield)
	}
	a, ok := server.GetAIAction(ai.ACTION_BLOCK_ATTACK).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary block dispatcher")
	}
	beforeUnit, beforeUpdate, beforeMissile := *unit, *update, *missile
	frame, fps, subclass := s.Frame(), s.TickRate(), uint32(unit.ObjSubClass)
	s.AI.StackChanged = false
	a.Update(unit)
	expectedUnit, expectedUpdate, expectedMissile, changed := *unit, *update, *missile, s.AI.StackChanged
	low := uint8(128)
	if mode == "monster-expired" || mode == "monster-wrap" {
		head := update.AIStackHead()
		if head == nil || head.Type() != ai.ACTION_BLOCK_FINISH {
			t.Fatal("ordinary real Pop/Push did not produce block-finish")
		}
		low = uint8(uintptr(unsafe.Pointer(head)))
	}
	if mode == "push-failed" && (update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_DEAD) {
		t.Fatal("ordinary DEAD-base push rejection did not persist")
	}
	lookup := key(shield, frame, fps, deadline, subclass, pushNil, low)
	want, exists := wantBytes[lookup]
	if !exists {
		t.Fatalf("missing exact frozen original char context %s", lookup)
	}
	*unit, *update, *missile = beforeUnit, beforeUpdate, beforeMissile
	s.AI.StackChanged = false
	t.Logf("BLOCK START ORDINARY COMPOSITION COMPLETED: mode=%s unit=%p update=%p missile=%p frame=%d fps=%d deadline=%d shield=%d subclass=%d push-nil=%t push-low=%d original-char=%d", mode, unit, update, missile, frame, fps, deadline, shield, subclass, pushNil, low, want)
	got := uint8(ccall.CallIntPtr(monsterBlockStartPublicC532070(), unit.CObj()))
	rawUnit := func(p *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	rawUpdate := func(p *server.MonsterUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))
	}
	if got != want || s.AI.StackChanged != changed || !bytes.Equal(rawUnit(unit), rawUnit(&expectedUnit)) || !bytes.Equal(rawUpdate(update), rawUpdate(&expectedUpdate)) || !bytes.Equal(rawUnit(missile), rawUnit(&expectedMissile)) {
		t.Fatalf("actual C composition mismatch: mode=%s char=%d/%d changed=%t/%t", mode, got, want, s.AI.StackChanged, changed)
	}
	t.Logf("BLOCK START NATIVE C PASS: mode=%s shield=%d original-char=%d native-object/update/missile=true real-Pop/Push=true full-canaries=true", mode, shield, got)
}
