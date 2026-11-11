//go:build darwin && arm64

package legacy

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"unsafe"
)

// Pointer fields below are NUMERIC observations only. A dead child address is
// never converted back to unsafe.Pointer or dereferenced by a parent verifier.
type rddNativeObservedArgument4E0B30 struct {
	Kind          string
	Word          uint32
	Pointer       uintptr
	DamageAddress uintptr
	DamageValue   uint32
	String        string
	Coordinates   [4]uint32
}
type rddNativeObserved4E0B30 struct {
	Sequence, Position                     int
	Kind                                   string
	Site, Entry                            uint32
	Field, Width                           int
	Complete, Write, CallerEntry, Indirect bool
	Base, Address, Pointer, Callback       uintptr
	Word, Result                           uint32
	Arguments                              []rddNativeObservedArgument4E0B30
}

func rddObserveArguments4E0B30(args []rawDefaultDamageNativeArgument4E0B30, knownDamage *int32) []rddNativeObservedArgument4E0B30 {
	out := make([]rddNativeObservedArgument4E0B30, len(args))
	for i, a := range args {
		out[i] = rddNativeObservedArgument4E0B30{Kind: a.Kind, Word: a.Word, Pointer: uintptr(a.Pointer), String: a.String, Coordinates: a.Coordinates}
		if a.Kind == "damage" {
			out[i].DamageAddress = uintptr(unsafe.Pointer(a.Damage))
			if a.Damage != nil && a.Damage == knownDamage {
				out[i].DamageValue = uint32(*a.Damage)
			}
		}
	}
	return out
}
func (c *rddNativeComparison4E0B30) record(event rddNativeObserved4E0B30) {
	event.Sequence = c.ObservedCount
	event.Position = c.Position
	c.ObservedCount++
	if c.Record != nil {
		if e := c.Record(event); e != nil {
			c.fail("required actual observation retention: %v", e)
		}
	}
}
func (c *rddNativeComparison4E0B30) effectObservation(step rddNativeStep4E0B30) rddNativeObserved4E0B30 {
	address, data, e := c.Projection.resolve(*step.Effect, step.EffectArgs)
	if e != nil {
		c.fail("effect actual observation: %v", e)
	}
	out := rddNativeObserved4E0B30{Kind: "effect", Complete: true, Field: -3, Width: len(data)}
	if address == rddCaller4E0B30+16 {
		out.Address = uintptr(unsafe.Pointer(c.Damage))
		out.Word = uint32(*c.Damage)
		return out
	}
	g := c.Projection.Graph
	if address >= 0x84ea04 && address+uint64(len(data)) <= 0x84ea08 {
		out.Field = -1
		out.Base = uintptr(g.base(c.Projection.ClockIndex))
		out.Address = out.Base + uintptr(address-0x84ea04)
		out.Word = uint32(rawParentInitialGoBits4E0B30(unsafe.Add(g.base(c.Projection.ClockIndex), uintptr(address-0x84ea04)), 0, len(data)))
		return out
	}
	for _, ef := range g.Fields {
		d := g.Definitions[ef.Field]
		start := uint64(g.Case.Input.Roles[ef.Role]) + uint64(d.Offset)
		if address < start || address+uint64(len(data)) > start+uint64(d.Width) {
			continue
		}
		index := ddiRoleIndex4E0B30(ef.Role)
		out.Field = d.ID
		out.Base = uintptr(g.base(index))
		offset := g.Offsets[d.ID] + uintptr(address-start)
		out.Address = out.Base + offset
		if d.pointer() {
			out.Width = int(unsafe.Sizeof(uintptr(0)))
			out.Pointer = rawParentInitialCBits4E0B30(g.base(index), offset, 0)
		} else {
			out.Word = uint32(rawParentInitialCBits4E0B30(g.base(index), offset, len(data)))
		}
		return out
	}
	c.fail("effect observation lacks exact field projection")
	return out
}

type rddRetainedNativeFault4E0B30 struct {
	NativeTokens                                                                                  map[uint32]uintptr
	NativeOffsets                                                                                 [27]uintptr
	ClockAddress, CallerAddress                                                                   uintptr
	CallerBits                                                                                    uint32
	CallerValid                                                                                   bool
	Signal                                                                                        int
	FaultPC, FaultAddress                                                                         uintptr
	FaultSite                                                                                     uint32
	FaultOperation, FaultWidth, FaultProgress                                                     int
	BinaryPCQualificationSHA, ActualExecutableSHA, SamePIDPREEXECReceiptSHA, FullAFTERArtifactSHA string
	ObservableJSONL                                                                               []byte
	AllActualNativePagesCompared, ValidationFailed, ActualRootDispatched                          bool
	Backend                                                                                       string
	RootEntry                                                                                     uintptr
	FaultClassification                                                                           string
	FaultPhysicalWidth, FaultByteOffset                                                           int
}

func rddVerifyRetainedNativeFault4E0B30(row *rddRow4E0B30, plan *rddNativeComparison4E0B30, p rddRetainedNativeFault4E0B30) error {
	bad := func(s string) error { return fmt.Errorf("retained native %s/%s: %s", row.Name, row.Mode, s) }
	if row.Status != "original-fault" || p.ValidationFailed || !p.ActualRootDispatched || p.Backend != "Go" && p.Backend != "C" || p.RootEntry <= 0xffffffff || !p.AllActualNativePagesCompared || (p.Signal != 10 && p.Signal != 11) || p.BinaryPCQualificationSHA == "" || p.ActualExecutableSHA == "" || p.SamePIDPREEXECReceiptSHA == "" || p.FullAFTERArtifactSHA == "" {
		return bad("missing actual source/root/PC/whole-AFTER proof or validation failure")
	}
	if p.NativeOffsets != plan.Projection.Graph.Offsets {
		return bad("native field offsets differ from actual frozen layout")
	}
	zero, zeroOK := p.NativeTokens[0]
	if !zeroOK || zero != 0 || p.ClockAddress <= 0xffffffff {
		return bad("nil/clock native identity")
	}
	identities := map[uintptr]uint32{}
	for word, bits := range p.NativeTokens {
		if word != 0 && bits <= 0xffffffff {
			return bad("native pointer truncation")
		}
		if old, ok := identities[bits]; ok && old != word {
			return bad("collapsed native pointer identities")
		}
		identities[bits] = word
	}
	ptr := func(word uint32) (uintptr, error) {
		v, ok := p.NativeTokens[word]
		if !ok {
			return 0, bad("canonical pointer token missing")
		}
		return v, nil
	}
	reader := bufio.NewReader(bytes.NewReader(p.ObservableJSONL))
	position, sequence := 0, 0
	pendingOperand, pendingCaller := false, false
	callbackAttempt := false
	callbackPending := false
	callbackSite := uint32(0)
	callbackTarget := uintptr(0)
	caller := uintptr(0)
	for {
		line, e := reader.ReadBytes('\n')
		if e == io.EOF && len(line) == 0 {
			break
		}
		if e != nil {
			return bad("partial or unreadable actual observation JSONL")
		}
		var actual rddNativeObserved4E0B30
		if e = json.Unmarshal(line, &actual); e != nil {
			return e
		}
		if actual.Sequence != sequence || actual.Position != position || position >= len(plan.Steps) {
			return bad("actual observation ordinal/progress")
		}
		sequence++
		step := plan.Steps[position]
		switch actual.Kind {
		case "operand":
			kind := "read"
			if actual.Write {
				kind = "write"
			}
			if step.Kind != kind || actual.Site != uint32(mustRDDNumber4E0B30(step.Event.PC)) || actual.Field != step.Field {
				return bad("actual ordered field operation")
			}
			base := uintptr(0)
			if step.Field == -1 {
				base = p.ClockAddress
			} else if step.Field >= 0 {
				base, e = ptr(step.Base)
				if e != nil {
					return e
				}
			}
			address := base
			width := step.Event.Width
			pointer := step.Field >= 0 && plan.Projection.Graph.Definitions[step.Field].pointer()
			if step.Field >= 0 && base != 0 {
				address += p.NativeOffsets[step.Field]
			}
			if pointer {
				width = 8
			}
			if actual.Base != base || actual.Address != address || actual.Width != width {
				return bad("actual numeric native operand extent")
			}
			if actual.Write {
				if pointer {
					want, e := ptr(uint32(step.Event.Value))
					if e != nil {
						return e
					}
					if actual.Pointer != want {
						return bad("actual attempted pointer store bits")
					}
				} else {
					mask := uint64(0xffffffff)
					if step.Event.Width == 1 {
						mask = 0xff
					} else if step.Event.Width == 2 {
						mask = 0xffff
					}
					if uint64(actual.Word) != step.Event.Value&mask {
						return bad("actual attempted scalar store bits")
					}
				}
			}
			if !actual.Complete {
				if pendingOperand {
					return bad("nested actual operand")
				}
				pendingOperand = true
				continue
			}
			if !pendingOperand || !step.Event.Completed {
				return bad("actual completed an original fault operand")
			}
			var data []byte
			if actual.Write {
				data, e = hex.DecodeString(step.Event.After)
			} else if step.Event.ValueHex != nil {
				data, e = hex.DecodeString(*step.Event.ValueHex)
			} else {
				return bad("completed original read without bytes")
			}
			if e != nil {
				return e
			}
			if pointer {
				want, e := ptr(uint32(ddiUnsigned4E0B30(data)))
				if e != nil {
					return e
				}
				if actual.Pointer != want {
					return bad("actual full pointer load/store")
				}
			} else if uint64(actual.Word) != ddiUnsigned4E0B30(data) {
				return bad("actual scalar read/store")
			}
			pendingOperand = false
			position++
		case "caller":
			if actual.Address == 0 {
				return bad("actual caller cell absent")
			}
			if actual.CallerEntry {
				if step.Kind != "caller-entry" || !actual.Complete || actual.Site != 0x4e0b30 || caller != 0 {
					return bad("actual caller entry order")
				}
				caller = actual.Address
			} else if step.Kind != "caller-write" || actual.Address != caller || actual.Site != uint32(mustRDDNumber4E0B30(step.Event.PC)) {
				return bad("actual caller assignment order")
			}
			text := step.Event.Before
			if actual.Complete {
				text = step.Event.After
			}
			data, e := hex.DecodeString(text)
			if e != nil || actual.Word != uint32(ddiUnsigned4E0B30(data)) {
				return bad("actual caller assignment bytes")
			}
			if actual.Complete {
				if !actual.CallerEntry && !pendingCaller {
					return bad("caller AFTER without BEFORE")
				}
				pendingCaller = false
				position++
			} else {
				if pendingCaller {
					return bad("nested caller assignment")
				}
				pendingCaller = true
			}
		case "call":
			kind := "call-before"
			if actual.Complete {
				kind = "call-after"
			}
			f := step.Frame
			if step.Kind != kind || actual.Site != uint32(mustRDDNumber4E0B30(f.CallSite)) || actual.Entry != uint32(mustRDDNumber4E0B30(f.Entry)) || len(actual.Arguments) != len(f.Args) {
				return bad("actual predicate/conversion call identity")
			}
			for i, arg := range actual.Arguments {
				if actual.Entry == 0x419a70 {
					if arg.Kind != "word" || arg.Word != f.Args[i] {
						return bad("actual converter f32 input")
					}
				} else {
					want, e := ptr(f.Args[i])
					if e != nil {
						return e
					}
					if arg.Kind != "object" || arg.Pointer != want {
						return bad("actual predicate pointer input")
					}
				}
			}
			if actual.Complete && actual.Result != f.After["eax"] {
				return bad("actual predicate/converter return")
			}
			position++
		case "service":
			if step.Kind != "service" || actual.Site != uint32(mustRDDNumber4E0B30(step.Event.Site)) || len(actual.Arguments) != len(step.ArgumentKinds) {
				return bad("actual external semantic call order")
			}
			spec := plan.Bundle.Calls[step.Event.Site]
			if spec.Indirect != actual.Indirect {
				return bad("actual explicit direct/indirect service kind")
			}
			if spec.Indirect {
				if !callbackPending || callbackSite != actual.Site || callbackTarget != actual.Callback {
					return bad("indirect service without exactly one actual callback attempt")
				}
				want, e := ptr(uint32(mustRDDNumber4E0B30(spec.Target)))
				if e != nil {
					return e
				}
				if actual.Callback != want {
					return bad("actual cached service callback identity")
				}
			}
			for i, arg := range actual.Arguments {
				if arg.Kind != step.ArgumentKinds[i] {
					return bad("actual argument semantic kind")
				}
				switch arg.Kind {
				case "word":
					if arg.Word != step.Event.Args[i] {
						return bad("actual full DWORD argument")
					}
				case "object":
					want, e := ptr(step.Event.Args[i])
					if e != nil {
						return e
					}
					if arg.Pointer != want {
						return bad("actual full object argument")
					}
				case "damage":
					if arg.DamageAddress != caller || arg.DamageValue != step.DamageBefore[i] {
						return bad("actual caller pointer/pre-service bits")
					}
				case "string":
					if arg.String != step.Strings[i] {
						return bad("actual string content")
					}
				case "coordinates":
					if arg.Coordinates != step.Coordinates[i] {
						return bad("actual four coordinate DWORDs")
					}
				}
			}
			if !spec.Indirect && callbackPending {
				return bad("pending callback before direct service")
			}
			callbackPending = false
			callbackSite = 0
			callbackTarget = 0
			position++
		case "effect":
			if step.Kind != "effect" || !actual.Complete {
				return bad("actual independent input effect order")
			}
			address, data, e := plan.Projection.resolve(*step.Effect, step.EffectArgs)
			if e != nil {
				return e
			}
			if address == rddCaller4E0B30+16 {
				if actual.Address != caller || actual.Word != uint32(ddiUnsigned4E0B30(data)) {
					return bad("actual inout effect AFTER")
				}
			} else if e = rddRetainedEffect4E0B30(plan, p, actual, address, data); e != nil {
				return e
			}
			position++
		case "callback":
			if callbackPending {
				return bad("duplicate actual callback attempt")
			}
			callbackPending = true
			callbackSite = actual.Site
			callbackTarget = actual.Callback
			if step.Kind == "service" {
				spec := plan.Bundle.Calls[step.Event.Site]
				want, e := ptr(uint32(mustRDDNumber4E0B30(spec.Target)))
				if e != nil {
					return e
				}
				if !spec.Indirect || actual.Site != uint32(mustRDDNumber4E0B30(step.Event.Site)) || actual.Callback != want {
					return bad("actual before-callback identity")
				}
			} else if step.Kind == "fault" && (step.Event.Access == 21 || step.Event.Access == 24) {
				var call struct {
					Site   uint32
					Target *uint32
				}
				if e = json.Unmarshal(row.PendingCall, &call); e != nil {
					return e
				}
				if call.Target == nil {
					return bad("missing original failed target")
				}
				want, e := ptr(*call.Target)
				if e != nil {
					return e
				}
				if actual.Site != call.Site || actual.Callback != want {
					return bad("actual invalid cached callback attempt")
				}
				callbackAttempt = true
			} else {
				return bad("unexpected actual callback attempt")
			}
		default:
			return bad("unknown actual observation kind")
		}
	}
	if sequence == 0 || pendingCaller || position != p.FaultProgress || caller == 0 || caller != p.CallerAddress || !p.CallerValid {
		return bad("actual complete-prefix/caller validity")
	}
	callerAfter, e := hex.DecodeString(row.CallerAfter)
	if e != nil || len(callerAfter) != 24 || p.CallerBits != uint32(ddiUnsigned4E0B30(callerAfter[16:20])) {
		return bad("sampled latest physical scalar AFTER")
	}
	if position >= len(plan.Steps) {
		return bad("fault after original trace ended")
	}
	faultPosition := position
	step := plan.Steps[position]
	if pendingOperand {
		if callbackPending {
			return bad("pending callback before operand fault")
		}
		if (step.Kind != "read" && step.Kind != "write") || step.Event.Completed || p.FaultSite != uint32(mustRDDNumber4E0B30(step.Event.PC)) {
			return bad("original incomplete operand identity")
		}
		base, e := ptr(step.Base)
		if e != nil {
			return e
		}
		address := base
		if base != 0 {
			address += p.NativeOffsets[step.Field]
		}
		width := step.Event.Width
		if plan.Projection.Graph.Definitions[step.Field].pointer() {
			width = 8
		}
		operation := rawDefaultDamageSharedRead4E0B30
		if step.Kind == "write" {
			operation = rawDefaultDamageSharedWrite4E0B30
		}
		if p.FaultAddress < address || p.FaultAddress >= address+uintptr(width) || p.FaultOperation != operation || p.FaultWidth != width {
			return bad("actual field fault operand/domain")
		}
		faultPosition++
	}
	if faultPosition+1 != len(plan.Steps) || plan.Steps[faultPosition].Kind != "fault" {
		return bad("missing/skipped terminal fault")
	}
	fault := plan.Steps[faultPosition].Event
	if fault.Access == 21 || fault.Access == 24 {
		var call struct {
			Site   uint32
			Target *uint32
		}
		if e = json.Unmarshal(row.PendingCall, &call); e != nil {
			return e
		}
		if call.Target == nil || !callbackAttempt || !callbackPending {
			return bad("missing actual failed callback invocation")
		}
		target, e := ptr(*call.Target)
		if e != nil {
			return e
		}
		if p.FaultPC != target || p.FaultAddress != target || p.FaultSite != call.Site || p.FaultOperation != 2 {
			return bad("actual invalid fetch identity")
		}
	}
	if row.PartialCount != 0 && (!pendingOperand || step.Kind != "write" || step.Event.Width != 2 || len(step.TerminalFragments) != row.PartialCount || p.FaultClassification != "incomplete-ordered-WORD-access" || p.FaultPhysicalWidth != 1 || p.FaultByteOffset != 1) {
		return bad("terminal fragment requires exact ordered WORD high-byte PC proof")
	}
	return nil
}
func rddRetainedEffect4E0B30(plan *rddNativeComparison4E0B30, p rddRetainedNativeFault4E0B30, actual rddNativeObserved4E0B30, address uint64, data []byte) error {
	if address >= 0x84ea04 && address+uint64(len(data)) <= 0x84ea08 {
		if actual.Field != -1 || actual.Address != p.ClockAddress+uintptr(address-0x84ea04) || actual.Width != len(data) || uint64(actual.Word) != ddiUnsigned4E0B30(data) {
			return fmt.Errorf("actual mutable frame effect")
		}
		return nil
	}
	g := plan.Projection.Graph
	for _, ef := range g.Fields {
		field := g.Definitions[ef.Field]
		start := uint64(g.Case.Input.Roles[ef.Role]) + uint64(field.Offset)
		if address < start || address+uint64(len(data)) > start+uint64(field.Width) {
			continue
		}
		base, ok := p.NativeTokens[g.Case.Input.Roles[ef.Role]]
		if !ok {
			return fmt.Errorf("effect native role absent")
		}
		if actual.Field != field.ID || actual.Base != base || actual.Address != base+p.NativeOffsets[field.ID]+uintptr(address-start) {
			return fmt.Errorf("actual declared field effect address")
		}
		if field.pointer() {
			want, ok := p.NativeTokens[uint32(ddiUnsigned4E0B30(data))]
			if !ok || actual.Width != 8 || actual.Pointer != want {
				return fmt.Errorf("actual full pointer input effect")
			}
		} else if actual.Width != len(data) || uint64(actual.Word) != ddiUnsigned4E0B30(data) {
			return fmt.Errorf("actual scalar input effect")
		}
		return nil
	}
	return fmt.Errorf("actual effect is outside frozen selected storage")
}
