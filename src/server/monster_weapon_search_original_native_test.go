package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
)

//go:embed testdata/monster_hunt_weapon_original_v1.json.gz
var monsterWeaponOriginal544B20 []byte

type monsterWeaponStockCase struct {
	Input struct {
		Name, Path   string
		CandidatePos [2]uint32 `json:"candidate_pos"`
		ActorPos     [2]uint32 `json:"actor_pos"`
		ClassWord    uint32    `json:"class_word"`
		StartingDist uint32    `json:"starting_distance"`
		Eligible     uint32
		Visible      uint32
	}
	Observed struct {
		Replaced     bool
		DistanceWord uint32 `json:"distance_word"`
		Services     []string
	}
}

func TestMonsterWeaponSearchStockStartup544B20(t *testing.T) {
	reader, err := gzip.NewReader(bytes.NewReader(monsterWeaponOriginal544B20))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(raw)), "1e1747974871dfb2b16ce49b8bdeefaf6d6ba58817caaf62e512601c71e72032"; got != want {
		t.Fatalf("stock fixture digest %s, want %s", got, want)
	}
	var fixture struct{ Rows []monsterWeaponStockCase }
	if err := json.Unmarshal(raw, &fixture); err != nil || len(fixture.Rows) != 240 {
		t.Fatalf("read stock fixture: %v, %d rows", err, len(fixture.Rows))
	}
	rows := 0
	for _, row := range fixture.Rows {
		in, expected := row.Input, row.Observed
		if in.Path != "candidate" {
			continue
		}
		rows++
		t.Run(in.Name, func(t *testing.T) {
			candidate := &Object{ObjClass: object.Class(in.ClassWord), PosVec: types.Ptf(
				math.Float32frombits(in.CandidatePos[0]), math.Float32frombits(in.CandidatePos[1]))}
			actor := &Object{PosVec: types.Ptf(math.Float32frombits(in.ActorPos[0]), math.Float32frombits(in.ActorPos[1]))}
			initial := &Object{}
			state := MonsterWeaponSearchState544AE0{BestObject: initial, BestDistance: math.Float32frombits(in.StartingDist)}
			var calls []string
			state.consider544B20(candidate, actor, func(item *Object, class player.Class) bool {
				if item != candidate || class != player.Warrior {
					t.Fatal("eligible callback received the wrong candidate or class")
				}
				calls = append(calls, "eligible")
				return in.Eligible != 0
			}, func(unit, item *Object, mask int) bool {
				if unit != actor || item != candidate || mask != 0 {
					t.Fatal("interaction callback received the wrong units or flags")
				}
				calls = append(calls, "interact")
				return in.Visible != 0
			})
			want := initial
			if expected.Replaced {
				want = candidate
			}
			distance := math.Float32bits(state.BestDistance)
			if distance != expected.DistanceWord || state.BestObject != want || !slices.Equal(calls, expected.Services) {
				t.Fatalf("distance %#08x winner %p calls %v, want %#08x %p %v",
					distance, state.BestObject, calls, expected.DistanceWord, want, expected.Services)
			}
		})
	}
	if rows != 192 {
		t.Fatalf("executed %d stock candidates, want 192", rows)
	}
}
