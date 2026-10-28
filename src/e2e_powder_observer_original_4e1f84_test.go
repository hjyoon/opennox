package opennox

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
)

func TestE2EPowderObserverOriginalArmor4E1F84(t *testing.T) {
	b, err := os.ReadFile("server/testdata/powder_observer_armor_original_4e1f84.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Cls          uint32
			Raw          int32
			Armor, Carry uint32
			Control      uint32 `json:"control_word"`
			Effective    int32
			Remainder    uint32
		}
	}
	if err := json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 480 {
		t.Fatal("seeded original prefix count")
	}
	actual, nearest := 0, 0
	for _, row := range fixture.Rows {
		if row.Control == 0x23f {
			nearest++
			continue
		}
		if row.Control != 0xe3f {
			t.Fatal("original control word")
		}
		actual++
		t.Run(fmt.Sprintf("class%d/raw%d/%08x/%08x", row.Cls, row.Raw, row.Armor, row.Carry), func(t *testing.T) {
			for _, immune := range []bool{false, true} {
				damage, carry := e2ePowderBarrelUnitDamage(row.Raw, true, immune, math.Float32frombits(row.Armor), math.Float32frombits(row.Carry))
				want := row.Effective
				if immune {
					want /= 2
				}
				if damage != want || carry != row.Remainder {
					t.Fatalf("independent observer = %d/%08x, original %d/%08x", damage, carry, want, row.Remainder)
				}
			}
		})
	}
	if actual != 240 || nearest != 240 {
		t.Fatal("separate actual/nearest context counts")
	}
}
