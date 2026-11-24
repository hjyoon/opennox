package legacy

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestMapgenProgressGoCallbackOrder(t *testing.T) {
	current, previous, sequence := uint32(0), uint32(5000), uint32(0xFFFF)
	var events []string
	mapgenUpdateProgress(0x9D, &current, &previous, mapgenProgressServices{
		ticks: func() uint64 {
			events = append(events, "ticks")
			return 300
		},
		poll: func() int32 {
			events = append(events, "poll")
			if current != 300 || previous != 0 {
				t.Fatalf("progress timing during input poll = %d/%d", current, previous)
			}
			current, previous = 400, 40
			return 0
		},
		threshold: func() int32 {
			events = append(events, "threshold")
			return 250
		},
		sequence: func() *uint32 {
			events = append(events, "sequence")
			return &sequence
		},
		display: func(packet [3]byte) {
			events = append(events, "display")
			if packet != [3]byte{0x9D, 0xFF, 0xFF} || previous != 40 || sequence != 0x10000 {
				t.Fatalf("progress state during display = %x, last %d, sequence %#x", packet, previous, sequence)
			}
			current, previous, sequence = 777, 888, 0xABCDEF12
		},
	})
	if strings.Join(events, ",") != "ticks,poll,threshold,sequence,display" || current != 777 || previous != 777 || sequence != 0xABCDEF12 {
		t.Fatalf("progress state after display = %v, timing %d/%d, sequence %#x", events, current, previous, sequence)
	}
}

func TestMapgenProgressGoHeadlessDoesNotTick(t *testing.T) {
	engine, ticks := noxflags.GetEngine(), PlatformTicks
	t.Cleanup(func() {
		PlatformTicks = ticks
		noxflags.ResetEngine()
		noxflags.SetEngine(engine)
	})
	noxflags.SetEngine(noxflags.EngineNoRendering)
	PlatformTicks = func() uint64 {
		t.Error("headless progress read the renderer clock")
		return 0
	}
	sequence := memmap.Uint32(0x5D4594, 2487572)
	nox_xxx_mapGenSetFlags_5235F0(-100)
	if got := memmap.Uint32(0x5D4594, 2487572); got != sequence {
		t.Fatalf("headless packet sequence = %#x, want %#x", got, sequence)
	}
}

func TestMapgenProgressNoRenderingR314(t *testing.T) {
	source, err := os.ReadFile("testdata/mapgen_progress_before_go.c")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	const start = "void nox_xxx_mapGenSetFlags_5235F0(char a1) {"
	if strings.Count(text, start) != 1 {
		t.Fatal("complete map-generation progress definition is not unique")
	}
	_, tail, _ := strings.Cut(text, start)
	// The top-level closing brace is unindented.
	const end = "\n}\n"
	body, _, ok := strings.Cut(tail, end)
	if !ok {
		t.Fatal("complete map-generation progress end is absent")
	}
	fixture, err := os.ReadFile("testdata/mapgen_progress_nodraw_r314_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "// COMPLETE_MAPGEN_PROGRESS_R314"
	if strings.Count(string(fixture), marker) != 1 {
		t.Fatal("complete progress insertion marker is not unique")
	}
	program := strings.Replace(string(fixture), marker, start+body+end, 1)
	for _, mode := range []struct {
		name  string
		flags []string
	}{
		{"O0", []string{"-O0"}},
		{"O2", []string{"-O2"}},
		{"sanitized", []string{"-O1", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-fno-omit-frame-pointer"}},
	} {
		t.Run(mode.name, func(t *testing.T) {
			if mode.name == "sanitized" && runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
				t.Skip("sanitizer execution requires the supported Darwin/Linux C runtime")
			}
			compiler := strings.Fields(os.Getenv("CC"))
			if len(compiler) == 0 {
				compiler = []string{"cc"}
			}
			binary := filepath.Join(t.TempDir(), "mapgen-progress")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			args := append([]string(nil), compiler[1:]...)
			args = append(args, "-std=c11", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-I.")
			args = append(args, mode.flags...)
			args = append(args, "-x", "c", "-", "-o", binary)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, compiler[0], args...)
			cmd.Stdin = strings.NewReader(program)
			output, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("compile complete mapgen progress: %v\n%s", err, output)
			}
			for _, scenario := range []string{"headless", "editor", "rendered"} {
				t.Run(scenario, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					output, err := exec.CommandContext(ctx, binary, scenario).CombinedOutput()
					if err != nil {
						t.Fatalf("complete mapgen progress %s: %v\n%s", scenario, err, output)
					}
					if !strings.Contains(string(output), "Mapgen progress "+scenario+": PASS") {
						t.Fatalf("missing complete mapgen progress receipt: %s", output)
					}
				})
			}
		})
	}
}
