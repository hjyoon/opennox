package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"testing"
)

// Earlier scalar-service tests used nearest-control expectations. Keep their
// actor inputs and later effect checks, but obtain the common protection tail
// from independent whole-original startup observations, not production math.
var defaultDamageHistoricalResults4E0B30 = sync.OnceValues(func() (map[[2]uint64]int32, error) {
	raw, err := os.ReadFile("testdata/default_damage_elemental_historical_startup.json")
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != "bdcd9836c2030d498430e514f72569fa2ff944fae2ef6c76e78746b416b846d8" {
		return nil, fmt.Errorf("changed original historical elemental projections")
	}
	var fixture struct {
		Original      string `json:"original_sha256"`
		Contexts      int    `json:"whole_original_contexts"`
		Calls         int    `json:"whole_original_calls"`
		Substitutions int    `json:"numeric_or_protection_return_substitutions"`
		Results       []struct {
			Protection uint64 `json:"protection_bits"`
			Damage     int32  `json:"damage"`
			Result     int32  `json:"result"`
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		return nil, err
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		fixture.Contexts != 552 || fixture.Calls != 2208 || fixture.Substitutions != 0 || len(fixture.Results) != 552 {
		return nil, fmt.Errorf("incomplete historical elemental projections")
	}
	results := make(map[[2]uint64]int32)
	for _, row := range fixture.Results {
		key := [2]uint64{row.Protection, uint64(uint32(row.Damage))}
		if _, exists := results[key]; exists {
			return nil, fmt.Errorf("duplicate original elemental projection")
		}
		results[key] = row.Result
	}
	return results, nil
})

func defaultDamageOriginalProtected4E0B30(t *testing.T, damage int32, protection float64) int32 {
	t.Helper()
	if protection == 1 {
		// Retained declared scalar-service boundary beyond the real getter's
		// cap: its exact zero multiplier/minimum result is unchanged.
		return 1
	}
	results, err := defaultDamageHistoricalResults4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	result, exists := results[[2]uint64{math.Float64bits(protection), uint64(uint32(damage))}]
	if !exists {
		t.Fatalf("missing original startup projection: damage=%d protection=%016x", damage, math.Float64bits(protection))
	}
	return result
}
