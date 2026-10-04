package opennox

import (
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EAIFearSchedule(t *testing.T) {
	for _, kind := range []string{"Spider", "Troll", "Urchin", "NPC"} {
		for _, slope := range []string{"ascending", "descending"} {
			mode := kind + "/" + slope
			t.Run(mode, func(t *testing.T) {
				gotKind, descending, ok := e2eAIFearMode(mode)
				if !ok || gotKind != kind || descending != (slope == "descending") {
					t.Fatal("stock fear mode was not preserved")
				}
				var sc e2eScenario
				sc.CheckAIFear(mode, "fear")
				if len(sc.steps) != 6 {
					t.Fatalf("steps=%d want=6", len(sc.steps))
				}
				for i, timeout := range []int{1200, 600, 600, 3600, 300} {
					if sc.steps[i].ready == nil || int(sc.steps[i].waitTimeout) != timeout || int(sc.steps[i].time) != i {
						t.Fatalf("step %d lost bounded readiness/time", i)
					}
				}
				if !strings.HasSuffix(sc.steps[1].name, " autonomous first attack then ordinary fear") ||
					!strings.HasSuffix(sc.steps[2].name, " stock fear response and client state") ||
					!strings.HasSuffix(sc.steps[3].name, " stock result and continued combat") ||
					!strings.HasSuffix(sc.steps[4].name, " ordinary deletion") || sc.steps[5].time-sc.steps[4].time != 12 {
					t.Fatal("ordinary attack/fear/natural expiry/cleanup order changed")
				}
			})
		}
	}
}

func TestE2EAIFearRejectsInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "Spider", "Wizard/ascending", "spider/ascending", "Spider/other", "NPC/ascending/forced", "Urchin/", "/ascending"} {
		t.Run(strings.ReplaceAll(mode, "/", ":"), func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid mode scheduled a probe")
				}
			}()
			sc.CheckAIFear(mode, "invalid")
		})
	}
}

func TestE2EAIFearDuration(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float64
		want  uint16
		ok    bool
	}{
		{"zero", 0, 0, false}, {"short", 24.999, 0, false}, {"minimum", 25, 25, true},
		{"truncation", 25.9, 25, true}, {"maximum", 2900.9, 2900, true}, {"long", 2901, 0, false},
		{"negative", -25, 0, false}, {"NaN", math.NaN(), 0, false},
		{"positive infinity", math.Inf(1), 0, false}, {"negative infinity", math.Inf(-1), 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, ok := e2eAIFearDuration(tc.value); got != tc.want || ok != tc.ok {
				t.Fatalf("duration=%d/%t want=%d/%t", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestE2EAIFearImmuneReadOnly(t *testing.T) {
	for _, mode := range []string{"stock NPC", "other immune monster", "nonimmune", "Coop NPC", "nil", "player", "other subclass"} {
		t.Run(mode, func(t *testing.T) {
			unit := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterNPC | object.MonsterImmuneFear)}
			coop := mode == "Coop NPC"
			switch mode {
			case "other immune monster":
				unit.ObjSubClass = object.SubClass(object.MonsterImmuneFear)
			case "nonimmune":
				unit.ObjSubClass = object.SubClass(object.MonsterNPC)
			case "nil":
				unit = nil
			case "player":
				unit.ObjClass = object.ClassPlayer
			case "other subclass":
				unit.ObjSubClass = object.SubClass(object.MonsterImmunePoison)
			}
			var before server.Object
			if unit != nil {
				before = *unit
			}
			if got := e2eAIFearImmune(unit, coop); got != (mode == "stock NPC" || mode == "other immune monster") {
				t.Fatalf("immune=%t mode=%s", got, mode)
			}
			if unit != nil && *unit != before {
				t.Fatal("immunity observer changed stock target state")
			}
		})
	}
}

func TestE2EAIFearScheduledReadOnly(t *testing.T) {
	for _, mode := range []string{"valid", "below top", "nil unit", "nil update", "nonmonster", "no fear", "other buff", "missing dependency", "wrong dependency", "not adjacent", "partial push", "outside live stack", "negative index", "invalid index", "no flee"} {
		t.Run(mode, func(t *testing.T) {
			data := &server.MonsterUpdateData{AIStackInd: 2}
			data.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_FIGHT)}
			data.AIStack[1] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_IS_ENCHANTED), Args: [4]uintptr{uintptr(server.ENCHANT_AFRAID)}}
			data.AIStack[2] = server.AIStackItem{Action: uint32(ai.ACTION_FLEE), Args: [4]uintptr{uintptr(math.Float32bits(123)), uintptr(math.Float32bits(-456))}}
			unit := &server.Object{ObjClass: object.ClassMonster, Buffs: 1 << server.ENCHANT_AFRAID, UpdateData: unsafe.Pointer(data)}
			switch mode {
			case "below top":
				data.AIStackInd = 3
				data.AIStack[3].Action = uint32(ai.ACTION_WAIT)
			case "nil unit":
				unit = nil
			case "nil update":
				unit.UpdateData = nil
			case "nonmonster":
				unit.ObjClass = object.ClassPlayer
			case "no fear":
				unit.Buffs = 0
			case "other buff":
				unit.Buffs = 1 << server.ENCHANT_CONFUSED
			case "missing dependency":
				data.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			case "wrong dependency":
				data.AIStack[1].Args[0] = uintptr(server.ENCHANT_CONFUSED)
			case "not adjacent":
				data.AIStackInd = 3
				data.AIStack[3] = data.AIStack[2]
				data.AIStack[2].Action = uint32(ai.ACTION_WAIT)
			case "partial push":
				data.AIStackInd = 1
			case "outside live stack":
				data.AIStackInd = 0
			case "negative index":
				data.AIStackInd = -1
			case "invalid index":
				data.AIStackInd = int8(len(data.AIStack))
			case "no flee":
				data.AIStack[2].Action = uint32(ai.ACTION_WAIT)
			}
			before := *data
			if got := e2eAIFearScheduled(unit); got != (mode == "valid" || mode == "below top") {
				t.Fatalf("fear/FLEE accepted=%t mode=%s", got, mode)
			}
			if *data != before {
				t.Fatal("observer supplied native action/coordinates")
			}
		})
	}
}

func TestE2EAIFearAway(t *testing.T) {
	for _, tc := range []struct {
		name     string
		to, away types.Pointf
		want     bool
	}{
		{"minimum", types.Ptf(16, 0), types.Ptf(10, 0), true},
		{"ascending", types.Ptf(16, 16), types.Ptf(10, 10), true},
		{"descending", types.Ptf(16, -16), types.Ptf(10, -10), true},
		{"short", types.Ptf(15.99, 0), types.Ptf(10, 0), false},
		{"towards", types.Ptf(-16, 0), types.Ptf(10, 0), false},
		{"cross", types.Ptf(0, 16), types.Ptf(10, 0), false},
		{"curved escape", types.Ptf(0, 30), types.Ptf(10, 0), true},
		{"opposite exit", types.Ptf(-40, 0), types.Ptf(10, 0), true},
		{"stationary", types.Pointf{}, types.Ptf(10, 0), false},
		{"zero away", types.Ptf(16, 0), types.Pointf{}, false},
		{"NaN position", types.Ptf(float32(math.NaN()), 16), types.Ptf(10, 0), false},
		{"NaN away", types.Ptf(16, 0), types.Ptf(float32(math.NaN()), 0), false},
		{"infinite position", types.Ptf(float32(math.Inf(1)), 0), types.Ptf(10, 0), false},
		{"infinite away", types.Ptf(16, 0), types.Ptf(float32(math.Inf(1)), 0), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			from := types.Ptf(100, -100)
			if got := e2eAIFearAway(from, from.Add(tc.to), tc.away); got != tc.want {
				t.Fatalf("away=%t want=%t", got, tc.want)
			}
		})
	}
}

func TestE2EAIFearFreshHit(t *testing.T) {
	for _, mode := range []string{"valid", "nil host", "nil health", "unchanged HP", "healed", "old frame", "expiry frame", "nil source"} {
		t.Run(mode, func(t *testing.T) {
			health := &server.HealthData{Cur: 990, Max: 2000}
			host := &server.Object{HealthData: health, Frame134: 701, Obj130: new(server.Object)}
			switch mode {
			case "nil host":
				host = nil
			case "nil health":
				host.HealthData = nil
			case "unchanged HP":
				health.Cur = 1000
			case "healed":
				health.Cur = 1001
			case "old frame":
				host.Frame134 = 699
			case "expiry frame":
				host.Frame134 = 700
			case "nil source":
				host.Obj130 = nil
			}
			before := *health
			if got := e2eAIFearFreshHit(host, 1000, 700); got != (mode == "valid") {
				t.Fatalf("fresh hit=%t mode=%s", got, mode)
			}
			if *health != before {
				t.Fatal("observer supplied HP results")
			}
		})
	}
}

func TestE2EAIFearApproach(t *testing.T) {
	for _, tc := range []struct {
		name  string
		delta types.Pointf
		walk  bool
	}{
		{"ascending", types.Ptf(400, 300), true},
		{"descending", types.Ptf(400, -300), true},
		{"left", types.Ptf(-300, 0), true},
		{"just outside", types.Ptf(48.01, 0), true},
		{"boundary", types.Ptf(48, 0), false},
		{"nearby", types.Ptf(16, 0), false},
		{"same point", types.Pointf{}, false},
		{"NaN", types.Ptf(float32(math.NaN()), 0), false},
		{"infinity", types.Ptf(float32(math.Inf(1)), 0), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			from := types.Ptf(1000, -1000)
			aim, walk := e2eAIFearApproach(from, from.Add(tc.delta))
			if walk != tc.walk {
				t.Fatalf("walk=%t want=%t", walk, tc.walk)
			}
			if walk {
				step := aim.Sub(from)
				if math.Abs(step.Len()-128) > 0.001 || step.X*tc.delta.X+step.Y*tc.delta.Y <= 0 {
					t.Fatal("aim is not a short world-space direction toward the target")
				}
			} else if aim != (types.Pointf{}) {
				t.Fatal("nearby/invalid geometry produced input")
			}
		})
	}
}

func TestE2EAIFearScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-ai-fear.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, step := range file.Steps {
		if step.Action != "check-ai-fear" {
			continue
		}
		if _, _, ok := e2eAIFearMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/repeated stock fear mode %q", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 8 {
		t.Fatalf("stock fear modes=%d want=8", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " stock result and continued combat") {
			checks++
		}
	}
	if checks != 8 {
		t.Fatalf("bounded natural observers=%d want=8", checks)
	}
}
