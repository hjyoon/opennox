//go:build amd64 || arm64

package legacy

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterActionNamesNativeDefinitionWire(t *testing.T) {
	for _, c := range monsterActionNameOriginalCases5345B0(t) {
		if c.Kind != "selected_name" {
			continue
		}
		var name string
		if err := json.Unmarshal(c.Result, &name); err != nil {
			t.Fatal(err)
		}
		for _, xor := range []bool{false, true} {
			t.Run(fmt.Sprintf("id-%08x/xor-%t", c.Value, xor), func(t *testing.T) {
				ud := new(server.MonsterUpdateData)
				ud.AIAction340 = c.Value
				// Keep the existing independently audited R237 definition
				// write boundaries. Only the final name varies, supplied by
				// the unchanged original 005345B0 execution above.
				chunks := [][]byte{{0}}
				for range 6 {
					chunks = append(chunks, make([]byte, 4))
				}
				chunks = append(chunks, xferStringChunks237("", true)...)
				chunks = append(chunks, make([]byte, 4))
				for range 10 {
					chunks = append(chunks, make([]byte, 2))
				}
				for range 3 {
					chunks = append(chunks, make([]byte, 4))
				}
				for range 3 {
					chunks = append(chunks, xferStringChunks237("SPELL_INVALID", true)...)
				}
				chunks = append(chunks, xferStringChunks237(name, true)...)
				checkXferChecksum237(t, xor, chunks, func(cf *cryptfile.CryptFile) error {
					return monsterXferDefinition528DB0(cf, ud, 64)
				})
			})
		}
	}
}
