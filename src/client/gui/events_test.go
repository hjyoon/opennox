package gui

import "testing"

func TestEventRespIntSignExtendsLegacyReturn(t *testing.T) {
	tests := []struct {
		name string
		resp WindowEventResp
		want int
	}{
		{name: "nil", resp: nil, want: 0},
		{name: "positive", resp: RawEventResp(17), want: 17},
		{name: "zero extended negative", resp: RawEventResp(uintptr(uint32(0xffffffff))), want: -1},
		{name: "native width negative", resp: RawEventResp(^uintptr(0)), want: -1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := EventRespInt(tc.resp); got != tc.want {
				t.Fatalf("EventRespInt(%v) = %d, want %d", tc.resp, got, tc.want)
			}
		})
	}
}
