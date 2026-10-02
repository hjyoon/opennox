//go:build !server

package opennox

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/cfg"
	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/env"
	"github.com/opennox/libs/strman"
	"github.com/spf13/viper"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/timer"
	"github.com/opennox/opennox/v1/server"
)

const optionsConfigChild = "NOX_TEST_OPTIONS_CONFIG_PHASE"

// Configuration/API tests, not queued GUI input. Every save, reload and reset
// phase is a fresh process with real client/server owners and a headless seat.
// E2E's deterministic bindings/write suppression cannot mask disk failures.
// Mandatory fixtures use generated configs in t.TempDir. The separate opt-in
// stock Reset test reads default.cfg and stages a private copy; it never writes
// stock or personal files. Reset's configuration layer and the C close/apply
// entry are checked separately from existing queued-input tests of GUI buttons.
func TestOptionsConfigDiskRoundTrip(t *testing.T) {
	for variant := 0; variant < 8; variant++ { // Every combination of audio mute flags.
		t.Run(strconv.Itoa(variant), func(t *testing.T) {
			dir := t.TempDir()
			for _, phase := range []string{"save", "load", "reset"} {
				runOptionsConfigChild(t, dir, phase, variant)
			}
		})
	}
}

func TestOptionsConfigFallbackAndGuards(t *testing.T) {
	for _, phase := range []string{"fallback", "missing", "malformed", "readonly", "e2e"} {
		t.Run(phase, func(t *testing.T) {
			runOptionsConfigChild(t, t.TempDir(), phase, 0)
		})
	}
}

func TestOptionsConfigStockDefaultReset(t *testing.T) {
	if os.Getenv("NOX_TEST_OPTIONS_STOCK_DEFAULT") == "" {
		t.Skip("opt-in read-only stock default.cfg; no original assets are bundled")
	}
	runOptionsConfigChild(t, t.TempDir(), "stock-reset", 0)
}

func runOptionsConfigChild(t *testing.T, dir, phase string, variant int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestOptionsConfigChild$", "-test.count=1", "-test.v")
	cmd.Dir = dir
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_OPTIONS_CONFIG_") || strings.HasPrefix(key, "NOX_E2E") || key == "NOX_DATA" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, optionsConfigChild+"="+phase, "NOX_TEST_OPTIONS_CONFIG_VARIANT="+strconv.Itoa(variant))
	if phase == "e2e" || phase == "video-apply-e2e" {
		cmd.Env = append(cmd.Env, "NOX_E2E=isolated-config-write-guard")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s subprocess failed: %v\n%s", phase, err, out)
	}
	if phase == "stock-reset" {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "read-only stock input Reset restored") {
				t.Log(strings.TrimSpace(line)) // Counts only, never stock config text.
			}
		}
	}
}

func TestOptionsConfigChild(t *testing.T) {
	phase := os.Getenv(optionsConfigChild)
	if phase == "" {
		t.Skip("only run as an isolated options configuration subprocess")
	}
	variant, err := strconv.Atoi(os.Getenv("NOX_TEST_OPTIONS_CONFIG_VARIANT"))
	if err != nil {
		t.Fatal(err)
	}
	if env.IsE2E() != (phase == "e2e" || phase == "video-apply-e2e") {
		t.Fatal("configuration phase used the wrong E2E mode")
	}
	legacy.InitBlobData()
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	noxServer = &Server{Server: s}
	c, err := NewClient(nil, nil, noxServer)
	if err != nil {
		t.Fatal(err)
	}
	noxClient = c
	t.Cleanup(c.Client.Close)
	t.Cleanup(c.GUI.DestroyAll)
	sc := headless.New(image.Pt(640, 480))
	t.Cleanup(func() { _ = sc.Close() })
	c.Seat = sc
	c.Win, err = render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	c.Inp = input.New(c.Log, sc, false, 0)
	keyBinding = keybind.New(s.Strings())
	c.videoSetGameMode(image.Pt(640, 480))
	if strings.HasPrefix(phase, "video-apply") {
		testOptionsVideoApply(t, c, sc, phase)
		return
	}
	var animation *gui.Anim
	if phase == "save" || phase == "e2e" {
		root := c.GUI.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
		root.SetID(300)
		animation = gui.NewAnim(root, image.Point{}, image.Pt(0, -480), image.Pt(0, 40), image.Pt(0, -40))
		animation.StateID = 300
		animation.Func12Ptr = legacy.OptionsStartOut4AA6B0()
		animation.SetState(gui.AnimInDone)
		gui.SetAnimGlobalState(gui.AnimInDone)
		legacy.OptionsStoreRoot4AA6B0(root)
		legacy.OptionsStoreAnimation4AA6B0(animation)
		t.Cleanup(animation.Free)
		t.Cleanup(func() {
			legacy.OptionsStoreRoot4AA6B0(nil)
			legacy.OptionsStoreAnimation4AA6B0(nil)
		})
		guiOptionsRes = image.Pt(800, 600)
	}

	seedWindowMode := variant % 3
	windowMode := []int{-3, -1, -2}[seedWindowMode]
	seed := fmt.Sprintf("Version = 65540\nGamma = 37\nVideoSize = 83\nInputSensitivity = 1.75\nFullscreen = %d\nStretched = %d\nFXVolume = 4096\nDialogVolume = 8192\nMusicVolume = 16384\n", seedWindowMode, variant&1)
	for _, key := range []string{"SoftShadowEdge", "TranslucentConsole", "RenderGlow", "FadeObjects", "DrawFrontWalls", "TranslucentFrontWalls", "HighResFrontWalls", "HighResFloors", "LockHighResFloors", "TexturedFloors", "RenderGUI"} {
		seed += fmt.Sprintf("%s = %d\n", key, variant&1)
	}
	seed += "---\nMousePickup = Right\nF9 = Jump\nF12 = ToggleInventory\nF10 + F11 = ToggleMap + MapZoomIn\n"
	write := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	read := func(path string, skip bool) {
		t.Helper()
		if err := nox_common_readcfgfile(path, skip); err != nil {
			t.Fatal(err)
		}
	}
	bindings := func() cfg.Section {
		var sect cfg.Section
		writeConfigHotkeys(&sect)
		return sect
	}
	checkKey := func(sect cfg.Section, key, want string) {
		t.Helper()
		if got, ok := sect.Get(key); !ok || got != want {
			t.Fatalf("%s = %q (present=%t), want %q", key, got, ok, want)
		}
	}
	if phase == "stock-reset" {
		path := os.Getenv("NOX_TEST_OPTIONS_STOCK_DEFAULT")
		if !filepath.IsAbs(path) {
			t.Fatal("stock default.cfg path must be absolute")
		}
		data, err := os.ReadFile(path) // Read-only; writes use the subprocess cwd.
		if err != nil {
			t.Fatal(err)
		}
		file, err := cfg.Parse(bytes.NewReader(data))
		if err != nil || len(file.Sections) != 2 {
			t.Fatalf("cannot parse the two stock default.cfg sections: %v", err)
		}
		write("default.cfg", string(data))
		if nox_client_parseConfigHotkeysLine("F9", "Jump") != 1 {
			t.Fatal("cannot seed a pending binding")
		}
		legacyGamma = 91
		configSetVolume(17, VolumeFX)
		read("default.cfg", true)
		got := bindings()
		var want cfg.Section
		for _, kv := range file.Sections[1] {
			if kv.Key == "" {
				continue // Config comments are not bindings.
			}
			if kv.Key == "MousePickup" {
				want.Set(kv.Key, kv.Value)
				continue
			}
			var keys, events []string
			for _, name := range strings.Split(kv.Key, "+") {
				key := keybind.KeyByName(strings.TrimSpace(name))
				if key == 0 {
					t.Fatalf("unknown stock key %q", name)
				}
				keys = append(keys, key.String())
			}
			for _, name := range strings.Split(kv.Value, "+") {
				event := keyBinding.EventByName(strings.TrimSpace(name))
				if event == nil || event.Event == 0 {
					t.Fatalf("unknown stock event %q", name)
				}
				events = append(events, event.Name)
			}
			want.Set(strings.Join(keys, " + "), strings.Join(events, " + "))
		}
		if len(got) != len(want) {
			t.Fatalf("stock reset produced %d bindings, want %d", len(got), len(want))
		}
		for _, kv := range want {
			checkKey(got, kv.Key, kv.Value)
		}
		if legacyGamma != 91 || configGetVolume(VolumeFX) != 17 {
			t.Fatal("stock input Reset parsed the general section")
		}
		if _, err := os.Stat("nox.cfg"); !os.IsNotExist(err) {
			t.Fatal("stock Reset wrote nox.cfg")
		}
		if source, err := os.ReadFile(path); err != nil || !bytes.Equal(source, data) {
			t.Fatal("stock default.cfg changed")
		}
		t.Logf("read-only stock input Reset restored %d bindings", len(got))
		return
	}
	if phase == "missing" || phase == "malformed" || phase == "fallback" {
		if phase != "missing" {
			write("default.cfg", seed)
		}
		if phase == "malformed" {
			write("nox.cfg", "Gamma = not-a-number\n---\nF9 = Jump\n")
		}
		err := nox_common_readcfgfile("nox.cfg", false)
		if phase == "fallback" {
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat("nox.cfg"); err != nil {
				t.Fatalf("default fallback did not create nox.cfg: %v", err)
			}
			checkKey(bindings(), "F9", "Jump")
			if legacyGamma != 37 {
				t.Fatal("fallback did not read the actual default.cfg")
			}
		} else if phase == "missing" {
			if !os.IsNotExist(err) {
				t.Fatalf("missing config error = %v", err)
			}
		} else {
			if err == nil || !strings.Contains(err.Error(), "Gamma") {
				t.Fatalf("malformed config did not fail: %v", err)
			}
			got, err := os.ReadFile("nox.cfg")
			if err != nil || !strings.Contains(string(got), "not-a-number") {
				t.Fatal("malformed existing config was overwritten by fallback")
			}
		}
		return
	}
	if phase == "readonly" {
		write("opennox.yml", "private yaml sentinel\n")
		configPath = filepath.Join(mustOptionsWorkdir(t), "opennox.yml")
		configReadOnly = true
		if err := writeConfig(); err != nil {
			t.Fatal(err)
		}
		if got, err := os.ReadFile("opennox.yml"); err != nil || string(got) != "private yaml sentinel\n" {
			t.Fatal("read-only YAML guard changed the file")
		}
		return
	}
	if phase == "e2e" {
		write("nox.cfg", "private legacy sentinel\n")
		write("opennox.yml", "private yaml sentinel\n")
		if got := animation.Func12(); got != 1 {
			t.Fatalf("C options close returned %d", got)
		}
		if sc.ScreenSize() != image.Pt(640, 480) || c.videoGetGameMode() != image.Pt(640, 480) {
			t.Fatal("E2E options close applied the pending resolution")
		}
		if animation.State() != gui.AnimOut || gui.AnimGlobalState() != gui.AnimOut {
			t.Fatal("E2E options close did not start the animation")
		}
		configPath = filepath.Join(mustOptionsWorkdir(t), "opennox.yml")
		configReadOnly = false
		if err := writeConfig(); err != nil {
			t.Fatal(err)
		}
		for _, file := range []struct{ name, want string }{{"nox.cfg", "private legacy sentinel\n"}, {"opennox.yml", "private yaml sentinel\n"}} {
			got, err := os.ReadFile(file.name)
			if err != nil || string(got) != file.want {
				t.Fatalf("E2E write guard changed %s", file.name)
			}
		}
		return
	}

	live := [3]int{4850, 8126, 12345}
	if variant&1 != 0 {
		live = [3]int{7, 1001, 16000}
	}
	wantVolume := live
	for i := range wantVolume {
		if variant&(1<<i) != 0 {
			wantVolume[i] = 0
		}
	}
	if phase == "save" {
		write("seed.cfg", seed)
		read("seed.cfg", false)
		// The live renderer mode is intentionally different from its save
		// checkpoint. Only the real close/apply path should synchronize it.
		c.UpdateFullScreen(windowMode)
		if g_fullscreen_cfg != seedWindowMode {
			t.Fatal("live window mode changed the save checkpoint too early")
		}
		for i, ptr := range []unsafe.Pointer{legacy.Get_dword_587000_127004(), legacy.Get_dword_587000_122852(), legacy.Get_dword_587000_93164()} {
			tmr := (*timer.Timer)(ptr)
			tmr.SetRaw(uint32(live[i]))
			tmr.Current = uint32(1+i) << 16 // Keep a genuinely pending raw update.
		}
		if variant&1 != 0 {
			legacy.Sub_453050()
		}
		if variant&2 != 0 {
			legacy.Sub_44D960()
		}
		if variant&4 != 0 {
			legacy.Sub_43DC00()
		}
		write("opennox.yml", "video:\n  stretch: false\n  filtering: true\noptions_test_unrelated: preserved\n")
		if err := readConfig("opennox.yml"); err != nil {
			t.Fatal(err)
		}
		c.Win.SetStretched(viper.GetBool(configVideoStretch))
		c.Win.SetFiltering(viper.GetBool(configVideoFiltering))
		c.SetStretch(true)
		c.ToggleFiltering()
		setGammaSliderOpts(73)
		viper.Set(configVideoWidth, 800)
		viper.Set(configVideoHeight, 600)
		writeConfigLater()
		maybeWriteConfig()
		// The existing writer deliberately saves Current (+4), not Target
		// (+8). Record this boundary before advancing the production timers;
		// an unserviced fixture is not evidence of a broken save operation.
		if err := writeConfigLegacy("pending.cfg"); err != nil {
			t.Fatal(err)
		}
		pendingData, err := os.ReadFile("pending.cfg")
		if err != nil {
			t.Fatal(err)
		}
		pending, err := cfg.Parse(bytes.NewReader(pendingData))
		if err != nil || len(pending.Sections) != 2 {
			t.Fatalf("pending config sections: %v, %v", pending, err)
		}
		checkKey(pending.Sections[0], "Fullscreen", strconv.Itoa(seedWindowMode))
		for i, ptr := range []unsafe.Pointer{legacy.Get_dword_587000_127004(), legacy.Get_dword_587000_122852(), legacy.Get_dword_587000_93164()} {
			want := i + 1
			if variant&(1<<i) != 0 {
				want = 0
			}
			checkKey(pending.Sections[0], []string{"FXVolume", "DialogVolume", "MusicVolume"}[i], strconv.Itoa(want))
			(*timer.TimerGroup)(ptr).Update()
		}
		// Unlike E2E, this isolated API fixture invokes the production C
		// animation callback with resolution changes and disk writes enabled.
		if sc.ScreenSize() != image.Pt(640, 480) || c.videoGetGameMode() != image.Pt(640, 480) {
			t.Fatal("resolution changed before options close/apply")
		}
		if got := animation.Func12(); got != 1 {
			t.Fatalf("C options close returned %d", got)
		}
		if sc.ScreenSize() != guiOptionsRes || c.videoGetGameMode() != guiOptionsRes || c.GetWindowMode() != windowMode || g_fullscreen_cfg != windowMode {
			t.Fatal("options close did not apply the size or retain the window mode")
		}
		if animation.State() != gui.AnimOut || gui.AnimGlobalState() != gui.AnimOut {
			t.Fatal("options close did not start the animation")
		}
		if configGetVolume(VolumeFX) != 4096 || configGetVolume(VolumeDialog) != 8192 || configGetVolume(VolumeMusic) != VolumeMax {
			t.Fatal("saving changed the startup volume scalars")
		}
	}
	data, err := os.ReadFile("nox.cfg")
	if err != nil {
		t.Fatal(err)
	}
	file, err := cfg.Parse(bytes.NewReader(data))
	if err != nil || len(file.Sections) != 2 {
		t.Fatalf("saved config sections: %v, %v", file, err)
	}
	for i, key := range []string{"FXVolume", "DialogVolume", "MusicVolume"} {
		checkKey(file.Sections[0], key, strconv.Itoa(wantVolume[i]))
	}
	checkKey(file.Sections[0], "VideoMode", "800 600 16")
	checkKey(file.Sections[0], "Fullscreen", strconv.Itoa(windowMode))
	checkKey(file.Sections[0], "Stretched", strconv.Itoa(variant&1))
	checkKey(file.Sections[1], "MousePickup", "Right")
	checkKey(file.Sections[1], "F10 + F11", "ToggleMap + MapZoomIn")
	if phase == "save" {
		return
	}
	if err := readConfig("opennox.yml"); err != nil {
		t.Fatal(err)
	}
	read("nox.cfg", false)
	// The real startup uses the YAML size/filtering/stretch after legacy read.
	c.Win.SetStretched(viper.GetBool(configVideoStretch))
	c.Win.SetFiltering(viper.GetBool(configVideoFiltering))
	videoUpdateGameMode(image.Pt(viper.GetInt(configVideoWidth), viper.GetInt(configVideoHeight)))
	if !c.GetStretch() || c.GetFiltering() || c.videoGetGameMode() != image.Pt(800, 600) || sc.ScreenSize() != image.Pt(800, 600) || viper.GetString("options_test_unrelated") != "preserved" || configDirty {
		t.Fatal("fresh process did not restore YAML settings/size or modified unrelated data")
	}
	if getGamma() != gammaMin+gammaMax*0.73 || legacyGamma != 37 || nox_video_getCutSize() != 83 || c.GetSensitivity() != 1.75 || c.GetWindowMode() != windowMode {
		t.Fatalf("legacy/YAML options did not load: gamma=%v legacy=%d cut=%d sensitivity=%v fullscreen=%d", getGamma(), legacyGamma, nox_video_getCutSize(), c.GetSensitivity(), c.GetWindowMode())
	}
	for i, want := range wantVolume {
		if got := configGetVolume(VolumeControl(i)); got != want {
			t.Fatalf("fresh volume %d = %d, want %d", i, got, want)
		}
	}
	for _, id := range []uint{2012, 2014, 2015, 2016, 2020, 2021, 2022, 2031, 2032, 2033, 2040} {
		if optionsAuditVideoValue(id) != (variant&1 != 0) {
			t.Fatalf("fresh video option %d did not round trip", id)
		}
	}
	checkKey(bindings(), "F9", "Jump")
	checkKey(bindings(), "F12", "ToggleInventory")
	if phase == "reset" {
		// A deliberately invalid general section proves skip=true does not
		// reset audio/video or silently use E2E's substituted bindings.
		write("default.cfg", "Gamma = must-not-parse\nFXVolume = must-not-parse\n---\nMousePickup = Left\nF10 = Jump\nF11 = ToggleInventory\n")
		read("default.cfg", true)
		reset := bindings()
		checkKey(reset, "MousePickup", "Left")
		checkKey(reset, "F10", "Jump")
		checkKey(reset, "F11", "ToggleInventory")
		if _, ok := reset.Get("F9"); ok {
			t.Fatal("Reset retained a previous binding")
		}
		if legacyGamma != 37 || configGetVolume(VolumeFX) != wantVolume[0] || c.GetSensitivity() != 1.75 || !c.GetStretch() || configDirty {
			t.Fatal("input Reset changed unrelated options")
		}
		for i, want := range wantVolume {
			if configGetVolume(VolumeControl(i)) != want {
				t.Fatalf("input Reset changed volume %d", i)
			}
		}
		for _, id := range []uint{2012, 2014, 2015, 2016, 2020, 2021, 2022, 2031, 2032, 2033, 2040} {
			if optionsAuditVideoValue(id) != (variant&1 != 0) {
				t.Fatalf("input Reset changed video option %d", id)
			}
		}
		if got, err := os.ReadFile("nox.cfg"); err != nil || !bytes.Equal(got, data) {
			t.Fatal("Reset wrote pending bindings over the saved config")
		}
		read("nox.cfg", true) // Cancel/reload restores the prior disk bindings.
		checkKey(bindings(), "F9", "Jump")
		checkKey(bindings(), "F12", "ToggleInventory")
	}
}

func mustOptionsWorkdir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
