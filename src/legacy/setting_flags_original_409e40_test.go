//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

type flagsWindow409E40 struct {
	Base   unsafe.Pointer
	Center int
	Before []byte
}
type flagsActualOperand409E40 struct {
	Site       uint32 `json:"site"`
	Field      int    `json:"field"`
	Address    uint64 `json:"native_address"`
	Width      uint32 `json:"width"`
	Value      uint32 `json:"value"`
	Write      bool   `json:"write"`
	Complete   bool   `json:"complete"`
	ValueValid bool   `json:"value_valid"`
}
type flagsActualRecord409E40 struct {
	Name                 string                     `json:"name"`
	Entry                uint32                     `json:"entry"`
	Argument             uint32                     `json:"argument_u32"`
	Backend              string                     `json:"backend"`
	InputSHA             string                     `json:"input_sha256"`
	NativeCallStarted    bool                       `json:"native_call_started"`
	NativeReturned       bool                       `json:"native_returned"`
	Return               uint32                     `json:"return_u32"`
	Before               [3]uint32                  `json:"before_words"`
	After                [3]uint32                  `json:"after_words"`
	Addresses            [3]uint64                  `json:"actual_cell_addresses"`
	Operands             []flagsActualOperand409E40 `json:"actual_operands"`
	Steps                []flagsStep409E40          `json:"actual_logical_steps"`
	Used                 map[string][]int           `json:"used_mutations"`
	Unused               map[string][]int           `json:"unused_mutations"`
	UnusedServices       []uint32                   `json:"unused_service_sites"`
	WindowBases          [3]uint64                  `json:"actual_window_bases"`
	WindowCenters        [3]int                     `json:"actual_window_centers"`
	WindowBefore         [3][]byte                  `json:"native_window_BEFORE_base64"`
	WindowAfter          [3][]byte                  `json:"native_window_AFTER_base64"`
	OriginalCWs          [2]uint16                  `json:"original_CW_references"`
	ReferenceComparisons int                        `json:"original_reference_comparisons_passed"`
	Qualified            bool                       `json:"qualified"`
	Error                string                     `json:"error,omitempty"`
}
type flagsNativeContext409E40 struct {
	Input       *flagsInput409E40
	Cells       [3]unsafe.Pointer
	Record      *flagsActualRecord409E40
	Pending     *settingFlagsNativeOperand409E40
	Used        map[string]map[int]bool
	Occurrence  map[string]int
	ServiceSeen map[uint32]bool
}

func flagsReadWords409E40(cells [3]unsafe.Pointer) (out [3]uint32) {
	for i, p := range cells {
		out[i] = *(*uint32)(p)
	}
	return
}
func flagsSlice409E40(v []int) []int {
	if v == nil {
		return []int{}
	}
	return v
}
func flagsSnapshotWindows409E40(windows [3]flagsWindow409E40) (out [3][]byte) {
	for i, w := range windows {
		out[i] = settingFlagsFixtureSnapshot409E40(w.Base, len(w.Before))
	}
	return
}
func flagsCellField409E40(address uint32) int {
	field := flagsPEField409E40(address)
	if field < 0 {
		panic("independent write outside selected three DWORDs")
	}
	return field
}
func (c *flagsNativeContext409E40) apply(writes []flagsInputWrite409E40) {
	for _, write := range writes {
		data, e := flagsHex409E40(write.Hex)
		if e != nil || len(data) != 4 {
			panic("independent whole DWORD write")
		}
		settingFlagsFixtureCopy409E40(c.Cells[flagsCellField409E40(write.Address)], data)
	}
}
func (c *flagsNativeContext409E40) mutate(kind string, site, address, width uint32) {
	key := fmt.Sprintf("%s/%08x/%08x/%d", kind, site, address, width)
	c.Occurrence[key]++
	var selected []flagsInputStimulus409E40
	switch kind {
	case "after_reads":
		selected = c.Input.AfterReads
	case "after_writes":
		selected = c.Input.AfterWrites
	case "after_services":
		selected = c.Input.AfterServices
	default:
		panic("unknown mutation kind")
	}
	for i, stimulus := range selected {
		matches := false
		if kind == "after_services" {
			matches = stimulus.Site == fmt.Sprintf("%08x", site)
		} else {
			matches = stimulus.PC == site && stimulus.Address == address && stimulus.Width == width
		}
		if matches && stimulus.Occurrence == c.Occurrence[key] {
			if c.Used[kind][i] {
				panic("mutation consumed twice")
			}
			c.apply(stimulus.Writes)
			c.Used[kind][i] = true
		}
	}
}
func (c *flagsNativeContext409E40) operand(event settingFlagsNativeOperand409E40) {
	if event.Field < 0 || event.Field > 1 || event.Address != c.Cells[event.Field] || event.Width != 4 {
		panic("actual operand C global/width identity")
	}
	c.Record.Operands = append(c.Record.Operands, flagsActualOperand409E40{Site: event.Site, Field: event.Field, Address: uint64(uintptr(event.Address)), Width: event.Width, Value: event.Value, Write: event.Write, Complete: event.Complete, ValueValid: event.ValueValid})
	if !event.Complete {
		if c.Pending != nil {
			panic("duplicate before operand")
		}
		if event.Write && !event.ValueValid || !event.Write && (event.ValueValid || event.Value != 0) {
			panic("before-read is not a replacement loaded value")
		}
		copy := event
		c.Pending = &copy
		return
	}
	if c.Pending == nil || c.Pending.Site != event.Site || c.Pending.Field != event.Field || c.Pending.Address != event.Address || c.Pending.Write != event.Write || !event.ValueValid {
		panic("actual before/complete operand pairing")
	}
	if event.Write && c.Pending.Value != event.Value {
		panic("actual store value changed")
	}
	c.Pending = nil
	kind, stimulus := "read", "after_reads"
	if event.Write {
		kind, stimulus = "write", "after_writes"
		if *(*uint32)(event.Address) != event.Value {
			panic("completed native DWORD store bytes")
		}
	}
	c.Record.Steps = append(c.Record.Steps, flagsStep409E40{Kind: kind, Site: event.Site, Field: event.Field, Word: event.Value})
	c.mutate(stimulus, event.Site, flagsPECells409E40[event.Field], 4)
}
func (c *flagsNativeContext409E40) service(site, mask uint32) (uint32, bool) {
	if c.Pending != nil || c.ServiceSeen[site] {
		panic("service inside incomplete operand or duplicate call")
	}
	packet, ok := c.Input.Services[fmt.Sprintf("%08x", site)]
	if !ok {
		panic("undeclared external service site")
	}
	switch site {
	case 0x409e90, 0x409f0a:
		if packet.Argc != 1 || mask != 1 {
			panic("actual Check(1) word ABI")
		}
	case 0x409f52:
		if packet.Argc != 1 || mask != 1056 {
			panic("actual Check(1056) word ABI")
		}
	case 0x409ea4, 0x409f28:
		if packet.Argc != 0 || mask != 0 {
			panic("actual Refresh zero-word ABI")
		}
	default:
		panic("unknown five-flags boundary")
	}
	c.Record.Steps = append(c.Record.Steps, flagsStep409E40{Kind: "service", Site: site, Mask: mask})
	c.ServiceSeen[site] = true
	c.apply(packet.Writes)
	c.mutate("after_services", site, 0, 0)
	return packet.Registers["eax"], true
}
func (c *flagsNativeContext409E40) finishConsumption() {
	c.Record.Used = make(map[string][]int)
	c.Record.Unused = make(map[string][]int)
	for kind, n := range map[string]int{"after_reads": len(c.Input.AfterReads), "after_writes": len(c.Input.AfterWrites), "after_services": len(c.Input.AfterServices)} {
		for i := 0; i < n; i++ {
			if c.Used[kind][i] {
				c.Record.Used[kind] = append(c.Record.Used[kind], i)
			} else {
				c.Record.Unused[kind] = append(c.Record.Unused[kind], i)
			}
		}
	}
	for _, site := range []uint32{0x409e90, 0x409ea4, 0x409f0a, 0x409f28, 0x409f52} {
		if !c.ServiceSeen[site] {
			c.Record.UnusedServices = append(c.Record.UnusedServices, site)
		}
	}
}
func flagsCompareNative409E40(record *flagsActualRecord409E40, reference *flagsReference409E40) error {
	if record.Return != reference.Result || record.After != reference.After {
		return fmt.Errorf("authentic original raw32 EAX/three state DWORDs CW=%04x", reference.CW)
	}
	if len(record.Steps) != len(reference.Steps) {
		return fmt.Errorf("ordered native/original scalar/service step count")
	}
	for i, actual := range record.Steps {
		wanted := reference.Steps[i]
		if actual.Kind != wanted.Kind || actual.Site != wanted.Site || actual.Field != wanted.Field || actual.Word != wanted.Word || actual.Mask != wanted.Mask {
			return fmt.Errorf("authentic original ordered operand/service %d CW=%04x", i, reference.CW)
		}
	}
	for _, kind := range []string{"after_reads", "after_writes", "after_services"} {
		if !reflect.DeepEqual(flagsSlice409E40(record.Used[kind]), flagsSlice409E40(reference.Used[kind])) || !reflect.DeepEqual(flagsSlice409E40(record.Unused[kind]), flagsSlice409E40(reference.Unused[kind])) {
			return fmt.Errorf("actual used/unused independent mutation identity")
		}
	}
	if !reflect.DeepEqual(record.UnusedServices, reference.UnusedServices) {
		return fmt.Errorf("actual unused external service inputs")
	}
	for i, before := range record.WindowBefore {
		expected := append([]byte(nil), before...)
		base := record.WindowBases[i]
		for field, address := range record.Addresses {
			if address < base || address+4 > base+uint64(len(expected)) {
				continue
			}
			binary.LittleEndian.PutUint32(expected[address-base:], reference.After[field])
		}
		if !bytes.Equal(expected, record.WindowAfter[i]) {
			return fmt.Errorf("actual C neighbour/private canary full-window bytes CW=%04x", reference.CW)
		}
	}
	return nil
}
func flagsCallGo409E40(entry uint32, word int32) int32 {
	switch entry {
	case 0x409e40:
		return sub409E40Native(word)
	case 0x409e70:
		return sub409E70Native(word)
	case 0x409ec0:
		return sub409EC0Native(word)
	case 0x409ef0:
		return sub409EF0Native(word)
	case 0x409f40:
		return sub409F40Native(word)
	}
	panic("unknown selected five-flags entry")
}

func flagsRunOne409E40(chosen *flagsChosen409E40, backend string, gameAllocation, game unsafe.Pointer, encoder *json.Encoder) (err error) {
	cells := [3]unsafe.Pointer{settingFlagsNativeCell409E40(0), settingFlagsNativeCell409E40(1), game}
	record := flagsActualRecord409E40{Name: chosen.Input.Name, Entry: chosen.Input.Entry, Argument: chosen.Input.Args[0], Backend: backend, InputSHA: flagsHash409E40(chosen.Raw), OriginalCWs: [2]uint16{0x0e3f, 0x023f}}
	context := &flagsNativeContext409E40{Input: &chosen.Input, Cells: cells, Record: &record, Used: map[string]map[int]bool{"after_reads": {}, "after_writes": {}, "after_services": {}}, Occurrence: make(map[string]int), ServiceSeen: make(map[uint32]bool)}
	var windows [3]flagsWindow409E40
	var original [3][]byte
	var pop func()
	prepared := false
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("actual five-flags call/observer panic: %v", recovered)
		}
		if pop != nil {
			pop()
		}
		if prepared {
			record.After = flagsReadWords409E40(cells)
			record.WindowAfter = flagsSnapshotWindows409E40(windows)
			context.finishConsumption()
			if err == nil && context.Pending != nil {
				err = fmt.Errorf("native returned with incomplete operand")
			}
			if err == nil && !record.NativeReturned {
				err = fmt.Errorf("actual native return absent")
			}
			if err == nil {
				for _, reference := range chosen.References {
					if e := flagsCompareNative409E40(&record, reference); e != nil {
						err = e
						break
					}
					record.ReferenceComparisons++
				}
			}
			// Capture the actual failure before restoration. Full known readable
			// windows restore only this test process's original C bytes, never files.
			for i, w := range windows {
				settingFlagsFixtureCopy409E40(w.Base, original[i])
			}
			for i, w := range windows {
				if !bytes.Equal(settingFlagsFixtureSnapshot409E40(w.Base, len(original[i])), original[i]) {
					err = fmt.Errorf("actual C state/canary restoration failed")
				}
			}
		}
		record.Qualified = err == nil && record.ReferenceComparisons == 2
		if err != nil {
			record.Error = err.Error()
		}
		if e := encoder.Encode(record); e != nil && err == nil {
			err = e
		}
		runtime.KeepAlive(gameAllocation)
	}()
	for i, p := range cells {
		if uintptr(p) <= 0xffffffff || uintptr(p)%4 != 0 {
			return fmt.Errorf("actual aligned C cell exceeds 4GiB")
		}
		record.Addresses[i] = uint64(uintptr(p))
		for j := 0; j < i; j++ {
			if cells[j] == p {
				return fmt.Errorf("three independent native cells")
			}
		}
	}
	for i, p := range cells {
		if i == 2 {
			windows[i] = flagsWindow409E40{Base: gameAllocation, Center: 16, Before: settingFlagsFixtureSnapshot409E40(gameAllocation, 36)}
		} else {
			base, n, center, e := settingFlagsFixtureWindow409E40(p)
			if e != nil {
				return e
			}
			windows[i] = flagsWindow409E40{Base: base, Center: center, Before: settingFlagsFixtureSnapshot409E40(base, n)}
		}
		original[i] = append([]byte(nil), windows[i].Before...)
	}
	prepared = true
	for i, p := range cells {
		word, ok := chosen.Input.Words[fmt.Sprintf("%08x", flagsPECells409E40[i])]
		if !ok {
			return fmt.Errorf("chosen initial DWORD absent")
		}
		*(*uint32)(p) = word
	}
	record.Before = flagsReadWords409E40(cells)
	for i, w := range windows {
		record.WindowBases[i] = uint64(uintptr(w.Base))
		record.WindowCenters[i] = w.Center
		record.WindowBefore[i] = settingFlagsFixtureSnapshot409E40(w.Base, len(w.Before))
	}
	pop = pushSettingFlagsNativeObserver409E40(&settingFlagsNativeObserver409E40{Operand: context.operand, Service: context.service})
	record.NativeCallStarted = true
	var result int32
	if backend == "Go" {
		result = flagsCallGo409E40(chosen.Input.Entry, int32(chosen.Input.Args[0]))
	} else if backend == "C" {
		result = settingFlagsFixtureCall409E40(chosen.Input.Entry, int32(chosen.Input.Args[0]))
	} else {
		return fmt.Errorf("selected native backend")
	}
	record.Return = uint32(result)
	record.NativeReturned = true
	return nil
}

func TestSettingFlagsNativeLayout409E40(t *testing.T) {
	for field := 0; field < 2; field++ {
		p := settingFlagsNativeCell409E40(field)
		if uintptr(p) <= 0xffffffff || settingFlagsNativeCellWidth409E40(field) != 4 || settingFlagsNativeCellAlign409E40(field) != 4 || uintptr(p)%4 != 0 {
			t.Fatalf("actual native C symbol DWORD layout field=%d", field)
		}
		base, n, center, e := settingFlagsFixtureWindow409E40(p)
		if e != nil || base == nil || n < 4 || n > 36 || center < 0 || center+4 > n || uintptr(base)+uintptr(center) != uintptr(p) {
			t.Fatalf("actual VM-bounded C window field=%d: %v", field, e)
		}
	}
	if settingFlagsNativeCell409E40(0) == settingFlagsNativeCell409E40(1) {
		t.Fatal("actual symbols must be independent")
	}
	t.Log("FIVE_FLAGS_ACTUAL_C_GLOBAL_LAYOUT_OK symbols=2 sizeof=4 align=4 native>4GiB typed-C-int-ABI=5 contiguous-layout-claim=false")
}

func TestSettingFlagsNativeOriginal409E40(t *testing.T) {
	chosen, e := flagsLoadChosen409E40()
	if e != nil {
		t.Fatal(e)
	}
	root := os.Getenv("OPENNOX_LOCAL_TEST_ARTIFACTS")
	if root == "" {
		root = os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	}
	if root == "" || !filepath.IsAbs(root) {
		t.Fatal("absolute unique retained local test artifacts directory required")
	}
	if e = os.MkdirAll(root, 0700); e != nil {
		t.Fatal(e)
	}
	directory, e := os.MkdirTemp(root, "setting-flags-native-409e40-")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.OpenFile(filepath.Join(directory, "actual-native-1261-Go-C-calls.jsonl.gz"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	compressed := gzip.NewWriter(raw)
	encoder := json.NewEncoder(compressed)
	defer func() {
		if e := compressed.Close(); e != nil {
			t.Error(e)
		}
		if e := raw.Close(); e != nil {
			t.Error(e)
		}
	}()
	allocation, game := settingFlagsFixtureGameNew409E40()
	defer settingFlagsFixtureFree409E40(allocation)
	attempted, passed := 0, 0
	for i := range chosen {
		for _, backend := range []string{"Go", "C"} {
			attempted++
			item := &chosen[i]
			if t.Run(backend+"/"+item.Input.Name, func(t *testing.T) {
				if e := flagsRunOne409E40(item, backend, allocation, game, encoder); e != nil {
					t.Fatal(e)
				}
				t.Logf("FIVE_FLAGS_NATIVE_ORIGINAL_OK backend=%s input=%s input-sha256=%s original-CW-comparisons=2 external-body-credit=0", backend, item.Input.Name, flagsHash409E40(item.Raw))
			}) {
				passed++
			}
		}
	}
	summary := map[string]any{"selected_inputs": 1261, "scheduled_backend_cases": attempted, "qualified_native_calls": passed, "separate_original_CW_comparisons_passed": passed * 2, "input_sha256": flagsInputSHA409E40, "original_reference_sha256": flagsReferenceSHA409E40, "actual_C_global_symbols": 2, "private_Game_DWORD_is_product_GoFlags_shadow": false, "actual_native_addresses": [3]uintptr{uintptr(settingFlagsNativeCell409E40(0)), uintptr(settingFlagsNativeCell409E40(1)), uintptr(game)}, "original_GPR_EFLAGS_caller_stack_physical_ABI_comparison": false, "external_Check_Refresh_body_execution_credit": false, "declared_ECX_EDX_physical_comparison": false, "Go_heap_pointer_domain_credit": false, "natural_product_campaign_credit": false, "new_original_execution": 0, "raw_actual_trace_file": "actual-native-1261-Go-C-calls.jsonl.gz"}
	data, e := json.MarshalIndent(summary, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(directory, "actual-five-flags-native-summary.json"), append(data, '\n'), 0600); e != nil {
		t.Fatal(e)
	}
	if attempted != 2522 || passed != 2522 {
		t.Fatalf("five-flags native calls scheduled=%d passed=%d", attempted, passed)
	}
	t.Logf("FIVE_FLAGS_NATIVE_ALL_OK actual-calls=%d original-CW-comparisons=%d rawtrace=%s", passed, passed*2, directory)
}

// No parallel global mutation, callbacks substituting memory loads, bool
// return narrowing or generated return model.
