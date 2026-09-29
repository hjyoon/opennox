package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestTradeP2PStartFailureMessage50EF10(t *testing.T) {
	tests := []struct {
		name       string
		result     server.TradeP2PStartResult50EF10
		want       string
		wantLoaded []string
	}{
		{name: "invalid", result: server.TradeP2PStartInvalid50EF10},
		{name: "same partner", result: server.TradeP2PStartSamePartner50EF10},
		{name: "complete", result: server.TradeP2PStartComplete50EF10},
		{name: "starter busy", result: server.TradeP2PStartStarterBusy50EF10, want: "starter is already trading", wantLoaded: []string{"StarterAlreadyTrading"}},
		{name: "other busy", result: server.TradeP2PStartOtherBusy50EF10, want: "Dun Mir is already trading", wantLoaded: []string{"OtherAlreadyTrading"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var loaded []string
			got := tradeP2PStartFailureMessage50EF10(test.result, "Dun Mir", func(key string) string {
				loaded = append(loaded, key)
				switch key {
				case "StarterAlreadyTrading":
					return "starter is already trading"
				case "OtherAlreadyTrading":
					return "%s is already trading"
				default:
					return ""
				}
			})
			if got != test.want || !reflect.DeepEqual(loaded, test.wantLoaded) {
				t.Fatalf("message = %q, loaded %v; want %q, %v", got, loaded, test.want, test.wantLoaded)
			}
		})
	}
}

func TestTradeP2PStartFailureMessageWithoutLoader50EF10(t *testing.T) {
	if got := tradeP2PStartFailureMessage50EF10(server.TradeP2PStartStarterBusy50EF10, "Dun Mir", nil); got != "" {
		t.Fatalf("message without loader = %q, want empty", got)
	}
}
