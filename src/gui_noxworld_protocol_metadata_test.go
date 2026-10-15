package opennox

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/opennox/opennox/v1/common/discover"
)

func TestNoxWorldSelectionUsesObservedProtocolWithoutGuessing(t *testing.T) {
	oldHost := clientGetServerHost()
	t.Cleanup(func() { noxWorldClearSelectedServer(); clientSetServerHost(oldHost) })
	for _, version := range []uint32{0x0001039a, 0x000f039a, 0, 0x8001039a, 0xffffffff} {
		for _, res := range []string{`{}`, `{"high_res":true,"width":1920,"height":1080}`, `{"width":640,"height":480}`} {
			t.Run(fmt.Sprintf("%08x/%s", version, res), func(t *testing.T) {
				var srv discover.Server
				data := fmt.Sprintf(`{"name":"ProtocolHost","addr":"127.0.0.1","port":18590,"map":"estate","res":%s,"protocol_version":%d}`, res, version)
				if err := json.Unmarshal([]byte(data), &srv); err != nil {
					t.Fatal(err)
				}
				if err := noxWorldSetSelectedServer(srv); err != nil {
					t.Fatal(err)
				}
				if got := noxWorldSelected.Version(); got != version {
					t.Fatalf("selected protocol = %#08x, want observed %#08x (res=%s)", got, version, res)
				}
				if noxWorldSelected.ServerName() != "ProtocolHost" || noxWorldSelected.Port() != 18590 || clientGetServerHost() != "127.0.0.1" {
					t.Fatal("selected endpoint/name changed")
				}
			})
		}
	}
}

func TestNoxWorldSelectionKeepsResolutionFallbackWhenProtocolUnknown(t *testing.T) {
	oldHost := clientGetServerHost()
	t.Cleanup(func() { noxWorldClearSelectedServer(); clientSetServerHost(oldHost) })
	for _, metadata := range []string{"", `,"protocol_version":null`} {
		for _, tc := range []struct {
			res  string
			want uint32
		}{
			{`{}`, 0x0001039a},
			{`{"high_res":true}`, 0x000f039a},
			{`{"width":1025}`, 0x000f039a},
			{`{"height":769}`, 0x000f039a},
			{`{"width":1024,"height":768}`, 0x0001039a},
		} {
			t.Run(metadata+tc.res, func(t *testing.T) {
				var srv discover.Server
				data := `{"name":"FallbackHost","addr":"127.0.0.1","port":18590,"res":` + tc.res + metadata + `}`
				if err := json.Unmarshal([]byte(data), &srv); err != nil {
					t.Fatal(err)
				}
				if err := noxWorldSetSelectedServer(srv); err != nil {
					t.Fatal(err)
				}
				if got := noxWorldSelected.Version(); got != tc.want {
					t.Fatalf("fallback protocol = %#08x, want %#08x", got, tc.want)
				}
			})
		}
	}
}
