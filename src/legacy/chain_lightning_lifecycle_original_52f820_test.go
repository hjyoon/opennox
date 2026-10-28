package legacy

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/internal/porttest/chainlightning"
	"github.com/opennox/opennox/v1/server"
)

type chainLightningOriginalServer52F820 struct {
	Server
	runtime server.SpellChainLightningRuntime52F820
}

func (s *chainLightningOriginalServer52F820) SpellChainLightningCreate52F820(record *server.DurSpell) int32 {
	return server.SpellChainLightningCreate52F820(record, s.runtime)
}

func (s *chainLightningOriginalServer52F820) SpellChainLightningDestroy530100(record *server.DurSpell) uint8 {
	return server.SpellChainLightningDestroy530100(record, s.runtime)
}

func TestChainLightningLifecycleOriginalNativeC52F820(t *testing.T) {
	data, err := os.ReadFile("../server/testdata/chain_lightning_lifecycle_original_52f820.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture chainlightning.Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 56 {
		t.Fatal("original fixture identity/count")
	}
	fake := &chainLightningOriginalServer52F820{}
	old := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = old })
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			actual, err := chainlightning.Run(row.Input, func(kind string, record *server.DurSpell, rt server.SpellChainLightningRuntime52F820) uint32 {
				fake.runtime = rt
				if kind == "create" {
					return chainLightningLifecycleCreateC52F820(record)
				}
				return chainLightningLifecycleDestroyC530100(record)
			})
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.Result, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("original native lifecycle mismatch\ngot %s\nwant %s", actual, row.Result)
			}
		})
	}
}
