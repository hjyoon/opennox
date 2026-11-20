package legacy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type mapRuleSaveTestServer struct {
	Server
	world *server.Server
}

func (s mapRuleSaveTestServer) S() *server.Server { return s.world }

func TestMapRuleSaveSkipsUnknownWeaponName(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	t.Chdir(t.TempDir())
	dir := filepath.Join("maps", "crypt")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	oldServer := GetServer
	world := &server.Server{}
	GetServer = func() Server { return mapRuleSaveTestServer{world: world} }
	t.Cleanup(func() { GetServer = oldServer })

	var settings server.Settings2
	copy(settings.Field0[:], "crypt")
	for i := range settings.Field24.Vals {
		settings.Field24.Vals[i] = ^uint32(0)
	}
	settings.Field44 = [4]byte{0xef, 0xff, 0xff, 0xff}
	settings.Field48 = ^uint32(0)
	settings.Field52 = 0x110
	Sub_57AAA0("user.rul", &settings, nil)
	got, err := os.ReadFile(filepath.Join(dir, "user.rul"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "\n" {
		t.Fatalf("rule file contains an unnamed weapon: %q", got)
	}
}
