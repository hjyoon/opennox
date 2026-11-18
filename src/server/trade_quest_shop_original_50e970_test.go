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
)

//go:embed testdata/quest_shop_stock_original_v1.json.gz
var questShopStockOriginal50E970 []byte

type questShopStockEvent50E970 struct {
	Service  string
	Name     string
	Input    uint32
	Result   uint32
	Value    uint32
	Session  uint32
	Item     uint32
	Marker   uint32
	Category uint32
	Stage    uint32
	Min      int32
	Max      int32
}

func TestQuestShopStockOriginal50E970(t *testing.T) {
	z, err := gzip.NewReader(bytes.NewReader(questShopStockOriginal50E970))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "86663280a3b338c1e44389c4fd3cbf92aefaf0d59e6b5a3a95d1d7187192bd5b" {
		t.Fatalf("original stock fixture digest %s", got)
	}
	var original struct {
		Rows []struct {
			Input struct {
				Name           string
				Profile        string
				Cutoff         uint32
				CompareStage   uint32 `json:"compare_stage"`
				RewardStage    uint32 `json:"reward_stage"`
				Random         int32
				MissingMarker  bool  `json:"missing_marker"`
				MissingFixed   []int `json:"missing_fixed"`
				MissingRewards []int `json:"missing_rewards"`
			}
			Observed struct{ Events []questShopStockEvent50E970 }
		}
	}
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if len(original.Rows) != 18 {
		t.Fatalf("original stock cases %d", len(original.Rows))
	}
	for _, row := range original.Rows {
		in, want := row.Input, row.Observed.Events
		t.Run(in.Name, func(t *testing.T) {
			if len(want) < 4 || want[0].Service != "flags" || want[1].Service != "balance" || want[2].Service != "convert" {
				t.Fatal("original startup service prefix")
			}
			cutoff := netFXTruncateSignedDWORD419A70(math.Float32frombits(in.Cutoff))
			if in.Profile == "nearest" {
				cutoff = questShopRoundFloat32ToInt32_50E970(math.Float32frombits(in.Cutoff))
			}
			if want[2].Input != in.Cutoff || want[2].Result != uint32(cutoff) {
				t.Fatalf("cutoff conversion %#x, want %#x", cutoff, want[2].Result)
			}
			const sessionID, markerID = 0x33000100, 0x33004000
			marker := &Object{}
			data := &RewardMarkerInitData{}
			roles := map[*Object]uint32{marker: markerID}
			var createCalls, rewardCalls, inserted, stageReads int
			index := 3
			emit := func(got questShopStockEvent50E970) {
				t.Helper()
				if index >= len(want) {
					t.Fatalf("extra stock event %+v", got)
				}
				if got != want[index] {
					t.Fatalf("stock event %d %+v, want %+v", index, got, want[index])
				}
				index++
			}
			currentStage := func() int32 {
				stage := in.CompareStage
				if stageReads != 0 {
					stage = in.RewardStage
				}
				stageReads++
				emit(questShopStockEvent50E970{Service: "quest_stage", Value: stage})
				return int32(stage)
			}
			loaded, _ := loadQuestShopItems50E970(currentStage, cutoff, questShopItemLoadHooks50E970{
				newObject: func(name string) *Object {
					i := createCalls
					createCalls++
					var item *Object
					if name == "RewardMarker" {
						if !in.MissingMarker {
							item = marker
						}
					} else if !slices.Contains(in.MissingFixed, i) {
						item = &Object{}
						roles[item] = 0x33005000 + uint32(i)*32
					}
					emit(questShopStockEvent50E970{Service: "create", Name: name, Result: roles[item]})
					return item
				},
				addItem: func(item *Object) bool {
					emit(questShopStockEvent50E970{Service: "insert", Session: sessionID, Item: roles[item]})
					inserted++
					return true
				},
				markerData: func(obj *Object) *RewardMarkerInitData {
					if obj != marker {
						t.Fatal("reward marker identity changed")
					}
					return data
				},
				activateReward: func(obj *Object, stage uint32) *Object {
					i := rewardCalls
					rewardCalls++
					var item *Object
					if !slices.Contains(in.MissingRewards, i) {
						item = &Object{}
						roles[item] = 0x33006000 + uint32(i)*32
					}
					emit(questShopStockEvent50E970{Service: "reward", Marker: roles[obj], Category: data.CategoryMask,
						Stage: stage, Result: roles[item]})
					return item
				},
				randomInt: func(lo, hi int32) int32 {
					emit(questShopStockEvent50E970{Service: "rng", Min: lo, Max: hi, Value: uint32(in.Random)})
					return in.Random
				},
				delayedDelete: func(obj *Object) {
					emit(questShopStockEvent50E970{Service: "delete", Marker: roles[obj]})
				},
			})
			if index != len(want) || inserted != loaded {
				t.Fatalf("stock events %d/%d, reported items %d/%d", index, len(want), loaded, inserted)
			}
		})
	}
}
