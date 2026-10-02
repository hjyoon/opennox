//go:build !server

package opennox

import (
	"bytes"
	"context"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/strman"
	"github.com/spf13/viper"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/server"
)

const optionsStretchChild = "NOX_TEST_OPTIONS_STRETCH_PHASE"

// These are API/configuration regressions, not GUI-input tests. Each phase is
// a fresh process so Viper overrides and env.IsE2E's cached value cannot make
// a missing disk write look like a successful load. Only a temporary YAML is
// read/written; no personal config, Save directory, or stock data is used.
func TestOptionsStretchConfigRoundTrip(t *testing.T) {
	for _, initial := range []bool{false, true} {
		t.Run(strconv.FormatBool(initial), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "opennox.yml")
			seed := []byte("video:\n  stretch: " + strconv.FormatBool(initial) + "\n  filtering: false\noptions_test_unrelated: preserved\n")
			if err := os.WriteFile(path, seed, 0o600); err != nil {
				t.Fatal(err)
			}
			runOptionsStretchChild(t, path, "save", !initial, false)
			runOptionsStretchChild(t, path, "load", !initial, false)
		})
	}
}

func TestOptionsStretchConfigWriteGuardsAndNil(t *testing.T) {
	for _, phase := range []string{"readonly", "e2e", "nil"} {
		t.Run(phase, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "opennox.yml")
			seed := []byte("video:\n  stretch: false\n  filtering: false\noptions_test_unrelated: preserved\n")
			if err := os.WriteFile(path, seed, 0o600); err != nil {
				t.Fatal(err)
			}
			runOptionsStretchChild(t, path, phase, true, phase == "e2e")
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, seed) {
				t.Fatalf("%s changed the configuration file: %q", phase, got)
			}
		})
	}
}

func runOptionsStretchChild(t *testing.T, path, phase string, want, e2e bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestOptionsStretchConfigChild$", "-test.count=1", "-test.v")
	cmd.Dir = filepath.Dir(path)
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_OPTIONS_STRETCH_") || key == "NOX_E2E" || key == "NOX_E2E_RECORD" || key == "NOX_E2E_OVERRIDE" || key == "NOX_DATA" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, optionsStretchChild+"="+phase, "NOX_TEST_OPTIONS_STRETCH_PATH="+path, "NOX_TEST_OPTIONS_STRETCH_VALUE="+strconv.FormatBool(want))
	if e2e {
		cmd.Env = append(cmd.Env, "NOX_E2E=isolated-config-write-guard")
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s subprocess failed: %v\n%s", phase, err, out)
	}
}

func TestOptionsStretchConfigChild(t *testing.T) {
	phase := os.Getenv(optionsStretchChild)
	if phase == "" {
		t.Skip("only run as an isolated configuration subprocess")
	}
	path := os.Getenv("NOX_TEST_OPTIONS_STRETCH_PATH")
	want, err := strconv.ParseBool(os.Getenv("NOX_TEST_OPTIONS_STRETCH_VALUE"))
	if err != nil {
		t.Fatal(err)
	}
	sc := headless.New(image.Pt(640, 480))
	t.Cleanup(func() { _ = sc.Close() })
	r, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	noxClient = &Client{Client: &client.Client{Seat: sc, Win: r}}
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	noxServer = &Server{Server: s}
	if err := readConfig(path); err != nil {
		t.Fatal(err)
	}
	if configDirty || configPath != path {
		t.Fatalf("existing config not clean: dirty=%t path=%q", configDirty, configPath)
	}
	// Mirror the native seat's final renderer initialization, without SDL or
	// a game loop. readConfig remains the real application configuration reader.
	r.SetStretched(viper.GetBool(configVideoStretch))
	if phase == "load" {
		if noxClient.GetStretch() != want || viper.GetBool(configVideoStretch) != want {
			t.Fatalf("fresh process loaded stretch=%t, want %t", noxClient.GetStretch(), want)
		}
	} else if phase == "nil" {
		var absent *Client
		absent.SetStretch(want)
		if configDirty || viper.GetBool(configVideoStretch) {
			t.Fatal("nil client scheduled or changed configuration")
		}
	} else {
		if noxClient.GetStretch() == want {
			t.Fatal("fixture did not require a real setting change")
		}
		noxClient.SetStretch(want)
		if !configDirty || noxClient.GetStretch() != want || viper.GetBool(configVideoStretch) != want {
			t.Fatalf("stretch change did not schedule save: dirty=%t live=%t yaml=%t", configDirty, noxClient.GetStretch(), viper.GetBool(configVideoStretch))
		}
		configReadOnly = phase == "readonly"
		maybeWriteConfig()
	}
	if viper.GetBool(configVideoFiltering) || viper.GetString("options_test_unrelated") != "preserved" {
		t.Fatal("stretch change/load modified an unrelated configuration value")
	}
}
