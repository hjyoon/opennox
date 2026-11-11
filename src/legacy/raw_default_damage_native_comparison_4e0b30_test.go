//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"unsafe"
)

type rddNativeStep4E0B30 struct {
	DerivedAttempt    bool
	Kind              string
	Field             int
	Base              uint32
	Event             rddEvent4E0B30
	Frame             rddFrame4E0B30
	Effect            *rddEffect4E0B30
	EffectArgs        []uint32
	ArgumentKinds     []string
	Strings           map[int]string
	Coordinates       map[int][4]uint32
	DamageBefore      map[int]uint32
	TerminalFragments []rddEvent4E0B30
}
type rddActualChildFrame4E0B30 struct {
	Entry       uint32
	WeaponClass uint32
	ClassKnown  bool
}
type rddNativeComparison4E0B30 struct {
	ActualCalls            []rddActualChildFrame4E0B30
	DerivedReadAttempts    int
	ExcludedImplicitWrites int
	CallbackPending        bool
	CallbackSite           uint32
	CallbackTarget         unsafe.Pointer
	Record                 func(rddNativeObserved4E0B30) error
	ObservedCount          int
	ValidationFailure      func(position int, message string)
	PendingCaller          *rawDefaultDamageNativeCaller4E0B30
	CallerSink             func(position int, event rawDefaultDamageNativeCaller4E0B30)
	CallerNativeAddress    uintptr
	Bundle                 *rddBundle4E0B30
	Row                    *rddRow4E0B30
	Input                  rddInput4E0B30
	Projection             *rddProjection4E0B30
	Steps                  []rddNativeStep4E0B30
	Position               int
	Pending                *rawDefaultDamageNativeOperand4E0B30
	Damage                 *int32
	DispatchedRoot         bool
	Backend                string
	Progress               func(position int, operand rawDefaultDamageNativeOperand4E0B30)
}

func rddNativeFieldPCs4E0B30(b *rddBundle4E0B30, f *ddiFixture4E0B30) (map[uint32]int, error) {
	data, e := b.member("provenance/independent-contract-review-v1/static-stack-field-width-certificate-v1.json")
	if e != nil {
		return nil, e
	}
	var table struct {
		Rows []struct {
			PC           string
			Displacement int
		} `json:"parent_role_base_memory_operands"`
	}
	if e = json.Unmarshal(data, &table); e != nil {
		return nil, e
	}
	offsets := map[int]int{}
	for _, field := range f.Fields {
		if _, exists := offsets[field.Offset]; exists {
			return nil, fmt.Errorf("ambiguous static PE field offset")
		}
		offsets[field.Offset] = field.ID
	}
	out := map[uint32]int{}
	for _, row := range table.Rows {
		pc := uint32(mustRDDNumber4E0B30(row.PC))
		if row.Displacement == 0x84ea04 {
			out[pc] = -1
			continue
		}
		id, ok := offsets[row.Displacement]
		if !ok {
			return nil, fmt.Errorf("unmapped original role operand %#x/%d", pc, row.Displacement)
		}
		out[pc] = id
	}
	for pc, id := range map[uint32]int{0x4e140c: 1, 0x4e141d: 2, 0x4e142c: 1, 0x4e1434: 2, 0x4e143d: 17, 0x4e1443: 25, 0x4e1456: 2, 0x4e1478: 1, 0x4e1481: 2} {
		out[pc] = id
	}

	// Additional exact original operands absent from the object-only65 list.
	// Bound by static-complete-field-PC-addendum-v1.json; constants are logical f32 inputs.
	out[0x004e0b8f] = 24 // mov dword ptr [eax + 0x88c], 0
	out[0x004e0baf] = 21 // or dword ptr [eax + 0x5a0], 0x80000
	out[0x004e0c24] = 1  // test byte ptr [ebp + 8], 6
	out[0x004e0da6] = -4 // fcomp dword ptr [0x58307c]
	out[0x004e0dcd] = -4 // fld dword ptr [0x583068]
	out[0x004e0e0f] = -4 // fcomp dword ptr [0x58307c]
	out[0x004e0e36] = -4 // fld dword ptr [0x583068]
	out[0x004e0e6d] = 20 // mov word ptr [edx + 0xa0], 2
	out[0x004e0e88] = 22 // mov byte ptr [eax + 0x82e], 2
	out[0x004e0ee2] = 24 // mov dword ptr [eax + 0x88c], 1
	out[0x004e0f1c] = 24 // mov dword ptr [eax + 0x88c], 1
	out[0x004e0f2a] = 23 // mov dword ptr [eax + 0x888], edx
	out[0x004e0fb0] = 21 // mov edx, dword ptr [eax + 0x5a0]
	out[0x004e0fb6] = 24 // mov ecx, dword ptr [eax + 0x88c]
	out[0x004e0fc1] = 21 // mov dword ptr [eax + 0x5a0], edx
	out[0x004e0fc9] = 24 // mov dword ptr [eax + 0x88c], 2
	out[0x004e0fd3] = 23 // mov dword ptr [eax + 0x888], ebp
	out[0x004e1025] = 26 // mov eax, dword ptr [eax + 0x20]
	out[0x004e114d] = 19 // mov al, byte ptr [eax + 0x58]
	out[0x004e119c] = 1  // test byte ptr [eax + 8], 2
	if len(out) != 94 {
		return nil, fmt.Errorf("complete94 nonstack field/literal PC correspondence")
	}
	return out, nil
}
func rddNativePlan4E0B30(b *rddBundle4E0B30, row *rddRow4E0B30, f *ddiFixture4E0B30, g *ddiGraph4E0B30) (*rddNativeComparison4E0B30, error) {
	if _, e := rddValidateRow4E0B30(b, row); e != nil {
		return nil, e
	}
	var in rddInput4E0B30
	if e := json.Unmarshal(row.Input, &in); e != nil {
		return nil, e
	}
	projection, e := rddNewProjection4E0B30(g, in)
	if e != nil {
		return nil, e
	}
	pcFields, e := rddNativeFieldPCs4E0B30(b, f)
	if e != nil {
		return nil, e
	}
	out := &rddNativeComparison4E0B30{Bundle: b, Row: row, Input: in, Projection: projection}
	out.Steps = append(out.Steps, rddNativeStep4E0B30{Kind: "caller-entry", Event: rddEvent4E0B30{PC: "0x004e0b30", Completed: true, After: hex.EncodeToString(ddiBytes4E0B30(uint64(in.Args[3]), 4))}})
	mem := &rddMemory4E0B30{}
	for _, ref := range row.MemoryBefore {
		data, e := b.snapshot(ref.Snapshot)
		if e != nil {
			return nil, e
		}
		mem.Areas = append(mem.Areas, rddArea4E0B30{Ref: ref, Address: mustRDDNumber4E0B30(ref.Address), Bytes: data})
	}
	var service rddEvent4E0B30
	var returned rddFrame4E0B30
	effectOffsets := map[string]int{}
	for index, raw := range row.Events {
		if index < row.Prefix {
			continue
		}
		var event rddEvent4E0B30
		if e = json.Unmarshal(raw, &event); e != nil {
			return nil, e
		}
		switch event.Kind {
		case "read", "write":
			if event.Kind == "write" && mustRDDNumber4E0B30(event.Address) == rddCaller4E0B30+16 {
				if event.Width != 4 {
					return nil, fmt.Errorf("unexpected original caller damage store width")
				}
				out.Steps = append(out.Steps, rddNativeStep4E0B30{Kind: "caller-write", Event: event})
			}
			field, ok := pcFields[uint32(mustRDDNumber4E0B30(event.PC))]
			if ok && event.Source == "actual-original" {
				definition := rddNativeStaticOperands4E0B30[uint32(mustRDDNumber4E0B30(event.PC))]
				required := 1
				if event.Kind == "write" {
					required = 2
				}
				if definition.Access&required == 0 {
					if event.Kind == "write" {
						out.ExcludedImplicitWrites++
					}
					ok = false
				} else if event.Width != definition.Width {
					return nil, fmt.Errorf("explicit original operand width mismatch")
				}
			}
			if ok && event.Source == "actual-original" {
				base := uint32(0)
				if field >= 0 {
					address := mustRDDNumber4E0B30(event.Address)
					offset := uint64(f.Fields[field].Offset)
					if address < offset {
						return nil, fmt.Errorf("canonical operand base underflow")
					}
					base = uint32(address - offset)
					if _, ok = g.Tokens[base]; !ok {
						return nil, fmt.Errorf("native operand lacks exact canonical pointer identity %#x", base)
					}
				}
				out.Steps = append(out.Steps, rddNativeStep4E0B30{Kind: event.Kind, Field: field, Base: base, Event: event})
			}
			if event.Kind == "write" && event.Completed {
				data, e := hex.DecodeString(event.After)
				if e != nil {
					return nil, e
				}
				if e = mem.write(mustRDDNumber4E0B30(event.Address), data); e != nil {
					return nil, e
				}
			}
		case "host-write":
			if event.Reason != "framework-exact-x87-service-payload" && event.Reason != "framework-exact-x87-fld-ret" {
				effect, args, e := rddDeclaredEffectForHost4E0B30(out, event, service, returned, effectOffsets)
				if e != nil {
					return nil, e
				}
				out.Steps = append(out.Steps, rddNativeStep4E0B30{Kind: "effect", Event: event, Effect: &effect, EffectArgs: args})
			}
			data, e := hex.DecodeString(event.After)
			if e != nil {
				return nil, e
			}
			if e = mem.write(mustRDDNumber4E0B30(event.Address), data); e != nil {
				return nil, e
			}
		case "external-service-input":
			service = event
			step := rddNativeStep4E0B30{Kind: "service", Event: event}
			if e = rddPrepareNativeArguments4E0B30(b.Calls[event.Site].Name, event.Args, mem, &step); e != nil {
				return nil, e
			}
			out.Steps = append(out.Steps, step)
		case "call-entered":
			var frame rddFrame4E0B30
			if e = json.Unmarshal(event.Frame, &frame); e != nil {
				return nil, e
			}
			if frame.Kind == "actual-original" {
				out.Steps = append(out.Steps, rddNativeStep4E0B30{Kind: "call-before", Frame: frame})
			}
		case "return-completed":
			var frame rddFrame4E0B30
			if e = json.Unmarshal(event.Frame, &frame); e != nil {
				return nil, e
			}
			if frame.Kind == "external-input-boundary" {
				returned = frame
			} else if frame.Entry != "0x004e0b30" {
				out.Steps = append(out.Steps, rddNativeStep4E0B30{Kind: "call-after", Frame: frame})
			}
		case "observed-partial-write-fragment":
			found := false
			for i := range out.Steps {
				step := &out.Steps[i]
				if step.Kind == "write" && step.Event.Sequence == event.OriginalWrite {
					step.TerminalFragments = append(step.TerminalFragments, event)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("certified terminal fragment lacks original write")
			}
			data, e := hex.DecodeString(event.After)
			if e != nil {
				return nil, e
			}
			if e = mem.write(mustRDDNumber4E0B30(event.Address), data); e != nil {
				return nil, e
			} // ORIGINAL shadow only.
		case "fault":
			if event.Access == 19 || event.Access == 22 {
				hooked := false
				if len(out.Steps) > 0 {
					previous := out.Steps[len(out.Steps)-1]
					hooked = previous.Kind == "read" && !previous.Event.Completed && previous.Event.PC == event.PC
				}
				if !hooked {
					pc := uint32(mustRDDNumber4E0B30(event.PC))
					definition, ok := rddNativeStaticOperands4E0B30[pc]
					if !ok || definition.Field < 0 || definition.Base == "" || definition.Encoding != b.Instructions[event.PC].Encoding {
						return nil, fmt.Errorf("unqualified original no-hook fault operand")
					}
					baseRegister, ok := event.Registers[definition.Base]
					if !ok {
						return nil, fmt.Errorf("original fault base register absent")
					}
					address := uint32(int64(baseRegister) + definition.Displacement)
					faultAddress := mustRDDNumber4E0B30(event.Address)
					if event.Width <= 0 || max(uint64(address), faultAddress) >= min(uint64(address)+uint64(definition.Width), faultAddress+uint64(event.Width)) {
						return nil, fmt.Errorf("static original attempted operand does not overlap observed fault")
					}
					field := definition.Field
					offset := uint32(f.Fields[field].Offset)
					base := address - offset
					if _, ok = g.Tokens[base]; !ok {
						return nil, fmt.Errorf("no-hook attempted operand lacks canonical base identity")
					}
					attempt := rddEvent4E0B30{Kind: "read", PC: event.PC, Address: fmt.Sprintf("0x%08x", address), Width: definition.Width, Source: "qualified-original-fault-operand-no-read-hook", Completed: false, Fault: true}
					out.Steps = append(out.Steps, rddNativeStep4E0B30{Kind: "read", Field: field, Base: base, Event: attempt, DerivedAttempt: true})
					out.DerivedReadAttempts++
				}
			}
			out.Steps = append(out.Steps, rddNativeStep4E0B30{Kind: "fault", Event: event})
		}
	}
	return out, nil
}
func (c *rddNativeComparison4E0B30) next() rddNativeStep4E0B30 {
	if c.Position >= len(c.Steps) {
		panic("native damage produced an extra observable action")
	}
	return c.Steps[c.Position]
}
func (c *rddNativeComparison4E0B30) fail(format string, args ...any) {
	message := fmt.Sprintf("native damage %s/%s step%d: %s", c.Row.Name, c.Row.Mode, c.Position, fmt.Sprintf(format, args...))
	if c.ValidationFailure != nil {
		c.ValidationFailure(c.Position, message)
	}
	panic(fmt.Errorf("%s", message))
}
func (c *rddNativeComparison4E0B30) pointer(word uint32) unsafe.Pointer {
	token, ok := c.Projection.Graph.Tokens[word]
	if !ok {
		c.fail("unknown canonical pointer %#x", word)
	}
	return token.Pointer
}
func (c *rddNativeComparison4E0B30) operand(event rawDefaultDamageNativeOperand4E0B30) {
	c.record(rddNativeObserved4E0B30{Kind: "operand", Site: event.Site, Field: event.Field, Width: event.Width, Complete: event.Complete, Write: event.Write, Base: uintptr(event.Base), Address: uintptr(event.Address), Pointer: uintptr(event.Value.Pointer), Word: event.Value.Word})
	step := c.next()
	kind := "read"
	if event.Write {
		kind = "write"
	}
	if step.Kind != kind || event.Site != uint32(mustRDDNumber4E0B30(step.Event.PC)) || event.Field != step.Field {
		c.fail("ordered operand identity")
	}
	g := c.Projection.Graph
	pointer := step.Field >= 0 && g.Definitions[step.Field].pointer()
	width := step.Event.Width
	if pointer {
		width = int(unsafe.Sizeof(uintptr(0)))
	}
	base := c.pointer(step.Base)
	if step.Field == -1 {
		base = g.base(c.Projection.ClockIndex)
	} else if step.Field == -4 {
		base = nil
	}
	address := base
	if step.Field >= 0 && base != nil {
		address = unsafe.Add(base, g.Offsets[step.Field])
	}
	if event.Base != base || event.Address != address || event.Width != width {
		c.fail("exact native base/field/operand extent")
	}
	if event.Write {
		if pointer {
			if event.Value.Pointer != c.pointer(uint32(step.Event.Value)) {
				c.fail("actual attempted full-width pointer store value")
			}
		} else {
			mask := uint64(0xffffffff)
			if step.Event.Width == 1 {
				mask = 0xff
			} else if step.Event.Width == 2 {
				mask = 0xffff
			}
			if uint64(event.Value.Word) != step.Event.Value&mask {
				c.fail("actual attempted scalar store operand bits")
			}
		}
	}
	if !event.Complete {
		if c.Pending != nil {
			c.fail("nested operand before completion")
		}
		copy := event
		c.Pending = &copy
		if len(step.TerminalFragments) != 0 {
			if e := c.stageTerminalFragments(step); e != nil {
				c.fail("terminal expected-shadow fragment: %v", e)
			}
		}
		if c.Progress != nil {
			c.Progress(c.Position, event)
		}
		return
	}
	if c.Pending == nil || !step.Event.Completed {
		c.fail("completed native operand outside original completed domain")
	}
	var original []byte
	var e error
	if kind == "read" {
		if step.Event.ValueHex == nil {
			c.fail("original completed read lacks bytes")
		}
		original, e = hex.DecodeString(*step.Event.ValueHex)
	} else {
		original, e = hex.DecodeString(step.Event.After)
	}
	if e != nil {
		c.fail("original operand bytes")
	}
	if pointer {
		if event.Value.Pointer != c.pointer(uint32(ddiUnsigned4E0B30(original))) {
			c.fail("full native pointer identity")
		}
	} else if uint64(event.Value.Word) != ddiUnsigned4E0B30(original) {
		c.fail("actual scalar operand bits")
	}
	if kind == "write" {
		if e = c.acceptNativeWrite(step, original); e != nil {
			c.fail("%v", e)
		}
	}
	c.Pending = nil
	c.Position++
	c.drainEffects()
}

// Prepare only the expected bank before the actual attempted store. The
// original row has already passed its independent fragment certificate. This
// never writes an actual page, canonical input or AFTER artifact; a missing or
// wrong physical low byte still fails the complete actual-vs-expected banks.
func (c *rddNativeComparison4E0B30) stageTerminalFragments(step rddNativeStep4E0B30) error {
	if step.Kind != "write" || step.Event.Completed || !step.Event.Fault || step.Event.Width != 2 || len(step.TerminalFragments) != 1 {
		return fmt.Errorf("unsupported terminal store fragment shape")
	}
	g := c.Projection.Graph
	if step.Field < 0 || g.Definitions[step.Field].pointer() {
		return fmt.Errorf("terminal fragment requires scalar field")
	}
	fragment := step.TerminalFragments[0]
	if fragment.OriginalWrite != step.Event.Sequence || fragment.WriteOffset != 0 || fragment.Width != 1 || fragment.OriginalCompleted || !fragment.PhysicalOnly || fragment.PC != step.Event.PC || mustRDDNumber4E0B30(fragment.Address) != mustRDDNumber4E0B30(step.Event.Address) {
		return fmt.Errorf("terminal fragment identity/order")
	}
	before, err := hex.DecodeString(fragment.Before)
	if err != nil || len(before) != 1 {
		return fmt.Errorf("terminal fragment BEFORE bytes")
	}
	after, err := hex.DecodeString(fragment.After)
	if err != nil || len(after) != 1 || after[0] != byte(step.Event.Value) {
		return fmt.Errorf("terminal fragment attempted value")
	}
	token, ok := g.Tokens[step.Base]
	if !ok || token.Kind != 1 || int(token.Allocation) >= len(g.Allocations) {
		return fmt.Errorf("terminal fragment native allocation")
	}
	allocation := &g.Allocations[token.Allocation]
	offset := int(token.Offset) + int(g.Offsets[step.Field])
	if offset != g.Page-1 || !bytes.Equal(allocation.Expected[offset:], before) {
		return fmt.Errorf("terminal fragment mapped prefix/BEFORE")
	}
	copy(allocation.Expected[offset:], after)
	return nil
}
func (c *rddNativeComparison4E0B30) acceptNativeWrite(step rddNativeStep4E0B30, data []byte) error {
	// No native write is performed here. Only expected canary and canonical
	// projection state are updated AFTER the independently executed store matched.
	g := c.Projection.Graph
	field := g.Definitions[step.Field]
	base := c.pointer(step.Base)
	token := g.Tokens[step.Base]
	if token.Kind != 1 || int(token.Allocation) >= len(g.Allocations) {
		return fmt.Errorf("actual parent write has no native allocation")
	}
	a := &g.Allocations[token.Allocation]
	offset := int(token.Offset) + int(g.Offsets[step.Field])
	native := data
	if field.pointer() {
		native = ddiBytes4E0B30(uint64(uintptr(c.pointer(uint32(ddiUnsigned4E0B30(data))))), 8)
	}
	if offset < 0 || offset+len(native) > g.Page {
		return fmt.Errorf("completed write crosses native guard")
	}
	copy(a.Expected[offset:], native)
	address := mustRDDNumber4E0B30(step.Event.Address)
	found := false
	for _, region := range g.Case.Input.Regions {
		if address < uint64(region.Address) || address+uint64(len(data)) > uint64(region.Address)+uint64(region.Size) {
			continue
		}
		at := int(address - uint64(region.Address))
		copy(c.Projection.Canonical[region.Name][at:], data)
		mirror := g.Mirrors[region.Name]
		ddiCopyIn4E0B30(unsafe.Add(mirror.Map, mirror.Base+at), data)
		copy(mirror.Expected[mirror.Base+at:], data)
		g.Mirrors[region.Name] = mirror
		found = true
		break
	}
	if !found {
		return fmt.Errorf("completed write lacks canonical region")
	}
	runtime.KeepAlive(base)
	return c.Projection.check()
}
func (c *rddNativeComparison4E0B30) drainEffects() {
	for c.Position < len(c.Steps) && c.Steps[c.Position].Kind == "effect" {
		step := c.Steps[c.Position]
		address, data, e := c.Projection.resolve(*step.Effect, step.EffectArgs)
		if e != nil {
			c.fail("input effect selector: %v", e)
		}
		if address == rddCaller4E0B30+16 {
			if c.Damage == nil || len(data) != 4 {
				c.fail("declared effect lacks actual native damage pointer")
			}
			*c.Damage = int32(uint32(ddiUnsigned4E0B30(data)))
		} else if e = c.Projection.apply(*step.Effect, step.EffectArgs); e != nil {
			c.fail("declared input mutation: %v", e)
		}
		c.record(c.effectObservation(step))
		c.Position++
		if address == rddCaller4E0B30+16 && c.CallerSink != nil {
			site := uint32(mustRDDNumber4E0B30(step.Event.ServiceSite))
			c.CallerSink(c.Position, rawDefaultDamageNativeCaller4E0B30{Site: site, Pointer: c.Damage, Value: uint32(*c.Damage), Complete: true})
		}
	}
}
func (c *rddNativeComparison4E0B30) call(call rawDefaultDamageNativeCall4E0B30) {
	c.record(rddNativeObserved4E0B30{Kind: "call", Site: call.Site, Entry: call.Entry, Complete: call.Complete, Result: call.Result, Arguments: rddObserveArguments4E0B30(call.Arguments, c.Damage)})
	step := c.next()
	kind := "call-before"
	if call.Complete {
		kind = "call-after"
	}
	frame := step.Frame
	if step.Kind != kind || call.Site != uint32(mustRDDNumber4E0B30(frame.CallSite)) || call.Entry != uint32(mustRDDNumber4E0B30(frame.Entry)) || len(call.Arguments) != len(frame.Args) {
		c.fail("actual original child/conversion call order")
	}
	for i, arg := range call.Arguments {
		if call.Entry == 0x419a70 {
			if arg.Kind != "word" || arg.Word != frame.Args[i] {
				c.fail("actual FST32 converter input bits")
			}
		} else if arg.Kind != "object" || arg.Pointer != c.pointer(frame.Args[i]) {
			c.fail("actual predicate full-width argument")
		}
	}
	if call.Complete && call.Result != frame.After["eax"] {
		c.fail("actual original predicate/conversion return bits")
	}
	if !call.Complete {
		c.ActualCalls = append(c.ActualCalls, rddActualChildFrame4E0B30{Entry: call.Entry})
	} else {
		if len(c.ActualCalls) == 0 || c.ActualCalls[len(c.ActualCalls)-1].Entry != call.Entry {
			c.fail("actual child/conversion call stack")
		}
		c.ActualCalls = c.ActualCalls[:len(c.ActualCalls)-1]
	}
	c.Position++
	c.drainEffects()
}
func (c *rddNativeComparison4E0B30) frame(site uint32, width int) uint32 {
	base := c.Projection.Graph.base(c.Projection.ClockIndex)
	event := rawDefaultDamageNativeOperand4E0B30{Site: site, Base: base, Address: base, Field: -1, Width: width}
	c.operand(event)
	value := uint32(rawParentInitialGoBits4E0B30(base, 0, width))
	event.Complete = true
	event.Value.Word = value
	c.operand(event)
	return value
}
func (c *rddNativeComparison4E0B30) install() func() {
	previous := damagePredicateRawObserve4E1400
	damagePredicateRawObserve4E1400 = func(complete bool, base unsafe.Pointer, kind int, value damagePredicateRawValue4E1400) {
		if len(c.ActualCalls) == 0 {
			c.fail("actual predicate read outside actual child invocation")
		}
		frame := &c.ActualCalls[len(c.ActualCalls)-1]
		site := uint32(0)
		field := 0
		if frame.Entry == 0x4e1400 {
			switch kind {
			case 0:
				site = 0x4e140c
				field = 1
			case 1:
				site = 0x4e141d
				field = 2
			case 2:
				site = 0x4e142c
				field = 1
			case 3:
				if !frame.ClassKnown {
					c.fail("actual weapon subclass before cached actual class")
				}
				field = 2
				if frame.WeaponClass&0x1000 != 0 {
					site = 0x4e1434
				} else {
					site = 0x4e1456
				}
			case 4:
				site = 0x4e143d
				field = 17
			case 5:
				site = 0x4e1443
				field = 25
			default:
				c.fail("unknown actual attack field kind")
			}
		} else if frame.Entry == 0x4e1470 {
			switch kind {
			case 2:
				site = 0x4e1478
				field = 1
			case 3:
				site = 0x4e1481
				field = 2
			default:
				c.fail("unknown actual friendly field kind")
			}
		} else {
			c.fail("actual predicate read inside another actual call")
		}
		width := 4
		if kind == 1 || kind == 5 {
			width = 1
		} else if kind == 4 {
			width = int(unsafe.Sizeof(uintptr(0)))
		}
		if complete && kind == 2 {
			frame.WeaponClass = value.Word
			frame.ClassKnown = true
		}
		address := base
		if base != nil {
			address = unsafe.Add(base, damagePredicateRawOffset4E1400(kind))
		}
		c.operand(rawDefaultDamageNativeOperand4E0B30{Site: site, Base: base, Address: address, Field: field, Width: width, Complete: complete, Value: rawDefaultDamageNativeValue4E0B30{Word: value.Word, Pointer: value.Pointer}})
	}

	pop := rawDefaultDamagePushNativeObserver4E0B30(&rawDefaultDamageNativeObserver4E0B30{ControlWord: uint16(c.Row.Control), Caller: c.caller, Operand: c.operand, Call: c.call, Frame: c.frame, Service: c.service, BeforeCallback: c.beforeCallback})
	return func() { pop(); damagePredicateRawObserve4E1400 = previous }
}
func (c *rddNativeComparison4E0B30) execute(backend string) (result int32, recovered any, err error) {
	if backend != "Go" && backend != "C" {
		return 0, nil, fmt.Errorf("unknown actual root backend")
	}
	if rawDefaultDamageNativeEngine4E0B30 == nil {
		return 0, nil, fmt.Errorf("actual native raw parent engine is absent")
	}
	if backend == "C" && rawParentABI4E0B30() != 2 {
		return 0, nil, fmt.Errorf("actual raw C symbol still has integer ABI")
	}
	pop := c.install()
	defer pop()
	defer func() { recovered = recover() }()
	args := c.Input.Args
	pointers := [3]unsafe.Pointer{c.pointer(args[0]), c.pointer(args[1]), c.pointer(args[2])}
	c.Backend = backend
	c.DispatchedRoot = true
	if backend == "Go" {
		result = rawDefaultDamageNativeEngine4E0B30(pointers[0], pointers[1], pointers[2], int32(args[3]), int32(args[4]))
	} else {
		result = rawParentPublic4E0B30(pointers[0], pointers[1], pointers[2], int32(args[3]), int32(args[4]))
	}
	runtime.KeepAlive(pointers)
	return
}
func (c *rddNativeComparison4E0B30) finishReturned(result int32) error {
	if !c.DispatchedRoot {
		return fmt.Errorf("no actual raw root dispatched; transport is not engine qualification")
	}
	if c.Row.Status != "returned" || c.Pending != nil || c.PendingCaller != nil || c.CallbackPending || c.Position != len(c.Steps) || uint32(result) != c.Row.RegistersAfter["eax"] {
		return fmt.Errorf("actual returned root/result/complete observable order")
	}
	if c.Damage == nil {
		return fmt.Errorf("actual native caller cell was never observed")
	}
	caller, e := hex.DecodeString(c.Row.CallerAfter)
	if e != nil || len(caller) != 24 || uint32(*c.Damage) != uint32(ddiUnsigned4E0B30(caller[16:20])) {
		return fmt.Errorf("actual returned native scalar AFTER differs from original caller damage word")
	}
	if e := c.Projection.check(); e != nil {
		return e
	}
	return nil
}
func (c *rddNativeComparison4E0B30) beforeCallback(site uint32, target unsafe.Pointer) {
	c.record(rddNativeObserved4E0B30{Kind: "callback", Site: site, Callback: uintptr(target)})
	// An invalid fetch has no entered-service event. Preserve the attempted
	// cached target as an actual invocation boundary for the supervised child.
	step := c.next()
	if c.CallbackPending {
		c.fail("duplicate actual callback attempt before entry")
	}
	if step.Kind == "service" {
		spec := c.Bundle.Calls[step.Event.Site]
		if !spec.Indirect || site != uint32(mustRDDNumber4E0B30(step.Event.Site)) || target != c.pointer(uint32(mustRDDNumber4E0B30(spec.Target))) {
			c.fail("actual callback before-invoke identity")
		}
		c.CallbackPending = true
		c.CallbackSite = site
		c.CallbackTarget = target
		return
	}
	if step.Kind != "fault" || step.Event.Access != 21 && step.Event.Access != 24 {
		c.fail("unexpected actual callback invocation")
	}
	var pending struct {
		Site   uint32
		Target *uint32
	}
	if e := json.Unmarshal(c.Row.PendingCall, &pending); e != nil {
		c.fail("pending callback certificate")
	}
	if pending.Target == nil || site != pending.Site || target != c.pointer(*pending.Target) {
		c.fail("actual cached callback target identity")
	}
	c.CallbackPending = true
	c.CallbackSite = site
	c.CallbackTarget = target
	if c.Progress != nil {
		c.Progress(c.Position, rawDefaultDamageNativeOperand4E0B30{Site: site, Address: target, Field: -2})
	}
}

// Terminal fragments affect only the expected bank before execution. Actual
// full AFTER pages must agree through the real ordered-byte store and fault.

type rddActualNativeFault4E0B30 struct {
	CallerAddress                  uintptr
	CallerBits                     uint32
	CallerValid                    bool
	Address, PC                    uintptr
	Site                           uint32
	Signal                         int
	Progress                       int
	FullAFTERArtifactSHA           string
	ActualPCBinaryQualificationSHA string
	AllActualNativePagesCompared   bool
}

func (c *rddNativeComparison4E0B30) finishFault(proof rddActualNativeFault4E0B30) error {
	if !c.DispatchedRoot {
		return fmt.Errorf("no actual raw root dispatched; control fault is not engine qualification")
	}
	if c.Row.Status != "original-fault" || (proof.Signal != 10 && proof.Signal != 11) || proof.FullAFTERArtifactSHA == "" || proof.ActualPCBinaryQualificationSHA == "" || !proof.AllActualNativePagesCompared || proof.Progress != c.Position {
		return fmt.Errorf("native fault needs raw full AFTER, binary-PC and exact completed-prefix proof")
	}
	step := c.next()
	faultPosition := c.Position
	if c.Pending != nil {
		if c.CallbackPending {
			return fmt.Errorf("pending callback before operand fault")
		}
		if (step.Kind != "read" && step.Kind != "write") || step.Event.Completed || proof.Site != uint32(mustRDDNumber4E0B30(step.Event.PC)) {
			return fmt.Errorf("native incomplete operand prefix identity")
		}
		address := uintptr(c.Pending.Address)
		width := uintptr(c.Pending.Width)
		if proof.Address < address || proof.Address >= address+width {
			return fmt.Errorf("actual operand fault address extent")
		}
		faultPosition++
	}
	if faultPosition >= len(c.Steps) || c.Steps[faultPosition].Kind != "fault" || faultPosition+1 != len(c.Steps) {
		return fmt.Errorf("native fault skipped or added observable original actions")
	}
	fault := c.Steps[faultPosition].Event
	if fault.Access == 21 || fault.Access == 24 {
		var pending struct {
			Site   uint32  `json:"site"`
			Target *uint32 `json:"target"`
		}
		if e := json.Unmarshal(c.Row.PendingCall, &pending); e != nil {
			return e
		}
		if pending.Target == nil {
			return fmt.Errorf("original invalid fetch lacks completed target")
		}
		target := uintptr(c.pointer(*pending.Target))
		if !c.CallbackPending || c.CallbackSite != pending.Site || uintptr(c.CallbackTarget) != target || proof.Site != pending.Site || proof.PC != target || proof.Address != target {
			return fmt.Errorf("actual attempted callback target/fetch identity")
		}
	}
	caller, e := hex.DecodeString(c.Row.CallerAfter)
	if e != nil || len(caller) != 24 || !proof.CallerValid || proof.CallerAddress != c.CallerNativeAddress || proof.CallerBits != uint32(ddiUnsigned4E0B30(caller[16:20])) {
		return fmt.Errorf("actual pinned physical scalar AFTER missing or differs from original caller damage word")
	}
	if c.Row.PartialCount != 0 && (c.Pending == nil || len(step.TerminalFragments) != c.Row.PartialCount || proof.Address != uintptr(c.Pending.Address)+1) {
		return fmt.Errorf("terminal WORD fragment lacks exact pending high-byte fault")
	}
	return nil
}

func (c *rddNativeComparison4E0B30) caller(event rawDefaultDamageNativeCaller4E0B30) {
	c.record(rddNativeObserved4E0B30{Kind: "caller", Site: event.Site, Complete: event.Complete, CallerEntry: event.Entry, Address: uintptr(unsafe.Pointer(event.Pointer)), Word: event.Value})
	step := c.next()
	if event.Pointer == nil {
		c.fail("nil actual caller-damage cell")
	}
	if event.Entry {
		if step.Kind != "caller-entry" || !event.Complete || event.Site != 0x4e0b30 || c.Damage != nil {
			c.fail("actual root caller entry")
		}
		data, e := hex.DecodeString(step.Event.After)
		if e != nil || event.Value != uint32(ddiUnsigned4E0B30(data)) {
			c.fail("actual initial by-value damage")
		}
		c.Damage = event.Pointer
		c.CallerNativeAddress = uintptr(unsafe.Pointer(event.Pointer))
		c.Position++
		if c.CallerSink != nil {
			c.CallerSink(c.Position, event)
		}
		c.drainEffects()
		return
	}
	if step.Kind != "caller-write" || event.Site != uint32(mustRDDNumber4E0B30(step.Event.PC)) || event.Pointer != c.Damage {
		c.fail("actual caller scalar write order/address")
	}
	if !event.Complete {
		if c.PendingCaller != nil {
			c.fail("nested caller write")
		}
		old, e := hex.DecodeString(step.Event.Before)
		if e != nil || event.Value != uint32(ddiUnsigned4E0B30(old)) {
			c.fail("actual caller scalar BEFORE bytes")
		}
		copy := event
		c.PendingCaller = &copy
		return
	}
	if c.PendingCaller == nil || !step.Event.Completed {
		c.fail("completed caller write outside original domain")
	}
	data, e := hex.DecodeString(step.Event.After)
	if e != nil || event.Value != uint32(ddiUnsigned4E0B30(data)) {
		c.fail("actual caller scalar narrowed AFTER bytes")
	}
	c.PendingCaller = nil
	c.Position++
	if c.CallerSink != nil {
		c.CallerSink(c.Position, event)
	}
	c.drainEffects()
}

// Exact immutable original memory operands, statically decoded only.
// These describe fault attempts, never a successful read or branch result.
type rddNativeStaticOperand4E0B30 struct {
	Base         string
	Displacement int64
	Width, Field int
	Access       int
	Encoding     string
}

var rddNativeStaticOperands4E0B30 = map[uint32]rddNativeStaticOperand4E0B30{
	0x004e0b52: {Access: 1, Base: "", Displacement: 8710660, Width: 1, Field: -1, Encoding: "f60504ea840003"},
	0x004e0b7b: {Access: 1, Base: "esi", Displacement: 8, Width: 1, Field: 1, Encoding: "8a4608"},
	0x004e0b86: {Access: 1, Base: "esi", Displacement: 748, Width: 4, Field: 18, Encoding: "8b86ec020000"},
	0x004e0b8f: {Access: 2, Base: "eax", Displacement: 2188, Width: 4, Field: 24, Encoding: "c7808c08000000000000"},
	0x004e0baf: {Access: 3, Base: "eax", Displacement: 1440, Width: 4, Field: 21, Encoding: "8188a005000000000800"},
	0x004e0bb9: {Access: 1, Base: "esi", Displacement: 16, Width: 4, Field: 3, Encoding: "8b4610"},
	0x004e0bde: {Access: 2, Base: "esi", Displacement: 520, Width: 4, Field: 9, Encoding: "898608020000"},
	0x004e0be4: {Access: 2, Base: "esi", Displacement: 524, Width: 4, Field: 10, Encoding: "89be0c020000"},
	0x004e0bea: {Access: 1, Base: "", Displacement: 8710660, Width: 4, Field: -1, Encoding: "8b0d04ea8400"},
	0x004e0bf1: {Access: 2, Base: "esi", Displacement: 536, Width: 4, Field: 13, Encoding: "898e18020000"},
	0x004e0c24: {Access: 1, Base: "ebp", Displacement: 8, Width: 1, Field: 1, Encoding: "f6450806"},
	0x004e0c6f: {Access: 1, Base: "esi", Displacement: 8, Width: 1, Field: 1, Encoding: "f6460806"},
	0x004e0c94: {Access: 1, Base: "esi", Displacement: 16, Width: 1, Field: 3, Encoding: "f6461002"},
	0x004e0cb1: {Access: 1, Base: "edi", Displacement: 8, Width: 1, Field: 1, Encoding: "f6470806"},
	0x004e0d05: {Access: 1, Base: "edi", Displacement: 716, Width: 4, Field: 15, Encoding: "ff97cc020000"},
	0x004e0d0b: {Access: 1, Base: "edi", Displacement: 8, Width: 1, Field: 1, Encoding: "8a4708"},
	0x004e0d20: {Access: 1, Base: "esi", Displacement: 8, Width: 1, Field: 1, Encoding: "8a4608"},
	0x004e0d29: {Access: 1, Base: "esi", Displacement: 12, Width: 4, Field: 2, Encoding: "8b460c"},
	0x004e0da6: {Access: 1, Base: "", Displacement: 5779580, Width: 4, Field: -4, Encoding: "d81d7c305800"},
	0x004e0db6: {Access: 1, Base: "", Displacement: 8710660, Width: 1, Field: -1, Encoding: "841d04ea8400"},
	0x004e0dcd: {Access: 1, Base: "", Displacement: 5779560, Width: 4, Field: -4, Encoding: "d90568305800"},
	0x004e0e0f: {Access: 1, Base: "", Displacement: 5779580, Width: 4, Field: -4, Encoding: "d81d7c305800"},
	0x004e0e1f: {Access: 1, Base: "", Displacement: 8710660, Width: 1, Field: -1, Encoding: "841d04ea8400"},
	0x004e0e36: {Access: 1, Base: "", Displacement: 5779560, Width: 4, Field: -4, Encoding: "d90568305800"},
	0x004e0e60: {Access: 1, Base: "esi", Displacement: 8, Width: 4, Field: 1, Encoding: "8b4608"},
	0x004e0e67: {Access: 1, Base: "esi", Displacement: 748, Width: 4, Field: 18, Encoding: "8b96ec020000"},
	0x004e0e6d: {Access: 2, Base: "edx", Displacement: 160, Width: 2, Field: 20, Encoding: "66c782a00000000200"},
	0x004e0e7c: {Access: 1, Base: "esi", Displacement: 12, Width: 1, Field: 2, Encoding: "f6460c10"},
	0x004e0e82: {Access: 1, Base: "esi", Displacement: 748, Width: 4, Field: 18, Encoding: "8b86ec020000"},
	0x004e0e88: {Access: 2, Base: "eax", Displacement: 2094, Width: 1, Field: 22, Encoding: "c6802e08000002"},
	0x004e0ea1: {Access: 1, Base: "ebx", Displacement: 8, Width: 4, Field: 1, Encoding: "f7430800100001"},
	0x004e0eaa: {Access: 1, Base: "edi", Displacement: 72, Width: 4, Field: 6, Encoding: "8b4f48"},
	0x004e0ead: {Access: 2, Base: "esi", Displacement: 528, Width: 4, Field: 11, Encoding: "898e10020000"},
	0x004e0eb3: {Access: 1, Base: "edi", Displacement: 76, Width: 4, Field: 7, Encoding: "8b574c"},
	0x004e0eb6: {Access: 2, Base: "esi", Displacement: 532, Width: 4, Field: 12, Encoding: "899614020000"},
	0x004e0ebe: {Access: 1, Base: "ebx", Displacement: 72, Width: 4, Field: 6, Encoding: "8b4348"},
	0x004e0ec1: {Access: 2, Base: "esi", Displacement: 528, Width: 4, Field: 11, Encoding: "898610020000"},
	0x004e0ec7: {Access: 1, Base: "ebx", Displacement: 76, Width: 4, Field: 7, Encoding: "8b4b4c"},
	0x004e0eca: {Access: 2, Base: "esi", Displacement: 532, Width: 4, Field: 12, Encoding: "898e14020000"},
	0x004e0ed4: {Access: 1, Base: "esi", Displacement: 8, Width: 1, Field: 1, Encoding: "f6460802"},
	0x004e0eda: {Access: 1, Base: "esi", Displacement: 748, Width: 4, Field: 18, Encoding: "8b86ec020000"},
	0x004e0ee2: {Access: 2, Base: "eax", Displacement: 2188, Width: 4, Field: 24, Encoding: "c7808c08000001000000"},
	0x004e0eec: {Access: 1, Base: "ebx", Displacement: 4, Width: 2, Field: 0, Encoding: "668b5304"},
	0x004e0ef2: {Access: 1, Base: "edi", Displacement: 72, Width: 4, Field: 6, Encoding: "8b4748"},
	0x004e0ef8: {Access: 2, Base: "esi", Displacement: 528, Width: 4, Field: 11, Encoding: "898610020000"},
	0x004e0efe: {Access: 1, Base: "edi", Displacement: 76, Width: 4, Field: 7, Encoding: "8b4f4c"},
	0x004e0f01: {Access: 2, Base: "esi", Displacement: 532, Width: 4, Field: 12, Encoding: "898e14020000"},
	0x004e0f0e: {Access: 1, Base: "esi", Displacement: 8, Width: 1, Field: 1, Encoding: "f6460802"},
	0x004e0f14: {Access: 1, Base: "esi", Displacement: 748, Width: 4, Field: 18, Encoding: "8b86ec020000"},
	0x004e0f1c: {Access: 2, Base: "eax", Displacement: 2188, Width: 4, Field: 24, Encoding: "c7808c08000001000000"},
	0x004e0f26: {Access: 1, Base: "edi", Displacement: 4, Width: 2, Field: 0, Encoding: "668b5704"},
	0x004e0f2a: {Access: 2, Base: "eax", Displacement: 2184, Width: 4, Field: 23, Encoding: "899088080000"},
	0x004e0f40: {Access: 2, Base: "esi", Displacement: 528, Width: 4, Field: 11, Encoding: "898610020000"},
	0x004e0f46: {Access: 2, Base: "esi", Displacement: 532, Width: 4, Field: 12, Encoding: "898614020000"},
	0x004e0f5d: {Access: 1, Base: "esi", Displacement: 8, Width: 4, Field: 1, Encoding: "8b4608"},
	0x004e0f68: {Access: 1, Base: "esi", Displacement: 12, Width: 1, Field: 2, Encoding: "f6460c10"},
	0x004e0f83: {Access: 2, Base: "esi", Displacement: 520, Width: 4, Field: 9, Encoding: "899e08020000"},
	0x004e0f8b: {Access: 2, Base: "esi", Displacement: 520, Width: 4, Field: 9, Encoding: "89be08020000"},
	0x004e0f91: {Access: 1, Base: "esi", Displacement: 8, Width: 1, Field: 1, Encoding: "8a4608"},
	0x004e0f94: {Access: 2, Base: "esi", Displacement: 524, Width: 4, Field: 10, Encoding: "89ae0c020000"},
	0x004e0f9a: {Access: 1, Base: "", Displacement: 8710660, Width: 4, Field: -1, Encoding: "8b0d04ea8400"},
	0x004e0fa2: {Access: 2, Base: "esi", Displacement: 536, Width: 4, Field: 13, Encoding: "898e18020000"},
	0x004e0faa: {Access: 1, Base: "esi", Displacement: 748, Width: 4, Field: 18, Encoding: "8b86ec020000"},
	0x004e0fb0: {Access: 1, Base: "eax", Displacement: 1440, Width: 4, Field: 21, Encoding: "8b90a0050000"},
	0x004e0fb6: {Access: 1, Base: "eax", Displacement: 2188, Width: 4, Field: 24, Encoding: "8b888c080000"},
	0x004e0fc1: {Access: 2, Base: "eax", Displacement: 1440, Width: 4, Field: 21, Encoding: "8990a0050000"},
	0x004e0fc9: {Access: 2, Base: "eax", Displacement: 2188, Width: 4, Field: 24, Encoding: "c7808c08000002000000"},
	0x004e0fd3: {Access: 2, Base: "eax", Displacement: 2184, Width: 4, Field: 23, Encoding: "89a888080000"},
	0x004e0fdd: {Access: 1, Base: "ebx", Displacement: 8, Width: 4, Field: 1, Encoding: "f7430800100001"},
	0x004e0ffb: {Access: 1, Base: "esi", Displacement: 8, Width: 4, Field: 1, Encoding: "f7460800100001"},
	0x004e1008: {Access: 1, Base: "edi", Displacement: 8, Width: 1, Field: 1, Encoding: "f6470802"},
	0x004e100e: {Access: 1, Base: "edi", Displacement: 748, Width: 4, Field: 18, Encoding: "8b87ec020000"},
	0x004e1025: {Access: 1, Base: "eax", Displacement: 32, Width: 4, Field: 26, Encoding: "8b4020"},
	0x004e1039: {Access: 1, Base: "esi", Displacement: 720, Width: 4, Field: 16, Encoding: "8b86d0020000"},
	0x004e106c: {Access: 1, Base: "esi", Displacement: 8, Width: 1, Field: 1, Encoding: "f6460806"},
	0x004e10e3: {Access: 1, Base: "edi", Displacement: 56, Width: 4, Field: 4, Encoding: "8b5738"},
	0x004e10f0: {Access: 1, Base: "edi", Displacement: 60, Width: 4, Field: 5, Encoding: "8b473c"},
	0x004e10f9: {Access: 1, Base: "esi", Displacement: 56, Width: 4, Field: 4, Encoding: "8b4e38"},
	0x004e1106: {Access: 1, Base: "esi", Displacement: 60, Width: 4, Field: 5, Encoding: "8b563c"},
	0x004e1136: {Access: 1, Base: "esi", Displacement: 8, Width: 1, Field: 1, Encoding: "8a4608"},
	0x004e1147: {Access: 1, Base: "esi", Displacement: 748, Width: 4, Field: 18, Encoding: "8b86ec020000"},
	0x004e114d: {Access: 1, Base: "eax", Displacement: 88, Width: 1, Field: 19, Encoding: "8a4058"},
	0x004e1188: {Access: 1, Base: "edi", Displacement: 8, Width: 1, Field: 1, Encoding: "f6470802"},
	0x004e1192: {Access: 1, Base: "edi", Displacement: 508, Width: 4, Field: 8, Encoding: "8b87fc010000"},
	0x004e119c: {Access: 1, Base: "eax", Displacement: 8, Width: 1, Field: 1, Encoding: "f6400802"},
	0x004e140c: {Access: 1, Base: "ecx", Displacement: 8, Width: 4, Field: 1, Encoding: "8b4108"},
	0x004e141d: {Access: 1, Base: "ecx", Displacement: 12, Width: 1, Field: 2, Encoding: "f6410c10"},
	0x004e142c: {Access: 1, Base: "ecx", Displacement: 8, Width: 4, Field: 1, Encoding: "8b4108"},
	0x004e1434: {Access: 1, Base: "ecx", Displacement: 12, Width: 4, Field: 2, Encoding: "f7410c00007f04"},
	0x004e143d: {Access: 1, Base: "ecx", Displacement: 736, Width: 4, Field: 17, Encoding: "8b89e0020000"},
	0x004e1443: {Access: 1, Base: "ecx", Displacement: 96, Width: 1, Field: 25, Encoding: "f6416002"},
	0x004e1456: {Access: 1, Base: "ecx", Displacement: 12, Width: 4, Field: 2, Encoding: "f7410cfe007f04"},
	0x004e1478: {Access: 1, Base: "eax", Displacement: 8, Width: 4, Field: 1, Encoding: "f7400800000001"},
	0x004e1481: {Access: 1, Base: "eax", Displacement: 12, Width: 4, Field: 2, Encoding: "8b480c"},
}
