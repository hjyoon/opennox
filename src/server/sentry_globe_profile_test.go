package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestSentryGlobeOracleProfilesRemainSeparate(t *testing.T) {
	type row struct {
		Input  json.RawMessage `json:"input"`
		Output json.RawMessage `json:"output"`
		Events json.RawMessage `json:"events"`
	}
	type fixture struct {
		Rows          []row  `json:"rows"`
		Substitutions int    `json:"arithmetic_or_conversion_substitutions"`
		Profile       string `json:"independent_trig_profile_sha256"`
	}
	load := func(name string) fixture {
		blob, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		var f fixture
		if err = json.Unmarshal(blob, &f); err != nil {
			t.Fatal(err)
		}
		return f
	}
	raw, profiled := load("sentry_globe_original.json"), load("sentry_globe_profiled.json")
	if len(raw.Rows) != 630 || len(profiled.Rows) != 630 || raw.Substitutions != 0 || profiled.Substitutions != 1400 {
		t.Fatal("original/profiled provenance")
	}
	profile, err := os.ReadFile("testdata/sentry_globe_x87_profile.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(profile)) != profiled.Profile {
		t.Fatal("independent instruction profile hash")
	}
	changed := []string{}
	for n, r := range raw.Rows {
		p := profiled.Rows[n]
		if !reflect.DeepEqual(r.Input, p.Input) {
			t.Fatal("profile changed caller inputs")
		}
		if !reflect.DeepEqual(r.Output, p.Output) || !reflect.DeepEqual(r.Events, p.Events) {
			var in struct {
				Name string `json:"name"`
			}
			if err = json.Unmarshal(r.Input, &in); err != nil {
				t.Fatal(err)
			}
			changed = append(changed, in.Name)
		}
	}
	want := []string{"update/word/0/5effffff", "update/word/0/5f000000", "update/word/0/7f800000", "update/word/0/ff800000"}
	if !reflect.DeepEqual(changed, want) {
		t.Fatalf("independent x87 differences %v want %v", changed, want)
	}
}
