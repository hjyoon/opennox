package legacy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPointSegmentProjectionOriginal427DF0NativeC(t *testing.T) {
	raw, err := os.ReadFile("testdata/point_segment_projection_427df0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Unchanged   bool   `json:"original_unchanged"`
		Startup     uint32 `json:"actual_startup_control_word"`
		Rows        []struct {
			Name   string   `json:"name"`
			Mode   string   `json:"mode"`
			Point  [2]int32 `json:"point"`
			Line   [4]int32 `json:"line"`
			Radius string   `json:"radius_bits"`
			Result int      `json:"result"`
			After  [2]int32 `json:"point_after"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if !fixture.Unchanged || fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Startup != 0xe3f || len(fixture.Rows) != 326 {
		t.Fatal("incomplete original full-projection fixture")
	}
	raw, err = os.ReadFile("GAME1_1.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	extract := func(marker string) string {
		t.Helper()
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
	definitions := []string{extract("int nox_float2int(float a1)"), extract("float nox_double2float(double a1)")}
	// Include each production dependency actually present in the translation
	// unit. Both before and after compile their real bodies against the same
	// independent original fixture; no expected value is derived from C.
	for _, marker := range []string{
		"static double nox_projection_add_chop53_427DF0(",
		"static double nox_projection_mul_chop53_427DF0(",
		"static double nox_projection_div_chop53_427DF0(",
		"static double nox_projection_sqrt_chop53_427DF0(",
		"static float nox_projection_spill_chop32_427DF0(",
	} {
		if strings.Contains(source, marker) {
			definitions = append(definitions, extract(marker))
		}
	}
	definitions = append(definitions, extract("int sub_427DF0(int2* point, int* a2, float a3)"))
	const harness = `
#include <stdint.h>
#include <inttypes.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <math.h>
#include <fenv.h>
typedef struct {int32_t field_0,field_4;} int2;
_Static_assert(sizeof(int)==4,"original signed DWORD coordinates");
// PRODUCTION_DEFINITIONS
int main(void){
 if(fegetround()!=FE_TONEAREST)return 3;
 int32_t px,py,line[4];uint32_t radius_bits;
 unsigned char canary[24];int2* point=malloc(32);if(!point)return 4;
 while(scanf("%" SCNd32 " %" SCNd32 " %" SCNd32 " %" SCNd32 " %" SCNd32 " %" SCNd32 " %" SCNx32,&px,&py,&line[0],&line[1],&line[2],&line[3],&radius_bits)==7){
  point->field_0=px;point->field_4=py;memset((char*)point+8,0xa5,24);
  int32_t saved_line[4];memcpy(saved_line,line,16);float radius;memcpy(&radius,&radius_bits,4);
  int result=sub_427DF0(point,line,radius);
  memset(canary,0xa5,24);
  if(memcmp((char*)point+8,canary,24)||memcmp(line,saved_line,16)||fegetround()!=FE_TONEAREST)abort();
  printf("%d %" PRId32 " %" PRId32 " %zu %" PRIxPTR "\n",result,point->field_0,point->field_4,sizeof(void*),(uintptr_t)point);
 }
 free(point);return ferror(stdin)?5:0;
}
`
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "point-segment-projection")
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-inline", "-x", "c", "-", "-lm", "-o", binary)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(strings.Replace(harness, "// PRODUCTION_DEFINITIONS", strings.Join(definitions, "\n"), 1))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile exact native projection definitions: %v\n%s", err, out)
	}
	var input strings.Builder
	selected := make([]int, 0, 163)
	for i, row := range fixture.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		selected = append(selected, i)
		fmt.Fprintf(&input, "%d %d %d %d %d %d %s\n", row.Point[0], row.Point[1], row.Line[0], row.Line[1], row.Line[2], row.Line[3], row.Radius)
	}
	if len(selected) != 163 {
		t.Fatalf("selected %d original startup observations", len(selected))
	}
	cmd = exec.Command(binary)
	cmd.Stdin = strings.NewReader(input.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execute complete native projection: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != len(selected) {
		t.Fatalf("native result rows=%d want=%d", len(lines), len(selected))
	}
	for j, i := range selected {
		row := fixture.Rows[i]
		t.Run(row.Name, func(t *testing.T) {
			var result, x, y, width int
			var ptr uint64
			if n, err := fmt.Sscanf(lines[j], "%d %d %d %d %x", &result, &x, &y, &width, &ptr); err != nil || n != 5 {
				t.Fatalf("native row: %q %v", lines[j], err)
			}
			if width != strconv.IntSize/8 || (width == 8 && ptr <= 0xffffffff) {
				t.Fatalf("native pointer width/address: %q", lines[j])
			}
			if result != row.Result || int32(x) != row.After[0] || int32(y) != row.After[1] {
				t.Fatalf("actual complete native projection result=%d point=(%d,%d), original result=%d point=%v", result, x, y, row.Result, row.After)
			}
		})
	}
}
