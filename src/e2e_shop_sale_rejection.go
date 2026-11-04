package opennox

import (
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

// Read the actual three client message slots written by printCentered_445490.
// A server-side string lookup alone cannot establish what the player received.
func e2eShopRejectionMessages() []string {
	var messages []string
	for slot := uintptr(0); slot < 3; slot++ {
		units := memmap.PtrT[[318]uint16](0x5D4594, 823804+644*slot)[:]
		for i, unit := range units {
			if unit == 0 {
				units = units[:i]
				break
			}
		}
		messages = append(messages, string(utf16.Decode(units)))
	}
	return messages
}

func (sc *e2eScenario) AssertShopSaleRejection(key, name string) {
	var expected string
	sc.add(0, name+" translation", func() {
		entry, ok := noxClient.Strings().GetInFile(strman.ID(key), "Trade.c")
		if !ok || len(entry.Vals) != 1 || entry.Vals[0].Str == "" {
			e2eError(fmt.Errorf("stock shop rejection translation %q is unavailable", key))
			return
		}
		expected = entry.Vals[0].Str
	})
	sc.addWhen(0, name, 120, func() bool {
		if expected == "" || !strings.Contains(e2eConsoleLines(), expected) {
			return false
		}
		for _, message := range e2eShopRejectionMessages() {
			if message == expected {
				return true
			}
		}
		return false
	}, func() {
		active, _, _ := legacy.Nox_gui_itemAmountState()
		if active {
			e2eError(fmt.Errorf("rejected sale unexpectedly opened a quote"))
			return
		}
		e2eLog.Printf("SHOP SALE REJECTION: key=Trade.c:%s text=%q centered=true console=true", key, expected)
	})
	sc.CaptureMagicFrame(name)
}
