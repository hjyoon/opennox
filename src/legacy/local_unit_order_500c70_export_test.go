package legacy

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/common/ntype"
)

type localUnitOrderLegacyServer500C70 struct {
	Server
	owner     ntype.PlayerInd
	orderType uint32
	result    int
	calls     int
}

func (s *localUnitOrderLegacyServer500C70) Nox_xxx_orderUnitLocal_500C70(owner ntype.PlayerInd, orderType uint32) int {
	s.owner = owner
	s.orderType = orderType
	s.calls++
	return s.result
}

func TestLocalUnitOrderExport500C70PreservesFixedWidthScalars(t *testing.T) {
	fake := &localUnitOrderLegacyServer500C70{}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	tests := []struct {
		name       string
		owner      int32
		orderType  int32
		result     int
		wantOrder  uint32
		wantResult int32
	}{
		{
			name:       "high bits",
			owner:      math.MinInt32,
			orderType:  -1985229329, // 0x89abcdef
			result:     math.MinInt32 + 0x4321,
			wantOrder:  0x89abcdef,
			wantResult: math.MinInt32 + 0x4321,
		},
		{
			name:       "positive limits",
			owner:      math.MaxInt32,
			orderType:  math.MaxInt32,
			result:     math.MaxInt32,
			wantOrder:  math.MaxInt32,
			wantResult: math.MaxInt32,
		},
		{
			name:       "all bits",
			owner:      -1,
			orderType:  -1,
			result:     -1,
			wantOrder:  math.MaxUint32,
			wantResult: -1,
		},
	}
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake.result = tc.result
			if got := localUnitOrderExportCall500C70(tc.owner, tc.orderType); got != tc.wantResult {
				t.Fatalf("result = %d, want %d", got, tc.wantResult)
			}
			if fake.owner != ntype.PlayerInd(tc.owner) || fake.orderType != tc.wantOrder {
				t.Fatalf("call = (%d, %#x), want (%d, %#x)", fake.owner, fake.orderType, tc.owner, tc.wantOrder)
			}
			if fake.calls != i+1 {
				t.Fatalf("calls = %d, want %d", fake.calls, i+1)
			}
		})
	}
}
