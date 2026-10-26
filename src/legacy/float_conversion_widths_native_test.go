package legacy

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Compile exact production definitions; expectations come exclusively from
// independently executed original helpers and the complete ArrowTrap caller.
func TestFloatConversionWidthsOriginalNativeC(t *testing.T) {
	raw, err := os.ReadFile("testdata/float_conversion_widths_419a90_419b10.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		Unchanged      bool   `json:"original_unchanged"`
		Helpers        struct {
			Startup uint32 `json:"actual_startup_control_word"`
			Rows    []struct {
				Helper string `json:"helper"`
				Mode   string `json:"mode"`
				Input  string `json:"input_bits"`
				Result string `json:"result_bits"`
			} `json:"rows"`
		} `json:"helpers"`
		Caller struct {
			Rows []struct {
				Name    string `json:"name"`
				Mode    string `json:"mode"`
				X       string `json:"x_bits"`
				Y       string `json:"y_bits"`
				Variant int    `json:"variant"`
				Send    struct {
					Packet string `json:"packet"`
				} `json:"send"`
			} `json:"rows"`
		} `json:"caller"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if !fixture.Unchanged || fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Helpers.Startup != 0xe3f || len(fixture.Helpers.Rows) != 984 || len(fixture.Caller.Rows) != 156 {
		t.Fatal("original conversion-width observations are incomplete")
	}
	extract := func(file, marker string) string {
		t.Helper()
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(raw)
		if strings.Count(source, marker) != 1 {
			t.Fatalf("missing or ambiguous production definition %s", marker)
		}
		start := strings.Index(source, marker)
		end := start + strings.Index(source[start:], "{") + 1
		for depth := 1; depth != 0; end++ {
			if end >= len(source) {
				t.Fatalf("unterminated production definition %s", marker)
			}
			switch source[end] {
			case '{':
				depth++
			case '}':
				depth--
			}
		}
		return source[start:end]
	}
	var definitions []string
	for _, marker := range []string{
		"int nox_float2int(float a1)",
		"short nox_float2int16(float a1)",
		"short nox_float2int16_abs(float a1)",
		"int nox_double2int(double a1)",
	} {
		definitions = append(definitions, extract("GAME1_1.c", marker))
	}
	definitions = append(definitions, extract("GAME4_2.c", "void nox_xxx_sendArrowTrapFX_5238A0(float* a1, char a2)"))
	const harness = `
#include <math.h>
#include <stdint.h>
#include <inttypes.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <fenv.h>
typedef struct {float field_0,field_4;} float2;
static float* position;
static unsigned calls;
static unsigned char packet[6];
static int nox_xxx_netSendFxAllCli_523030(float2* pos,void* data,int size) {
    if ((void*)pos!=(void*)position || size!=6 || calls++) abort();
    memcpy(packet,data,6);
    return 0;
}
// EXACT_PRODUCTION_DEFINITIONS
int main(int argc,char** argv) {
    if (argc!=5 || fegetround()!=FE_TONEAREST) return 3;
    int mode=atoi(argv[1]);
    uint64_t bits=strtoull(argv[2],NULL,16);
    uint32_t low=(uint32_t)bits;
    float f;double d;
    memcpy(&f,&low,4);memcpy(&d,&bits,8);
    uint32_t result;
    switch(mode) {
    case 0: result=(uint16_t)nox_float2int16(f); break;
    case 1: result=(uint16_t)nox_float2int16_abs(f); break;
    case 2: result=(uint32_t)nox_double2int(d); break;
    case 3: {
        uint32_t y=(uint32_t)strtoul(argv[3],NULL,16);
        unsigned char before[24];
        position=malloc(24);if (!position) return 4;
        memcpy(position,&low,4);memcpy(position+1,&y,4);
        memset((char*)position+8,0x5a,16);memcpy(before,position,24);
        nox_xxx_sendArrowTrapFX_5238A0(position,(char)atoi(argv[4]));
        if (calls!=1 || memcmp(before,position,24)) abort();
        printf("%zu %" PRIxPTR " ",sizeof(void*),(uintptr_t)position);
        for(unsigned i=0;i<6;i++) printf("%02x",packet[i]);
        puts("");free(position);return 0;
    }
    default:return 5;
    }
    printf("%08" PRIx32 "\n",result);
    return 0;
}
`
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "native-conversion-widths")
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-inline", "-x", "c", "-", "-lm", "-o", binary)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(strings.Replace(harness, "// EXACT_PRODUCTION_DEFINITIONS", strings.Join(definitions, "\n"), 1))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile exact production definitions: %v\n%s", err, output)
	}
	for _, group := range []struct {
		name   string
		helper string
		mode   string
		count  int
	}{
		{"word", "float-to-DWORD-low-WORD", "0", 39},
		{"absolute-word", "fabs-float-to-DWORD-low-WORD", "1", 39},
		{"double-qword-low-dword", "double-to-signed-QWORD-low-DWORD", "2", 45},
	} {
		t.Run(group.name, func(t *testing.T) {
			selected := 0
			for _, row := range fixture.Helpers.Rows {
				if row.Helper != group.helper || row.Mode != "actual-startup-inherited" {
					continue
				}
				selected++
				t.Run(row.Input, func(t *testing.T) {
					output, err := exec.Command(binary, group.mode, row.Input, "0", "0").CombinedOutput()
					if err != nil {
						t.Fatalf("native production conversion: %v\n%s", err, output)
					}
					got, err := strconv.ParseUint(strings.TrimSpace(string(output)), 16, 32)
					want, wantErr := strconv.ParseUint(row.Result, 16, 32)
					if err != nil || wantErr != nil || got != want {
						t.Fatalf("native production conversion input=%s got=%s original=%s parse=%v/%v", row.Input, strings.TrimSpace(string(output)), row.Result, err, wantErr)
					}
				})
			}
			if selected != group.count {
				t.Fatalf("selected %d original observations, want %d", selected, group.count)
			}
		})
	}
	t.Run("arrow-trap-packet", func(t *testing.T) {
		selected := 0
		for _, row := range fixture.Caller.Rows {
			if row.Mode != "actual-startup-inherited" {
				continue
			}
			selected++
			t.Run(row.Name, func(t *testing.T) {
				output, err := exec.Command(binary, "3", row.X, row.Y, strconv.Itoa(row.Variant)).CombinedOutput()
				if err != nil {
					t.Fatalf("native complete ArrowTrap packet caller: %v\n%s", err, output)
				}
				fields := strings.Fields(string(output))
				if len(fields) != 3 {
					t.Fatalf("native packet diagnostic: %s", output)
				}
				width, widthErr := strconv.Atoi(fields[0])
				pointer, ptrErr := strconv.ParseUint(fields[1], 16, 64)
				if widthErr != nil || ptrErr != nil || width != strconv.IntSize/8 || (width == 8 && pointer <= 0xffffffff) {
					t.Fatalf("native source pointer width/address: %s", output)
				}
				if fields[2] != row.Send.Packet {
					t.Fatalf("native complete ArrowTrap packet got=%s original=%s", fields[2], row.Send.Packet)
				}
			})
		}
		if selected != 39 {
			t.Fatalf("selected %d original packet observations, want 39", selected)
		}
	})
}
