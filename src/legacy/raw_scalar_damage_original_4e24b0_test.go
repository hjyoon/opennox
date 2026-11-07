//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"unsafe"
)

const rawScalarDamageReferenceSHA4E24B0 = "423bf6344e1874539b475417161a2f620e069dfddb987880938e2c5138ec85de"
const rawScalarDamageInputSHA4E24B0 = "fc910f9ba3e307d095e5f349c233ffa744ed82eb48f3da508662b33f856d701e"

var rawScalarDamageSymbols4E24B0 = []string{"sub_4E24B0", "sub_4E24E0", "nox_xxx_damageFlammable_4E2520", "nox_xxx_damageBlackPowder_4E2560"}

type rawScalarDamageInput4E24B0 struct {
	Name     string   `json:"name"`
	Function string   `json:"function"`
	Args     []uint32 `json:"args_u32"`
	Parent   struct {
		Response map[string]uint32 `json:"response"`
	} `json:"external_parent"`
}
type rawScalarDamageRow4E24B0 struct {
	Name               string `json:"name"`
	Function           string `json:"function"`
	CW                 uint32 `json:"FPU_control"`
	InputSHA           string `json:"input_case_sha256"`
	Status             string `json:"status"`
	Valid              bool   `json:"certificate_valid"`
	Result             uint32 `json:"actual_result_u32"`
	ParentInstructions int    `json:"actual_parent_original_instruction_count"`
	ParentCredit       int    `json:"parent_body_execution_credit"`
	Unused             bool   `json:"declared_external_response_unused"`
	Consumed           bool   `json:"declared_external_response_consumed"`
	Services           []struct {
		Site           uint32            `json:"CALL_site"`
		Args           []uint32          `json:"actual_arguments_u32"`
		Response       map[string]uint32 `json:"declared_response_input"`
		ParentExecuted bool              `json:"actual_parent_instruction_executed"`
	} `json:"actual_external_services"`
}
type rawScalarDamageData4E24B0 struct {
	Inputs   map[string]rawScalarDamageInput4E24B0
	InputSHA map[string]string
	Rows     []rawScalarDamageRow4E24B0
}

func rawScalarDamageReadGzip4E24B0(t *testing.T, path, digest string) *gzip.Reader {
	t.Helper()
	packed, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(packed)) != digest {
		t.Fatal("authentic frozen reference/input bytes", path)
	}
	stream, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stream.Close() })
	return stream
}

func rawScalarDamageLoad4E24B0(t *testing.T) *rawScalarDamageData4E24B0 {
	t.Helper()
	base := "testdata/retained_scalar_four_4e24b0/"
	inputStream := rawScalarDamageReadGzip4E24B0(t, base+"inputs-v2.json.gz", rawScalarDamageInputSHA4E24B0)
	var chosen struct {
		Cases []json.RawMessage `json:"cases"`
	}
	if err := json.NewDecoder(io.LimitReader(inputStream, 2<<20)).Decode(&chosen); err != nil {
		t.Fatal(err)
	}
	data := &rawScalarDamageData4E24B0{Inputs: make(map[string]rawScalarDamageInput4E24B0), InputSHA: make(map[string]string)}
	for _, raw := range chosen.Cases {
		var input rawScalarDamageInput4E24B0
		if err := json.Unmarshal(raw, &input); err != nil {
			t.Fatal(err)
		}
		if _, exists := data.Inputs[input.Name]; exists || len(input.Args) != 5 {
			t.Fatal("unique complete five-word input")
		}
		data.Inputs[input.Name] = input
		data.InputSHA[input.Name] = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	if len(data.Inputs) != 215 {
		t.Fatal("complete independently chosen215 inputs")
	}
	stream := rawScalarDamageReadGzip4E24B0(t, base+"reference-v1.json.gz", rawScalarDamageReferenceSHA4E24B0)
	limit := &io.LimitedReader{R: stream, N: 128 << 20}
	decoder := json.NewDecoder(limit)
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		t.Fatal("original document", err)
	}
	var schema string
	var source struct {
		PE           string `json:"original_EXE_sha256"`
		Count        int    `json:"input_count"`
		ParentCredit int    `json:"parent_original_body_execution_credit"`
		Bytes        int    `json:"complete_original_caller_bytes"`
		Bindings     map[string]struct {
			SHA string `json:"sha256"`
		} `json:"bindings"`
	}
	var fatal json.RawMessage
	keys := make(map[string]bool)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			t.Fatal(err)
		}
		key := keyToken.(string)
		if keys[key] {
			t.Fatal("duplicate original top field")
		}
		keys[key] = true
		switch key {
		case "schema":
			err = decoder.Decode(&schema)
		case "source":
			err = decoder.Decode(&source)
		case "fatal_recording_error":
			err = decoder.Decode(&fatal)
		case "rows":
			var delimiter json.Token
			delimiter, err = decoder.Token()
			if err != nil || delimiter != json.Delim('[') {
				t.Fatal("original row array", err)
			}
			for decoder.More() {
				var row rawScalarDamageRow4E24B0
				if err := decoder.Decode(&row); err != nil {
					t.Fatal(err)
				}
				data.Rows = append(data.Rows, row)
			}
			delimiter, err = decoder.Token()
			if err != nil || delimiter != json.Delim(']') {
				t.Fatal("row array end", err)
			}
		default:
			t.Fatal("unknown original document field", key)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		t.Fatal("original end", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF || limit.N <= 0 {
		t.Fatal("bounded original EOF", err)
	}
	if schema != "opennox-retained-scalar-four-original-boundary-reference-v1" || source.PE != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || source.Count != 215 || source.ParentCredit != 0 || source.Bytes != 196 || source.Bindings["inputs"].SHA != rawScalarDamageInputSHA4E24B0 || len(data.Rows) != 430 || string(fatal) != "null" {
		t.Fatal("complete actual original430 source binding")
	}
	seen := make(map[string]bool)
	consumed, unused := 0, 0
	for _, row := range data.Rows {
		input, exists := data.Inputs[row.Name]
		key := fmt.Sprintf("%x/%s", row.CW, row.Name)
		if !exists || seen[key] || (row.CW != 0xe3f && row.CW != 0x23f) || row.InputSHA != data.InputSHA[row.Name] || row.Function != input.Function || row.Status != "returned" || !row.Valid || row.ParentInstructions != 0 || row.ParentCredit != 0 || len(row.Services) > 1 || row.Consumed != (len(row.Services) == 1) || row.Unused != (len(row.Services) == 0) {
			t.Fatal("exact actual original row certificate", key)
		}
		seen[key] = true
		if row.Consumed {
			consumed++
		} else {
			unused++
		}
		for _, service := range row.Services {
			if len(service.Args) != 5 || service.ParentExecuted || service.Response["eax_u32"] != input.Parent.Response["eax_u32"] {
				t.Fatal("actual external frame/raw response")
			}
		}
	}
	if len(seen) != 430 || consumed != 376 || unused != 54 {
		t.Fatal("actual observed430/376boundary/54unused")
	}
	return data
}

func TestRawScalarDamageOriginalReference4E24B0(t *testing.T) { rawScalarDamageLoad4E24B0(t) }

type rawScalarDamagePool4E24B0 struct {
	Pointers map[uint32]unsafe.Pointer
	Bases    []unsafe.Pointer
	Before   [][]byte
}

func rawScalarDamagePoolNew4E24B0(t *testing.T, words []uint32) *rawScalarDamagePool4E24B0 {
	t.Helper()
	pool := &rawScalarDamagePool4E24B0{Pointers: map[uint32]unsafe.Pointer{0: nil}}
	t.Cleanup(func() {
		for _, base := range pool.Bases {
			rawScalarDamageFixtureFree4E24B0(base)
		}
	})
	for _, word := range words {
		if _, exists := pool.Pointers[word]; exists {
			continue
		}
		base := rawScalarDamageFixtureAlloc4E24B0()
		if base == nil {
			t.Fatal("C-owned opaque canary allocation")
		}
		pool.Bases = append(pool.Bases, base)
		frame := unsafe.Slice((*byte)(base), 32)
		for index := range frame {
			frame[index] = byte(word + uint32(index)*17)
		}
		pool.Before = append(pool.Before, append([]byte(nil), frame...))
		pointer := unsafe.Add(base, uintptr(word&3))
		if uintptr(pointer) <= 0xffffffff || uintptr(pointer) == uintptr(rawScalarDamageFixtureBadU324E24B0(pointer)) {
			t.Fatal("actual high native pointer/truncation negative control")
		}
		pool.Pointers[word] = pointer
	}
	return pool
}
func (pool *rawScalarDamagePool4E24B0) compare(t *testing.T) {
	t.Helper()
	for index, base := range pool.Bases {
		if !bytes.Equal(unsafe.Slice((*byte)(base), 32), pool.Before[index]) {
			t.Fatal("all native opaque canaries unchanged")
		}
	}
}
func rawScalarDamageIndex4E24B0(t *testing.T, symbol string) int {
	t.Helper()
	for index, name := range rawScalarDamageSymbols4E24B0 {
		if name == symbol {
			return index
		}
	}
	t.Fatal("unknown original function", symbol)
	return -1
}
func rawScalarDamageCall4E24B0(backend string, index int, target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	if backend == "C" {
		return rawScalarDamageFixtureCall4E24B0(index, target, source, weapon, damage, typ)
	}
	switch index {
	case 0:
		return rawStoneDamageNative4E24B0(target, source, weapon, damage, typ)
	case 1:
		return rawMechGolemDamageNative4E24E0(target, source, weapon, damage, typ)
	case 2:
		return rawFlammableDamageNative4E2520(target, source, weapon, damage, typ)
	case 3:
		return rawBlackPowderDamageNative4E2560(target, source, weapon, damage, typ)
	default:
		panic("unknown scalar test entry")
	}
}

type rawScalarDamageObserved4E24B0 struct {
	Site                   uint32
	Target, Source, Weapon unsafe.Pointer
	Damage, Type           int32
}

func rawScalarDamageCheck4E24B0(t *testing.T, row rawScalarDamageRow4E24B0, calls []rawScalarDamageObserved4E24B0, pointers map[uint32]unsafe.Pointer) {
	t.Helper()
	if len(calls) != len(row.Services) {
		t.Fatal("actual CALL admission differs from original", len(calls), len(row.Services))
	}
	for index, call := range calls {
		original := row.Services[index]
		for _, word := range original.Args[:3] {
			if _, exists := pointers[word]; !exists {
				t.Fatal("original pointer token outside independent input")
			}
		}
		if call.Site != original.Site || call.Target != pointers[original.Args[0]] || call.Source != pointers[original.Args[1]] || call.Weapon != pointers[original.Args[2]] || uint32(call.Damage) != original.Args[3] || uint32(call.Type) != original.Args[4] {
			t.Fatal("actual five arguments differ from original frame", call, original.Args)
		}
	}
}
func rawScalarDamageDuringParent4E24B0(observe func(uint32, unsafe.Pointer, unsafe.Pointer, unsafe.Pointer, int32, int32) int32, run func() int32) int32 {
	previous := rawScalarDamageObserveParent4E24B0
	rawScalarDamageObserveParent4E24B0 = observe
	defer func() { rawScalarDamageObserveParent4E24B0 = previous }()
	return run()
}

func TestRawScalarDamageOriginalConsumers4E24B0(t *testing.T) {
	data := rawScalarDamageLoad4E24B0(t)
	if rawScalarDamageObserveParent4E24B0 != nil {
		t.Fatal("diagnostic seam must start nil")
	}
	for _, row := range data.Rows {
		for _, backend := range []string{"Go", "C"} {
			t.Run(fmt.Sprintf("%x/%s/%s", row.CW, backend, row.Name), func(t *testing.T) {
				input := data.Inputs[row.Name]
				pool := rawScalarDamagePoolNew4E24B0(t, input.Args[:3])
				index := rawScalarDamageIndex4E24B0(t, input.Function)
				if rawScalarDamageFixtureAddress4E24B0(index) <= 0xffffffff {
					t.Fatal("actual typed C entry above4GiB")
				}
				var calls []rawScalarDamageObserved4E24B0
				result := rawScalarDamageDuringParent4E24B0(func(site uint32, target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
					calls = append(calls, rawScalarDamageObserved4E24B0{site, target, source, weapon, damage, typ})
					return int32(input.Parent.Response["eax_u32"])
				}, func() int32 {
					return rawScalarDamageCall4E24B0(backend, index, pool.Pointers[input.Args[0]], pool.Pointers[input.Args[1]], pool.Pointers[input.Args[2]], int32(input.Args[3]), int32(input.Args[4]))
				})
				if uint32(result) != row.Result {
					t.Fatal("actual raw signed EAX differs from original", result, row.Result)
				}
				rawScalarDamageCheck4E24B0(t, row, calls, pool.Pointers)
				pool.compare(t)
				if rawScalarDamageObserveParent4E24B0 != nil {
					t.Fatal("seam not restored")
				}
				t.Logf("RAW_SCALAR_DAMAGE_ORIGINAL_OK backend=%s case=%s mode=%x function=%s calls=%d result=%08x target=%p source=%p weapon=%p entry=%#x parent-boundary=true all-canaries=true", backend, row.Name, row.CW, row.Function, len(calls), uint32(result), pool.Pointers[input.Args[0]], pool.Pointers[input.Args[1]], pool.Pointers[input.Args[2]], rawScalarDamageFixtureAddress4E24B0(index))
			})
		}
	}
}

func TestRawScalarDamageSeamReentry4E24B0(t *testing.T) {
	data := rawScalarDamageLoad4E24B0(t)
	find := func(name string) rawScalarDamageRow4E24B0 {
		for _, row := range data.Rows {
			if row.Name == name && row.CW == 0xe3f {
				return row
			}
		}
		t.Fatal("missing actual nested source row", name)
		return rawScalarDamageRow4E24B0{}
	}
	outer := find("StoneDamage/opaque-pointer-ABI/all-nil")
	inner := find("FlammableDamage/opaque-pointer-ABI/all-nil")
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			previous := rawScalarDamageObserveParent4E24B0
			defer func() { rawScalarDamageObserveParent4E24B0 = previous }()
			outerInput, innerInput := data.Inputs[outer.Name], data.Inputs[inner.Name]
			var outerCalls, innerCalls, priorCalls []rawScalarDamageObserved4E24B0
			var innerResult int32
			invoke := func(row rawScalarDamageRow4E24B0, input rawScalarDamageInput4E24B0) int32 {
				return rawScalarDamageCall4E24B0(backend, rawScalarDamageIndex4E24B0(t, row.Function), nil, nil, nil, int32(input.Args[3]), int32(input.Args[4]))
			}
			rawScalarDamageObserveParent4E24B0 = func(site uint32, target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
				priorCalls = append(priorCalls, rawScalarDamageObserved4E24B0{site, target, source, weapon, damage, typ})
				return int32(outerInput.Parent.Response["eax_u32"])
			}
			result := rawScalarDamageDuringParent4E24B0(func(site uint32, target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
				outerCalls = append(outerCalls, rawScalarDamageObserved4E24B0{site, target, source, weapon, damage, typ})
				if len(outerCalls) == 1 {
					innerResult = rawScalarDamageDuringParent4E24B0(func(site uint32, target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
						innerCalls = append(innerCalls, rawScalarDamageObserved4E24B0{site, target, source, weapon, damage, typ})
						return int32(innerInput.Parent.Response["eax_u32"])
					}, func() int32 { return invoke(inner, innerInput) })
				}
				return int32(outerInput.Parent.Response["eax_u32"])
			}, func() int32 {
				first := invoke(outer, outerInput)
				second := invoke(outer, outerInput)
				if first != second {
					t.Fatalf("reentered caller raw returns differ: %08x vs %08x", uint32(first), uint32(second))
				}
				return first
			})
			probe := invoke(outer, outerInput)
			if uint32(result) != outer.Result || uint32(innerResult) != inner.Result || uint32(probe) != outer.Result || len(outerCalls) != 2 || len(innerCalls) != 1 || len(priorCalls) != 1 {
				t.Fatal("nested actual entry/seam save-restore results or counts")
			}
			nilPointers := map[uint32]unsafe.Pointer{0: nil}
			for _, call := range outerCalls {
				rawScalarDamageCheck4E24B0(t, outer, []rawScalarDamageObserved4E24B0{call}, nilPointers)
			}
			rawScalarDamageCheck4E24B0(t, inner, innerCalls, nilPointers)
			rawScalarDamageCheck4E24B0(t, outer, priorCalls, nilPointers)
			t.Logf("RAW_SCALAR_DAMAGE_REENTRY_OK backend=%s outer=2 inner=1 prior=1 restored=true", backend)
		})
	}
}

func TestRawScalarDamageDefaultPublicNilTarget4E24B0(t *testing.T) {
	if rawScalarDamageObserveParent4E24B0 != nil || rawDefaultDamageNativeObserverStack4E0B30 != nil {
		t.Fatal("default path requires all observers nil")
	}
	for index, symbol := range rawScalarDamageSymbols4E24B0 {
		t.Run(symbol, func(t *testing.T) {
			pool := rawScalarDamagePoolNew4E24B0(t, []uint32{0x21000100, 0x21000200})
			result := rawScalarDamageFixtureCall4E24B0(index, nil, pool.Pointers[0x21000100], pool.Pointers[0x21000200], 27, 0)
			// Existing independently qualified raw-parent nil-target control is1;
			// no scalar430 external-reply row is used as a parent-body oracle.
			if result != 1 || rawScalarDamageObserveParent4E24B0 != nil || rawDefaultDamageNativeObserverStack4E0B30 != nil {
				t.Fatal("actual raw parent nil-target binding", result)
			}
			pool.compare(t)
			t.Logf("RAW_SCALAR_DAMAGE_DEFAULT_C_NIL_OK function=%s result=1 parent-seam=nil parent-observers=nil opaque-canaries=true", symbol)
		})
	}
}
