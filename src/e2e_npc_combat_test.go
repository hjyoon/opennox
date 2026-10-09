package opennox

import (
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
	"gopkg.in/yaml.v2"
)

func TestE2ENPCAttackAnimationSchedule(t *testing.T) {
	for item, index := range map[string]int{"LongSword": 28, "StaffWooden": 29, "WarHammer": 39} {
		t.Run(item, func(t *testing.T) {
			if got, ok := e2eNPCAttackAnimationIndex(item); !ok || got != index {
				t.Fatalf("animation=%d/%t want=%d", got, ok, index)
			}
			var sc e2eScenario
			sc.CheckNPCAttackAnimation(item, "animation")
			if len(sc.steps) != 3 {
				t.Fatalf("steps=%d want=3", len(sc.steps))
			}
			for _, step := range sc.steps[:2] {
				if step.ready == nil || step.fnc == nil || step.waitTimeout != 1200 {
					t.Fatalf("unbounded live animation check: %+v", step)
				}
			}
			if sc.steps[2].time-sc.steps[1].time != 12 {
				t.Fatal("missing ordinary deletion wait")
			}
		})
	}
}

func TestE2ENPCIncomingAttackSchedule(t *testing.T) {
	for kind, damage := range map[string]object.DamageType{
		"GruntAxe": object.DamageBlade, "StoneGolem": object.DamageCrush,
		"Scorpion": object.DamageImpale, "Ghost": object.DamageDrain,
		"Spider": object.DamageBite, "VileZombie": object.DamageClaw,
	} {
		t.Run(kind, func(t *testing.T) {
			if got, ok := e2eNPCIncomingAttackType(kind); !ok || got != damage {
				t.Fatalf("damage=%d/%t want=%d", got, ok, damage)
			}
			var sc e2eScenario
			sc.CheckNPCIncomingAttack(kind, "incoming")
			if len(sc.steps) != 3 {
				t.Fatalf("steps=%d want=3", len(sc.steps))
			}
			for _, step := range sc.steps[:2] {
				if step.ready == nil || step.fnc == nil || step.waitTimeout != 1200 {
					t.Fatalf("unbounded live NPC incoming check: %+v", step)
				}
			}
			if sc.steps[2].time-sc.steps[1].time != 12 {
				t.Fatal("missing ordinary deletion wait")
			}
		})
	}
}

func TestE2ENPCCombatRejectsInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "NPC", "LongSword/injected", "spider", "MechanicalGolem/forced"} {
		for _, incoming := range []bool{false, true} {
			t.Run(mode+"/"+map[bool]string{false: "animation", true: "incoming"}[incoming], func(t *testing.T) {
				var sc e2eScenario
				defer func() {
					if recover() == nil || len(sc.steps) != 0 {
						t.Fatal("invalid NPC mode was scheduled")
					}
				}()
				if incoming {
					sc.CheckNPCIncomingAttack(mode, "invalid")
				} else {
					sc.CheckNPCAttackAnimation(mode, "invalid")
				}
			})
		}
	}
}

func TestE2ENPCCombatScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-npc-combat.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	animations, incoming := make(map[string]bool), make(map[object.DamageType]bool)
	for _, step := range file.Steps {
		switch step.Action {
		case "check-npc-attack-animation":
			if _, ok := e2eNPCAttackAnimationIndex(step.Item); !ok || animations[step.Item] {
				t.Fatalf("invalid/repeated NPC weapon: %s", step.Item)
			}
			animations[step.Item] = true
		case "check-npc-incoming-attack":
			damage, ok := e2eNPCIncomingAttackType(step.Item)
			if !ok || incoming[damage] {
				t.Fatalf("invalid/repeated NPC damage: %s", step.Item)
			}
			incoming[damage] = true
		}
	}
	if len(animations) != 3 || len(incoming) != 6 {
		t.Fatalf("weapon animations=%d incoming types=%d want=3/6", len(animations), len(incoming))
	}
	var sc e2eScenario
	sc.Load(path)
	counts := map[string]int{}
	for _, step := range sc.steps {
		for _, suffix := range []string{" three full server and client attack cycles", " three natural enemy hits on NPC"} {
			if strings.HasSuffix(step.name, suffix) {
				counts[suffix]++
			}
		}
	}
	if counts[" three full server and client attack cycles"] != 3 || counts[" three natural enemy hits on NPC"] != 6 {
		t.Fatalf("live checks=%v", counts)
	}
	for _, action := range []string{"damage-monster", "damage-player", "cast", "set-monster-health", "prepare-urchin-ranged-attack"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("injected attack result: %s", action)
		}
	}
}
