//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const flagsInputSHA409E40 = "12c9afc88bb8c7c6e6fc795ca9756972238cb9c211e09dadfb18f505c06a6c52"
const flagsReferenceSHA409E40 = "13bcaf6088c6f8a6a5e79fe1295e0d747f8e552a1f8f78e0c6f84d0a1abad3b4"

var flagsPECells409E40 = [3]uint32{0x5d5330, 0x5d5360, 0x5d53a4}

type flagsInputWrite409E40 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type flagsInputStimulus409E40 struct {
	PC         uint32                  `json:"pc"`
	Site       string                  `json:"site"`
	Address    uint32                  `json:"address"`
	Width      uint32                  `json:"width"`
	Occurrence int                     `json:"occurrence"`
	Writes     []flagsInputWrite409E40 `json:"writes"`
}
type flagsInputPacket409E40 struct {
	Target    uint32                  `json:"target"`
	Argc      int                     `json:"argc32"`
	Registers map[string]uint32       `json:"register_effects_u32"`
	Writes    []flagsInputWrite409E40 `json:"writes"`
}
type flagsInput409E40 struct {
	Name          string                            `json:"name"`
	Entry         uint32                            `json:"entry"`
	Args          []uint32                          `json:"args_u32"`
	Words         map[string]uint32                 `json:"initial_words"`
	Services      map[string]flagsInputPacket409E40 `json:"services"`
	AfterReads    []flagsInputStimulus409E40        `json:"after_reads"`
	AfterWrites   []flagsInputStimulus409E40        `json:"after_writes"`
	AfterServices []flagsInputStimulus409E40        `json:"after_services"`
}
type flagsOriginalEvent409E40 struct {
	Kind      string          `json:"kind"`
	PC        uint32          `json:"PC"`
	Address   uint32          `json:"address"`
	Width     uint32          `json:"width"`
	Sequence  int             `json:"sequence"`
	Value     json.RawMessage `json:"value"`
	Completed bool            `json:"completed"`
}
type flagsOriginalService409E40 struct {
	Site      uint32   `json:"site"`
	Target    uint32   `json:"target"`
	Arguments []uint32 `json:"arguments_u32"`
	Frame     string   `json:"frame_hex"`
}
type flagsOriginalMemory409E40 struct {
	Address uint32 `json:"address"`
	Before  string `json:"before"`
	After   string `json:"after"`
	Exact   bool   `json:"replay_exact"`
}
type flagsOriginalRow409E40 struct {
	Name           string                       `json:"name"`
	Entry          uint32                       `json:"entry"`
	CW             uint16                       `json:"control_word"`
	InputSHA       string                       `json:"input_sha256"`
	Status         string                       `json:"status"`
	Valid          bool                         `json:"certificate_valid"`
	Result         *uint32                      `json:"actual_EAX_u32"`
	Events         []flagsOriginalEvent409E40   `json:"events"`
	Services       []flagsOriginalService409E40 `json:"services"`
	Memory         []flagsOriginalMemory409E40  `json:"memory"`
	Used           map[string][]int             `json:"used_mutations"`
	Unused         map[string][]int             `json:"unused_mutations"`
	UnusedServices []uint32                     `json:"unused_service_sites"`
	HelperOpcodes  int                          `json:"helper_original_opcodes_executed"`
	Cleanup        []string                     `json:"hook_cleanup_errors"`
}
type flagsStep409E40 struct {
	Kind     string `json:"kind"`
	Site     uint32 `json:"site"`
	Field    int    `json:"field"`
	Word     uint32 `json:"word"`
	Mask     uint32 `json:"mask"`
	Sequence int    `json:"original_sequence,omitempty"`
}
type flagsReference409E40 struct {
	CW             uint16
	Result         uint32
	Steps          []flagsStep409E40
	After          [3]uint32
	Used           map[string][]int
	Unused         map[string][]int
	UnusedServices []uint32
}
type flagsChosen409E40 struct {
	Input      flagsInput409E40
	Raw        json.RawMessage
	References []*flagsReference409E40
}

func flagsHash409E40(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func flagsHex409E40(text string) ([]byte, error) { return hex.DecodeString(text) }
func flagsPEField409E40(address uint32) int {
	for i, a := range flagsPECells409E40 {
		if address == a {
			return i
		}
	}
	return -1
}
func flagsOpenFixture409E40(name, sha string) (*os.File, *gzip.Reader, error) {
	path := filepath.Join("testdata", "setting_flags_409e40", name)
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, nil, e
	}
	if flagsHash409E40(data) != sha {
		return nil, nil, fmt.Errorf("frozen five-flags fixture SHA %s", name)
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, nil, e
	}
	z, e := gzip.NewReader(f)
	if e != nil {
		f.Close()
		return nil, nil, e
	}
	return f, z, nil
}

func flagsReduceReference409E40(row *flagsOriginalRow409E40) (*flagsReference409E40, error) {
	if !row.Valid || row.Status != "returned" || row.Result == nil || row.HelperOpcodes != 0 || len(row.Cleanup) != 0 {
		return nil, fmt.Errorf("authentic original returned certificate")
	}
	out := &flagsReference409E40{CW: row.CW, Result: *row.Result, Used: row.Used, Unused: row.Unused, UnusedServices: row.UnusedServices}
	haveData := false
	for _, memory := range row.Memory {
		if memory.Address != 0x5d5000 {
			continue
		}
		data, e := flagsHex409E40(memory.After)
		if e != nil || len(data) != 4096 || !memory.Exact {
			return nil, fmt.Errorf("original full state page AFTER")
		}
		for i, a := range flagsPECells409E40 {
			out.After[i] = binary.LittleEndian.Uint32(data[a-memory.Address:])
		}
		haveData = true
	}
	if !haveData {
		return nil, fmt.Errorf("original state page absent")
	}
	for _, event := range row.Events {
		field := flagsPEField409E40(event.Address)
		if field < 0 || event.Kind != "read" && event.Kind != "write" {
			continue
		}
		if field == 2 || event.Width != 4 || !event.Completed {
			return nil, fmt.Errorf("original caller scalar operand domain")
		}
		step := flagsStep409E40{Kind: event.Kind, Site: event.PC, Field: field, Sequence: event.Sequence}
		if event.Kind == "read" {
			var text string
			if e := json.Unmarshal(event.Value, &text); e != nil {
				return nil, e
			}
			data, e := flagsHex409E40(text)
			if e != nil || len(data) != 4 {
				return nil, fmt.Errorf("original DWORD load")
			}
			step.Word = binary.LittleEndian.Uint32(data)
		} else {
			if e := json.Unmarshal(event.Value, &step.Word); e != nil {
				return nil, e
			}
		}
		out.Steps = append(out.Steps, step)
	}
	for _, service := range row.Services {
		sequence := -1
		for _, event := range row.Events {
			if event.Kind != "host-frame-read" || event.PC != service.Target {
				continue
			}
			var text string
			if e := json.Unmarshal(event.Value, &text); e != nil {
				return nil, e
			}
			frame, e := flagsHex409E40(text)
			if e != nil || len(frame) < 4 {
				return nil, fmt.Errorf("original actual CALL frame")
			}
			if binary.LittleEndian.Uint32(frame) != service.Site+5 {
				continue
			}
			if sequence >= 0 || text != service.Frame || len(frame) != 4*(1+len(service.Arguments)) {
				return nil, fmt.Errorf("unique original service frame")
			}
			for i, a := range service.Arguments {
				if binary.LittleEndian.Uint32(frame[4+i*4:]) != a {
					return nil, fmt.Errorf("original service argument frame binding")
				}
			}
			sequence = event.Sequence
		}
		if sequence < 0 {
			return nil, fmt.Errorf("original service actual CALL frame absent")
		}
		mask := uint32(0)
		switch service.Site {
		case 0x409e90, 0x409f0a, 0x409f52:
			if len(service.Arguments) != 1 {
				return nil, fmt.Errorf("original Check one-word ABI")
			}
			mask = service.Arguments[0]
		case 0x409ea4, 0x409f28:
			if len(service.Arguments) != 0 {
				return nil, fmt.Errorf("original Refresh zero-word ABI")
			}
		default:
			return nil, fmt.Errorf("original external service site")
		}
		out.Steps = append(out.Steps, flagsStep409E40{Kind: "service", Site: service.Site, Mask: mask, Sequence: sequence})
	}
	sort.Slice(out.Steps, func(i, j int) bool { return out.Steps[i].Sequence < out.Steps[j].Sequence })
	return out, nil
}

func flagsLoadChosen409E40() ([]flagsChosen409E40, error) {
	f, z, e := flagsOpenFixture409E40("inputs-v1.json.gz", flagsInputSHA409E40)
	if e != nil {
		return nil, e
	}
	var inputs struct {
		Cases []json.RawMessage `json:"cases"`
	}
	e = json.NewDecoder(z).Decode(&inputs)
	z.Close()
	f.Close()
	if e != nil {
		return nil, e
	}
	if len(inputs.Cases) != 1261 {
		return nil, fmt.Errorf("independent 1261 cases")
	}
	out := make([]flagsChosen409E40, len(inputs.Cases))
	indices := make(map[string]int)
	for i, raw := range inputs.Cases {
		var compact bytes.Buffer
		if e = json.Compact(&compact, raw); e != nil {
			return nil, e
		}
		out[i].Raw = append(json.RawMessage(nil), compact.Bytes()...)
		if e = json.Unmarshal(raw, &out[i].Input); e != nil {
			return nil, e
		}
		in := &out[i].Input
		if len(in.Args) != 1 || len(in.Words) != 3 || len(in.Services) != 5 {
			return nil, fmt.Errorf("raw input argument/state/service domain")
		}
		if _, ok := indices[in.Name]; ok {
			return nil, fmt.Errorf("unique input names")
		}
		indices[in.Name] = i
	}
	f, z, e = flagsOpenFixture409E40("reference-v1.json.gz", flagsReferenceSHA409E40)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	defer z.Close()
	d := json.NewDecoder(z)
	tok, e := d.Token()
	if e != nil || tok != json.Delim('{') {
		return nil, fmt.Errorf("original reference object")
	}
	found, count := false, 0
	for d.More() {
		key, e := d.Token()
		if e != nil {
			return nil, e
		}
		if key != "rows" {
			var raw json.RawMessage
			if e = d.Decode(&raw); e != nil {
				return nil, e
			}
			if key == "fatal" && !bytes.Equal(raw, []byte("null")) {
				return nil, fmt.Errorf("original recorder fatal")
			}
			continue
		}
		if found {
			return nil, fmt.Errorf("duplicate original rows")
		}
		found = true
		tok, e = d.Token()
		if e != nil || tok != json.Delim('[') {
			return nil, fmt.Errorf("original rows array")
		}
		for d.More() {
			var row flagsOriginalRow409E40
			if e = d.Decode(&row); e != nil {
				return nil, e
			}
			i, ok := indices[row.Name]
			if !ok || row.Entry != out[i].Input.Entry || row.InputSHA != flagsHash409E40(out[i].Raw) || row.CW != 0x0e3f && row.CW != 0x023f {
				return nil, fmt.Errorf("original case/input/CW binding")
			}
			for _, prior := range out[i].References {
				if prior.CW == row.CW {
					return nil, fmt.Errorf("duplicate original case/CW")
				}
			}
			ref, e := flagsReduceReference409E40(&row)
			if e != nil {
				return nil, e
			}
			out[i].References = append(out[i].References, ref)
			count++
		}
		if _, e = d.Token(); e != nil {
			return nil, e
		}
	}
	if _, e = d.Token(); e != nil {
		return nil, e
	}
	var tail any
	if e = d.Decode(&tail); e != io.EOF {
		return nil, fmt.Errorf("original complete gzip JSON/CRC: %v", e)
	}
	if !found || count != 2522 {
		return nil, fmt.Errorf("actual original 2522 rows")
	}
	for i := range out {
		if len(out[i].References) != 2 {
			return nil, fmt.Errorf("two authentic CW references per input")
		}
	}
	return out, nil
}
