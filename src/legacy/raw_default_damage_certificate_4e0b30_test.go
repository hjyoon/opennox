//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// These tables are original recorder/static certificates, not a native parent
// implementation. The immutable ZIP hash binds the table and all raw rows.
type rddInstruction4E0B30 struct {
	Size            int
	Encoding, Range string
}
type rddCallSpec4E0B30 struct {
	Argc                       int
	Target, Name, Kind, Return string
	Indirect                   bool
}

func (b *rddBundle4E0B30) loadContracts() error {
	b.Instructions = map[string]rddInstruction4E0B30{}
	b.Calls = map[string]rddCallSpec4E0B30{}
	data, e := b.member("runs/baseline1034/actual-original-parent-recorder-proof-v2.json")
	if e != nil {
		return e
	}
	var proof struct {
		Boundaries map[string][]struct {
			PC, Encoding string
			Size         int `json:"bytes"`
		} `json:"original_static_instruction_boundaries"`
	}
	if e = json.Unmarshal(data, &proof); e != nil {
		return e
	}
	for name, rows := range proof.Boundaries {
		for _, row := range rows {
			if _, ok := b.Instructions[row.PC]; ok {
				return fmt.Errorf("duplicate sealed instruction")
			}
			b.Instructions[row.PC] = rddInstruction4E0B30{row.Size, row.Encoding, name}
		}
	}
	if len(b.Instructions) != 650 {
		return fmt.Errorf("four original static instruction sets")
	}
	data, e = b.member("sources/service-boundaries-4e0b30-v1.json")
	if e != nil {
		return e
	}
	var spec struct {
		Schema   string
		Direct   []struct{ Site, Target string } `json:"direct_calls"`
		External []struct {
			Address, Name string
			Argc          int    `json:"argument_words"`
			Return        string `json:"return_kind"`
			Sites         []string
			Cleanup       int `json:"callee_cleanup_bytes"`
		} `json:"external_direct_services"`
		Indirect []struct {
			Site, Target, Name string
			Argc               int    `json:"argument_words"`
			Return             string `json:"return_kind"`
			Cleanup            int    `json:"callee_cleanup_bytes"`
		} `json:"indirect_calls"`
	}
	if e = json.Unmarshal(data, &spec); e != nil {
		return e
	}
	if spec.Schema != "opennox-default-damage-parent-service-boundaries-v1" || len(spec.Direct) != 50 || len(spec.External) != 25 || len(spec.Indirect) != 3 {
		return fmt.Errorf("service certificate dimensions")
	}
	external := map[string]rddCallSpec4E0B30{}
	for _, row := range spec.External {
		if row.Cleanup != 0 {
			return fmt.Errorf("external cdecl cleanup")
		}
		for _, site := range row.Sites {
			if _, ok := external[site]; ok {
				return fmt.Errorf("duplicate external site")
			}
			external[site] = rddCallSpec4E0B30{Argc: row.Argc, Target: row.Address, Name: row.Name, Kind: "external-input-boundary", Return: row.Return}
		}
	}
	originals := map[string]rddCallSpec4E0B30{
		"0x004e1400": {Argc: 2, Target: "0x004e1400", Name: "attack_predicate", Kind: "actual-original"},
		"0x004e1470": {Argc: 1, Target: "0x004e1470", Name: "friendly_predicate", Kind: "actual-original"},
		"0x00419a70": {Argc: 1, Target: "0x00419a70", Name: "signed_float32_converter", Kind: "actual-original"},
	}
	for _, row := range spec.Direct {
		call, ok := external[row.Site]
		if !ok {
			call, ok = originals[row.Target]
		}
		if !ok || call.Target != row.Target {
			return fmt.Errorf("undeclared direct call contract")
		}
		b.Calls[row.Site] = call
	}
	for _, row := range spec.Indirect {
		if row.Cleanup != 0 {
			return fmt.Errorf("indirect cdecl cleanup")
		}
		b.Calls[row.Site] = rddCallSpec4E0B30{Argc: row.Argc, Target: row.Target, Name: row.Name, Kind: "external-input-boundary", Return: row.Return, Indirect: true}
	}
	if len(b.Calls) != 53 || len(external) != 39 {
		return fmt.Errorf("53 call/42 external site contracts")
	}
	return nil
}
func mustRDDNumber4E0B30(text string) uint64 {
	v, e := rddNumber4E0B30(text)
	if e != nil {
		panic(e)
	}
	return v
}
func rddOptionalSameJSON4E0B30(a, b json.RawMessage) bool {
	if len(a) == 0 || string(a) == "null" {
		return len(b) == 0 || string(b) == "null"
	}
	return rddSameJSON4E0B30(a, b)
}
func rddInstructionSource4E0B30(event rddEvent4E0B30) string {
	if event.Kind == "external-x87-trampoline-instruction" {
		return "external-x87-trampoline"
	}
	return "actual-original"
}
func rddInitialPrefix4E0B30(in rddInput4E0B30, caller []byte, cw uint32) ([]rddHost4E0B30, error) {
	var out []rddHost4E0B30
	for _, row := range in.ImageInitializers {
		data, e := hex.DecodeString(row.Hex)
		if e != nil {
			return nil, e
		}
		out = append(out, rddHost4E0B30{Reason: "input-image-initializer", Address: uint64(row.Address), Bytes: data})
	}
	out = append(out, rddHost4E0B30{Reason: "input-caller-words", Address: rddCaller4E0B30, Bytes: caller}, rddHost4E0B30{Reason: "framework-fninit-fldcw", Address: 0x23000200, Bytes: []byte{0xdb, 0xe3, 0xd9, 0x2d, 0, 3, 0, 0x23}}, rddHost4E0B30{Reason: "framework-control-word", Address: 0x23000300, Bytes: ddiBytes4E0B30(uint64(cw), 2)})
	return out, nil
}
func rddValidateFault4E0B30(event rddEvent4E0B30, current, issued *rddEvent4E0B30) error {
	pc, e := rddNumber4E0B30(event.PC)
	if e != nil {
		return e
	}
	address, e := rddNumber4E0B30(event.Address)
	if e != nil {
		return e
	}
	if event.Width <= 0 || address > 0xffffffff || pc != uint64(event.Registers["eip"]) || current == nil || event.FPU.Control != 0x0e3f && event.FPU.Control != 0x023f {
		return fmt.Errorf("fault register/address/width certificate")
	}
	switch event.Access {
	case 21, 24:
		if issued == nil || issued.Target == "" || pc != mustRDDNumber4E0B30(issued.Target) || !current.Completed {
			return fmt.Errorf("failed fetch must retain completed CALL and pending callee")
		}
	case 19, 20, 22, 23:
		if event.PC != current.PC || current.Completed {
			return fmt.Errorf("operand fault instruction remains incomplete")
		}
	default:
		return fmt.Errorf("unknown original memory fault access")
	}
	return nil
}
func rddTerminalFrames4E0B30(row *rddRow4E0B30, frames []rddFrame4E0B30, pending *rddFrame4E0B30, issued *rddEvent4E0B30, issuedReturn uint32) error {
	if len(row.Open) != len(frames) {
		return fmt.Errorf("complete open frame count")
	}
	for i, raw := range row.Open {
		var actual rddFrame4E0B30
		if e := json.Unmarshal(raw, &actual); e != nil {
			return e
		}
		if !reflect.DeepEqual(actual, frames[i]) {
			return fmt.Errorf("complete open frame identity/state")
		}
	}
	if pending == nil {
		if len(row.PendingReturn) > 0 && string(row.PendingReturn) != "null" {
			return fmt.Errorf("unexpected pending return")
		}
	} else {
		var actual rddFrame4E0B30
		if e := json.Unmarshal(row.PendingReturn, &actual); e != nil {
			return e
		}
		if !reflect.DeepEqual(actual, *pending) {
			return fmt.Errorf("pending return frame certificate")
		}
	}
	if issued == nil {
		if len(row.PendingCall) > 0 && string(row.PendingCall) != "null" {
			return fmt.Errorf("unexpected pending call")
		}
	} else {
		var actual struct {
			Site, ReturnPC, StackBefore uint32
			Target                      *uint32
		}
		var raw struct {
			Site        uint32  `json:"site"`
			ReturnPC    uint32  `json:"return_pc"`
			StackBefore uint32  `json:"stack_before"`
			Target      *uint32 `json:"target"`
		}
		if e := json.Unmarshal(row.PendingCall, &raw); e != nil {
			return e
		}
		actual.Site = raw.Site
		actual.ReturnPC = raw.ReturnPC
		actual.StackBefore = raw.StackBefore
		actual.Target = raw.Target
		if actual.Site != uint32(mustRDDNumber4E0B30(issued.Site)) || actual.ReturnPC != issuedReturn || actual.StackBefore != uint32(mustRDDNumber4E0B30(issued.StackBefore)) {
			return fmt.Errorf("pending CALL stack/site/return certificate")
		}
		if issued.Target == "" {
			if actual.Target != nil {
				return fmt.Errorf("unreadable indirect operand target")
			}
		} else if actual.Target == nil || *actual.Target != uint32(mustRDDNumber4E0B30(issued.Target)) {
			return fmt.Errorf("pending CALL full target bits")
		}
	}
	return nil
}

// This checks the exact original Unicorn observation. It never writes a native
// graph, predicts ARM fault atomicity, or substitutes a platform golden.
func rddValidateFaultWrite4E0B30(write, event rddEvent4E0B30, fragments map[int]rddEvent4E0B30, mem *rddMemory4E0B30) error {
	var observation struct {
		Operand struct {
			Base, Index                *string
			Segment                    *string
			Scale, Displacement, Width int
		} `json:"decoded_memory_operand"`
		Spans []struct {
			Area, Address              string
			Width, Offset, Permissions int
			Before, After              string
			Changed                    []int `json:"changed_byte_offsets"`
		} `json:"-"`
		Mapped []struct {
			Area, Address string
			Width         int
			Offset        int `json:"write_offset"`
			Permissions   int
			Before        string `json:"before_hex"`
			After         string `json:"after_hex"`
			Changed       []int  `json:"changed_byte_offsets"`
		} `json:"mapped_spans_after"`
		Changed   []int `json:"changed_fragment_sequences"`
		Completed bool  `json:"original_instruction_completed"`
		Physical  bool  `json:"emulator_physical_observation_only"`
	}
	if e := json.Unmarshal(write.Observation, &observation); e != nil {
		return e
	}
	if observation.Completed || !observation.Physical || observation.Operand.Width != write.Width || observation.Operand.Segment != nil || !reflect.DeepEqual(observation.Changed, event.ChangedSequences) || write.Width < 1 || write.Width > 8 {
		return fmt.Errorf("incomplete original store observation extent")
	}
	base, index := uint32(0), uint32(0)
	if observation.Operand.Base != nil {
		var ok bool
		base, ok = write.RegistersAtWrite[*observation.Operand.Base]
		if !ok {
			return fmt.Errorf("store operand base register")
		}
	}
	if observation.Operand.Index != nil {
		var ok bool
		index, ok = write.RegistersAtWrite[*observation.Operand.Index]
		if !ok {
			return fmt.Errorf("store operand index register")
		}
	}
	start := mustRDDNumber4E0B30(write.Address)
	effective := uint32(int64(base) + int64(index)*int64(observation.Operand.Scale) + int64(observation.Operand.Displacement))
	if uint64(effective) != start {
		return fmt.Errorf("decoded original effective store address")
	}
	if len(observation.Mapped) != len(write.MappedBefore) {
		return fmt.Errorf("mapped before/after span set")
	}
	expectedStarts := []uint64{}
	for _, a := range mem.Areas {
		left := max(start, a.Address)
		right := min(start+uint64(write.Width), a.Address+uint64(len(a.Bytes)))
		if left < right {
			expectedStarts = append(expectedStarts, left)
		}
	}
	sort.Slice(expectedStarts, func(i, j int) bool { return expectedStarts[i] < expectedStarts[j] })
	if len(expectedStarts) != len(observation.Mapped) {
		return fmt.Errorf("whole mapped intersections")
	}
	usedFragments := map[int]bool{}
	previousSequence := -1
	for i, span := range observation.Mapped {
		before := write.MappedBefore[i]
		address := mustRDDNumber4E0B30(span.Address)
		if address != expectedStarts[i] || span.Address != before.Address || span.Width != before.Width || span.Offset != before.Offset || span.Before != before.Before || span.Offset < 0 || span.Offset+span.Width > write.Width || address != start+uint64(span.Offset) {
			return fmt.Errorf("mapped fault span identity")
		}
		area := mem.area(address, span.Width)
		if area == nil || area.Ref.Name != span.Area || area.Ref.Permissions != span.Permissions {
			return fmt.Errorf("mapped span provenance")
		}
		old, e := hex.DecodeString(span.Before)
		if e != nil {
			return e
		}
		after, e := hex.DecodeString(span.After)
		if e != nil {
			return e
		}
		actual, e := mem.read(address, span.Width)
		if e != nil || len(old) != span.Width || len(after) != span.Width || !bytes.Equal(actual, after) {
			return fmt.Errorf("physical mapped post-fault bytes")
		}
		changed := []int{}
		for n := range old {
			if old[n] != after[n] {
				changed = append(changed, n)
			}
		}
		if !reflect.DeepEqual(changed, span.Changed) {
			return fmt.Errorf("exact changed byte offsets")
		}
		covered := map[int]bool{}
		for _, seq := range event.ChangedSequences {
			fragment, ok := fragments[seq]
			if !ok {
				return fmt.Errorf("unknown fragment")
			}
			at := int(mustRDDNumber4E0B30(fragment.Address) - address)
			if at < 0 || at+fragment.Width > span.Width {
				continue
			}
			if seq <= previousSequence || usedFragments[seq] {
				return fmt.Errorf("fragment order/duplication")
			}
			previousSequence = seq
			usedFragments[seq] = true
			fBefore, e := hex.DecodeString(fragment.Before)
			if e != nil {
				return e
			}
			fAfter, e := hex.DecodeString(fragment.After)
			if e != nil {
				return e
			}
			if !bytes.Equal(fBefore, old[at:at+fragment.Width]) || !bytes.Equal(fAfter, after[at:at+fragment.Width]) {
				return fmt.Errorf("complete changed-run bytes")
			}
			for n := at; n < at+fragment.Width; n++ {
				if covered[n] || old[n] == after[n] {
					return fmt.Errorf("unchanged/overlapping byte cannot be credited")
				}
				covered[n] = true
			}
		}
		if len(covered) != len(changed) {
			return fmt.Errorf("all physical byte changes must have exact fragment certificates")
		}
		for _, n := range changed {
			if !covered[n] {
				return fmt.Errorf("uncovered changed byte")
			}
		}
	}
	if len(usedFragments) != len(event.ChangedSequences) {
		return fmt.Errorf("fragment outside mapped store domain")
	}
	return nil
}
