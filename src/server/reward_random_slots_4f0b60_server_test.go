package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/prand"
)

func TestRewardRandomSlotsServerUsesExactLogicRNG4F0B60(t *testing.T) {
	tests := []struct {
		stage uint32
		seed  int
	}{
		{1, 0},
		{3, 2011},
		{4, 4095},
		{7, 1777},
		{9, 42},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("stage%d_seed%d", test.stage, test.seed), func(t *testing.T) {
			wantRNG := prand.New(test.seed)
			want := rewardRandomSlots4F0B60(test.stage, func(minimum, maximum int32) int32 {
				return int32(wantRNG.IntClamp(int(minimum), int(maximum)))
			})
			srv := new(Server)
			srv.Rand.Logic = prand.New(test.seed)
			got := srv.RewardRandomSlots4F0B60(test.stage)
			if got != want || srv.Rand.Logic.Index() != wantRNG.Index() {
				t.Fatalf("result/index = %#x/%d, want %#x/%d", got, srv.Rand.Logic.Index(), want, wantRNG.Index())
			}
		})
	}
}

func TestRewardRandomSlotsServerFixedStagesDoNotDereferenceRNG4F0B60(t *testing.T) {
	srv := new(Server)
	if got := srv.RewardRandomSlots4F0B60(0); got != 1 {
		t.Fatalf("stage zero = %#x, want 1", got)
	}
	if got := srv.RewardRandomSlots4F0B60(10); got != 16 {
		t.Fatalf("stage ten = %#x, want 16", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("weighted stage did not preserve nil logic-RNG fault")
		}
	}()
	srv.RewardRandomSlots4F0B60(1)
}
