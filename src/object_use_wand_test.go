package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestWandUseConsumeCharge53F290(t *testing.T) {
	owner := new(server.Object)
	wand := new(server.Object)
	data := &server.WandUseData{Charge: 4, MaxCharge: 5, Progress: 80}
	wandUseConsumeCharge53F290(owner, wand, data)
	if data.Charge != 3 || data.Progress != 60 {
		t.Fatalf("charge/progress = %d/%d, want 3/60", data.Charge, data.Progress)
	}

	unlimited := &server.WandUseData{Charge: 7, Progress: 91}
	wandUseConsumeCharge53F290(owner, wand, unlimited)
	if unlimited.Charge != 7 || unlimited.Progress != 91 {
		t.Fatalf("unlimited charge/progress = %d/%d, want 7/91", unlimited.Charge, unlimited.Progress)
	}
}

func TestWandUse53F290AllowsMissingUseData(t *testing.T) {
	if !nox_xxx_useLesserFireballStaff_53F290(new(server.Object), new(server.Object)) {
		t.Fatal("missing use data result = false, want true")
	}
}
