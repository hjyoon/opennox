package server_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/internal/porttest/chainlightning"
	"github.com/opennox/opennox/v1/server"
)

func TestChainLightningLifecycleOriginal52F820(t *testing.T) {
	data, err := os.ReadFile("testdata/chain_lightning_lifecycle_original_52f820.json")
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
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			actual, err := chainlightning.Run(row.Input, func(kind string, record *server.DurSpell, rt server.SpellChainLightningRuntime52F820) uint32 {
				if kind == "create" {
					return uint32(server.SpellChainLightningCreate52F820(record, rt))
				}
				return uint32(server.SpellChainLightningDestroy530100(record, rt))
			})
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.Result, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("original lifecycle mismatch\ngot %s\nwant %s", actual, row.Result)
			}
		})
	}
}
