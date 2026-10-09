package opennox

import (
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestE2ENPCHoverSoundSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckNPCHoverSound("Tyler", "hover")
	if len(sc.steps) != 16 {
		t.Fatalf("steps=%d want=16", len(sc.steps))
	}
	for _, index := range []int{0, 3, 4, 9, 10} {
		step := sc.steps[index]
		want := time.Duration(1200)
		if index == 4 || index == 10 {
			want = 600
		}
		if step.ready == nil || step.fnc == nil || step.waitTimeout != want {
			t.Fatalf("step %d (%s) is not a bounded live check: %+v", index, step.name, step)
		}
	}
	for _, index := range []int{6, 12} {
		if dt := sc.steps[index].time - sc.steps[index-2].time; dt != 180 {
			t.Fatalf("held mouse duration=%d want=180", dt)
		}
		if dt := sc.steps[index+2].time - sc.steps[index].time; dt != 90 {
			t.Fatalf("natural mouse leave duration=%d want=90", dt)
		}
	}
	if sc.steps[15].fnc == nil || !strings.HasSuffix(sc.steps[15].name, " restore setup position") {
		t.Fatal("missing fixture cleanup")
	}
}

func TestE2EClothingDeathSchedule(t *testing.T) {
	for _, player := range []bool{false, true} {
		var sc e2eScenario
		sc.CheckClothingDeath(player, "F2Gearhart", "death")
		if len(sc.steps) != 5 {
			t.Fatalf("player=%t steps=%d want=5", player, len(sc.steps))
		}
		if step := sc.steps[0]; step.ready == nil || step.fnc == nil || step.waitTimeout != 1200 {
			t.Fatalf("player=%t setup is not a bounded live check", player)
		}
		if dt := sc.steps[2].time - sc.steps[0].time; dt != 60 {
			t.Fatalf("player=%t ordinary equip settling=%d want=60", player, dt)
		}
		if dt := sc.steps[4].time - sc.steps[2].time; dt != 60 {
			t.Fatalf("player=%t corpse observation=%d want=60", player, dt)
		}
		if sc.steps[2].fnc == nil || sc.steps[4].fnc == nil {
			t.Fatal("missing real death dispatch or final inventory observation")
		}
	}
}

func TestE2ENPCHoverClothingScenario(t *testing.T) {
	const path = "../scripts/e2e/solo-warrior-npc-hover-clothing.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	var sequence []string
	for _, step := range file.Steps {
		switch step.Action {
		case "audit-npc-clothing":
			sequence = append(sequence, "stock audit")
		case "check-gameplay-audio":
			if step.Text != "stock-npc-hover" || step.Mode < 0 || step.Mode > 1 {
				t.Fatalf("invalid FX checkpoint: %+v", step)
			}
			sequence = append(sequence, []string{"FX begin", "FX end"}[step.Mode])
		case "check-npc-hover-sound":
			if step.Item != "Tyler" {
				t.Fatalf("hover NPC=%s want=Tyler", step.Item)
			}
			sequence = append(sequence, "real hover")
		case "check-clothing-death":
			if step.Mode == 0 && step.Item == "F2Gearhart" {
				sequence = append(sequence, "stock NPC death")
			} else if step.Mode == 1 && step.Item == "" {
				sequence = append(sequence, "player death")
			} else {
				t.Fatalf("invalid death fixture: %+v", step)
			}
		}
	}
	if got := strings.Join(sequence, ","); got != "stock audit,FX begin,real hover,FX end,stock NPC death,player death" {
		t.Fatalf("scenario order=%s", got)
	}
	var sc e2eScenario
	sc.Load(path)
	counts := map[string]int{}
	for _, step := range sc.steps {
		for _, suffix := range []string{" one greeting", " natural wait completion", " verify worn clothing retained and ordinary loot dropped"} {
			if strings.HasSuffix(step.name, suffix) {
				counts[suffix]++
			}
		}
	}
	for _, suffix := range []string{" one greeting", " natural wait completion", " verify worn clothing retained and ordinary loot dropped"} {
		if counts[suffix] != 2 {
			t.Fatalf("%q checks=%d want=2", suffix, counts[suffix])
		}
	}
}
