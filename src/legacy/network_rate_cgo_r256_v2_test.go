//go:build !server

package legacy

import "testing"

func TestNetworkTables256V2LiveCgoRate(t *testing.T) {
	original := networkTableOriginal256(t)
	installNetworkTableStock256(t, original)
	for _, row := range original.Rows {
		if row.Name != "rate" || row.Variant != "stock" {
			continue
		}
		for _, pair := range row.Cases {
			if got := Sub_40A710(int(int32(pair[0]))); got != pair[1] {
				t.Fatalf("rate %#x = %#x, GAME.EXE = %#x", pair[0], got, pair[1])
			}
		}
	}
}
