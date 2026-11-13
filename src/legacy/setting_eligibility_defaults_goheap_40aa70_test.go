//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sync"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const seExtraInputsSHA40AA70 = "808b1c79535399ce4d08db7160dfd1f9453b86ac368314e3f777a15fbe42eb18"
const seExtraReferenceSHA40AA70 = "03cad4011f68d89fe502ed93967281b994d5c623d525c85b0bb21dc4f5836ec2"

var seExtraNames40AA70 = []string{
	"0040aa70/nil-argument-final-check-packets/check128-00000000-check1024-00000000",
	"0040aa70/status-DWORD-and-global-DWORD/status-00000100-gate-00000000",
	"0040aa70/raw-check-and-signed-count770/check-00000001-count-00000000",
	"0040aa70/single-site-raw-full-EAX/site-0040aacf-word-00000001",
	"0040aa70/physical-alias-partitions/partition-1-arg-nil-0",
}

type seExtraPacket40AA70 struct {
	Registers map[string]uint32 `json:"register_effects_u32"`
	Writes    []json.RawMessage `json:"writes"`
}
type seExtraInput40AA70 struct {
	Name    string            `json:"name"`
	Args    []uint32          `json:"args_u32"`
	Roles   map[string]uint32 `json:"roles"`
	Regions []struct {
		Base       uint32 `json:"base"`
		Accessible int    `json:"accessible_bytes"`
		Initial    string `json:"initial_hex"`
	} `json:"regions"`
	InitialWords map[string]uint32 `json:"initial_words"`
	Services     map[string]struct {
		Target    uint32                `json:"target"`
		Argc      int                   `json:"argc32"`
		Responses []seExtraPacket40AA70 `json:"responses"`
		Repeat    bool                  `json:"repeat_last"`
	} `json:"services"`
	AfterReads    []json.RawMessage `json:"after_reads"`
	AfterServices []json.RawMessage `json:"after_services"`
	RawSHA        string            `json:"-"`
}
type seExtraReference40AA70 struct {
	Name     string  `json:"name"`
	CW       uint16  `json:"control_word"`
	InputSHA string  `json:"input_sha256"`
	Status   string  `json:"status"`
	Valid    bool    `json:"certificate_valid"`
	EAX      *uint32 `json:"actual_EAX_u32"`
	Services []struct {
		Site   uint32              `json:"site"`
		Args   []uint32            `json:"arguments_u32"`
		Packet seExtraPacket40AA70 `json:"packet"`
	} `json:"services"`
	Events []struct {
		Kind     string          `json:"kind"`
		PC       uint32          `json:"PC"`
		Width    int             `json:"width"`
		Value    json.RawMessage `json:"value"`
		Complete bool            `json:"completed"`
	} `json:"events"`
}

// Both data-only loaders stream one row at a time and retain only five input
// or ten matching reference rows. References never drive the native observer.
// The reference loader is first called after an actual native entry.
func seExtraLoad40AA70(reference bool) (out map[string]json.RawMessage, err error) {
	path, digest, count := "testdata/setting_eligibility_40aa70/inputs-v1.json.gz", seExtraInputsSHA40AA70, 662
	key := "cases"
	if reference {
		path, digest, count, key = "testdata/setting_eligibility_40aa70/reference-v1.json.gz", seExtraReferenceSHA40AA70, 1324, "rows"
	}
	f, z, err := rskOpenGzip4E23C0(path, digest)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := z.Close(); err == nil {
			err = e
		}
		if e := f.Close(); err == nil {
			err = e
		}
	}()
	d := json.NewDecoder(z)
	tok, err := d.Token()
	if err != nil || tok != json.Delim('{') {
		return nil, fmt.Errorf("input object: %v", err)
	}
	wanted := map[string]bool{}
	for _, name := range seExtraNames40AA70 {
		wanted[name] = true
	}
	out = map[string]json.RawMessage{}
	seen := map[string]bool{}
	rows := 0
	collection := false
	for d.More() {
		tok, e := d.Token()
		if e != nil {
			return nil, e
		}
		name, ok := tok.(string)
		if !ok || seen[name] {
			return nil, fmt.Errorf("unique data key")
		}
		seen[name] = true
		if name == key {
			collection = true
			tok, e = d.Token()
			if e != nil || tok != json.Delim('[') {
				return nil, fmt.Errorf("row array")
			}
			for d.More() {
				var raw json.RawMessage
				if e = d.Decode(&raw); e != nil {
					return nil, e
				}
				rows++
				var identity struct {
					Name string `json:"name"`
					CW   uint16 `json:"control_word"`
				}
				if e = json.Unmarshal(raw, &identity); e != nil {
					return nil, e
				}
				if wanted[identity.Name] {
					id := identity.Name
					if reference {
						id = fmt.Sprintf("%s/%04x", id, identity.CW)
					}
					if _, exists := out[id]; exists {
						return nil, fmt.Errorf("duplicate selected data")
					}
					out[id] = append(json.RawMessage(nil), raw...)
				}
			}
			tok, e = d.Token()
			if e != nil || tok != json.Delim(']') {
				return nil, fmt.Errorf("row array end")
			}
		} else {
			var raw json.RawMessage
			if e = d.Decode(&raw); e != nil {
				return nil, e
			}
			if name == "schema" {
				var s string
				if e = json.Unmarshal(raw, &s); e != nil {
					return nil, e
				}
				want := "opennox-aa70-independent-inputs-v1"
				if reference {
					want = "opennox-aa70-original346-boundary-reference-v1"
				}
				if s != want {
					return nil, fmt.Errorf("frozen document schema")
				}
			}
			if name == "fatal" && !bytes.Equal(raw, []byte("null")) {
				return nil, fmt.Errorf("original fatal remains failure")
			}
		}
	}
	tok, err = d.Token()
	if err != nil || tok != json.Delim('}') {
		return nil, fmt.Errorf("data object end")
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return nil, fmt.Errorf("whole gzip EOF/CRC: %v", e)
	}
	want := 5
	if reference {
		want = 10
	}
	if !collection || rows != count || len(out) != want || !seen["schema"] {
		return nil, fmt.Errorf("whole selected data count")
	}
	return out, nil
}

var seExtraInputOnce40AA70 sync.Once
var seExtraInputRows40AA70 map[string]json.RawMessage
var seExtraInputError40AA70 error
var seExtraReferenceOnce40AA70 sync.Once
var seExtraReferenceRows40AA70 map[string]json.RawMessage
var seExtraReferenceError40AA70 error

func seExtraInputCase40AA70(t *testing.T, name string) seExtraInput40AA70 {
	t.Helper()
	seExtraInputOnce40AA70.Do(func() { seExtraInputRows40AA70, seExtraInputError40AA70 = seExtraLoad40AA70(false) })
	if seExtraInputError40AA70 != nil {
		t.Fatal(seExtraInputError40AA70)
	}
	raw, ok := seExtraInputRows40AA70[name]
	if !ok {
		t.Fatal("frozen independent selected input")
	}
	var in seExtraInput40AA70
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	in.RawSHA = rskHash4E23C0(raw)
	if len(in.Args) != 1 || len(in.AfterReads) != 0 || len(in.AfterServices) != 0 {
		t.Fatal("extra scope has no scheduled mutation")
	}
	for _, service := range in.Services {
		for _, packet := range service.Responses {
			if len(packet.Writes) != 0 {
				t.Fatal("extra scope excludes packet writes")
			}
		}
	}
	return in
}

func seExtraRefsAfterCall40AA70(t *testing.T, in *seExtraInput40AA70, got int32) [2]seExtraReference40AA70 {
	t.Helper()
	seExtraReferenceOnce40AA70.Do(func() { seExtraReferenceRows40AA70, seExtraReferenceError40AA70 = seExtraLoad40AA70(true) })
	if seExtraReferenceError40AA70 != nil {
		t.Fatal(seExtraReferenceError40AA70)
	}
	var result [2]seExtraReference40AA70
	for i, cw := range []uint16{0x0e3f, 0x023f} {
		raw := seExtraReferenceRows40AA70[fmt.Sprintf("%s/%04x", in.Name, cw)]
		if err := json.Unmarshal(raw, &result[i]); err != nil {
			t.Fatal(err)
		}
		r := &result[i]
		if !r.Valid || r.Status != "returned" || r.Name != in.Name || r.CW != cw || r.InputSHA != in.RawSHA || r.EAX == nil || uint32(got) != *r.EAX {
			t.Fatalf("actual raw result vs authentic selected CW%04x", cw)
		}
	}
	return result
}

func seExtraBytes40AA70(t *testing.T, in *seExtraInput40AA70, token uint32, offset, width int) []byte {
	t.Helper()
	a := uint64(token) + uint64(offset)
	for _, r := range in.Regions {
		if uint64(r.Base) <= a && a+uint64(width) <= uint64(r.Base)+uint64(r.Accessible) {
			raw, err := hex.DecodeString(r.Initial)
			if err != nil || len(raw) != r.Accessible {
				t.Fatal("independent region bytes")
			}
			return append([]byte(nil), raw[a-uint64(r.Base):a-uint64(r.Base)+uint64(width)]...)
		}
	}
	t.Fatal("selected independent field wholly accessible")
	return nil
}

func seExtraInitializePlayer40AA70(t *testing.T, p *server.Player, in *seExtraInput40AA70, token uint32) {
	t.Helper()
	p.ArmorEquip, p.WeaponEquip, p.NetCodeVal = 0xa5a50001, 0x5a5a0002, 0x87654321
	// Only the argument role has a status field in the original narrow input.
	// Other complete Go Player allocations retain an unrelated scalar canary.
	p.Field3680 = 0xa5b60100
	if token == in.Args[0] {
		p.Field3680 = binary.LittleEndian.Uint32(seExtraBytes40AA70(t, in, token, 3680, 4))
	}
	p.Field2068 = binary.LittleEndian.Uint32(seExtraBytes40AA70(t, in, token, 2068, 4))
	p.Field2140 = binary.LittleEndian.Uint32(seExtraBytes40AA70(t, in, token, 2140, 4))
}

func seExtraCall40AA70(t *testing.T, backend string, player unsafe.Pointer) int32 {
	t.Helper()
	if backend == "Go" {
		return sub40AA70Native(player)
	}
	if backend == "C" {
		return settingEligibilityFixtureCall40AA70(player)
	}
	t.Fatal("declared backend")
	return 0
}

func seExtraPointer40AA70(t *testing.T, ptr unsafe.Pointer) {
	t.Helper()
	w := twoSettingsPointerFixturePointerWitness40A740(ptr)
	if !w.Above4GiB || !w.Low32LosesBits || w.Native != uintptr(ptr) {
		t.Fatal("actual C full-pointer witness; numeric low32 loss only")
	}
}

type seExtraServer40AA70 struct {
	Server
	native *server.Server
}

func (s *seExtraServer40AA70) S() *server.Server { return s.native }

func seExtraWorld40AA70(t *testing.T, in *seExtraInput40AA70, flags uint32) {
	t.Helper()
	if settingEligibilityNativeObserverCurrent40AA70() != nil {
		t.Fatal("actual nil observer control")
	}
	previous, oldGame, oldGameplay := GetServer, noxflags.GetGame(), noxflags.GetGamePlay()
	srv := server.NewHigh(nil, nil, strman.New())
	GetServer = func() Server { return &seExtraServer40AA70{native: srv} }
	t.Cleanup(func() {
		srv.Close()
		GetServer = previous
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGamePlay(oldGameplay)
	})
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameFlag(flags))
	noxflags.UnsetGamePlay(noxflags.GetGamePlay())
	cell := settingEligibilityNativeCell40AA70()
	seExtraPointer40AA70(t, cell)
	old := *(*uint32)(cell)
	*(*uint32)(cell) = in.InitialWords["005d5348"]
	t.Cleanup(func() { *(*uint32)(cell) = old })
	info := settingEligibilityNativeServiceDefault40AA70(0x40aa72, nil).Pointer
	seExtraPointer40AA70(t, info)
	window := unsafe.Slice((*byte)(unsafe.Add(info, 48)), 8)
	saved := append([]byte(nil), window...)
	t.Cleanup(func() { copy(window, saved) })
	window[4], window[5] = seExtraBytes40AA70(t, in, in.Roles["info0"], 52, 1)[0], 0
	limit := settingEligibilityFixtureLimitCell40AA70()
	seExtraPointer40AA70(t, limit)
	oldLimit := *(*uint32)(limit)
	*(*uint32)(limit) = 0x80000000
	t.Cleanup(func() { *(*uint32)(limit) = oldLimit })
	for _, site := range []uint32{0x40ab34, 0x40ab56} {
		if v := settingEligibilityNativeServiceDefault40AA70(site, nil).Word; v > 255 {
			t.Fatal("actual C AL default zero extension")
		}
	}
	if settingEligibilityNativeServiceDefault40AA70(0x40ab63, nil).Word != 0x80000000 {
		t.Fatal("actual separate C signed-limit DWORD bit preservation")
	}
	before := append([]byte(nil), window...)
	global := *(*uint32)(cell)
	t.Cleanup(func() {
		if !bytes.Equal(window, before) || *(*uint32)(cell) != global {
			t.Error("actual default read-only info/global window changed")
		}
	})
}

func TestSettingEligibilityDefaultLinked40AA70(t *testing.T) {
	controls := []struct {
		name  string
		index int
		flags uint32
	}{{"nil-argument", 0, 0}, {"status-global", 1, 0}, {"count770", 2, 4096}, {"actual-char-AL", 3, 32768}}
	for _, backend := range []string{"Go", "C"} {
		for _, control := range controls {
			t.Run(backend+"/"+control.name, func(t *testing.T) {
				in := seExtraInputCase40AA70(t, seExtraNames40AA70[control.index])
				seExtraWorld40AA70(t, &in, control.flags)
				var arg unsafe.Pointer
				if in.Args[0] != 0 {
					p, free := alloc.NewHigh(server.Player{})
					t.Cleanup(free)
					seExtraInitializePlayer40AA70(t, p, &in, in.Args[0])
					arg = unsafe.Pointer(p)
					seExtraPointer40AA70(t, arg)
				}
				got := seExtraCall40AA70(t, backend, arg)
				seExtraRefsAfterCall40AA70(t, &in, got)
				if settingEligibilityNativeObserverCurrent40AA70() != nil {
					t.Fatal("observer remains nil")
				}
				t.Logf("SETTING_ELIGIBILITY_DEFAULT_LINKED_OK backend=%s control=%s observer=nil original-return-CW-comparisons=2 actual-linked-services=true actual-production-global-cell=true actual-C-AL-zeroextend-probes=2 actual-C-limit-DWORD-probe=1 source-bound-not-raw-default-trace=true helper-original-body-credit=0", backend, control.name)
			})
		}
	}
}

type seExtraObservedService40AA70 struct {
	Site  uint32
	Args  []settingEligibilityNativeArg40AA70
	Reply settingEligibilityNativeReply40AA70
}

func TestSettingEligibilityPinnedGoHeap40AA70(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			if settingEligibilityNativeObserverCurrent40AA70() != nil {
				t.Fatal("serial observer stack starts empty")
			}
			in := seExtraInputCase40AA70(t, seExtraNames40AA70[4])
			if in.Args[0] != in.Roles["iterator0"] {
				t.Fatal("frozen argument/iterator0 alias")
			}
			var pins runtime.Pinner
			defer pins.Unpin()
			players := [3]*server.Player{new(server.Player), new(server.Player), new(server.Player)}
			tokens := [3]uint32{in.Roles["iterator0"], in.Roles["iterator1"], in.Roles["iterator2"]}
			identities := map[uint32]unsafe.Pointer{}
			before := [3][]byte{}
			for i, p := range players {
				seExtraInitializePlayer40AA70(t, p, &in, tokens[i])
				pins.Pin(p)
				ptr := unsafe.Pointer(p)
				seExtraPointer40AA70(t, ptr)
				identities[tokens[i]] = ptr
				before[i] = append([]byte(nil), unsafe.Slice((*byte)(ptr), int(unsafe.Sizeof(*p)))...)
			}
			for _, role := range []string{"info0", "info1"} {
				p, free := alloc.MallocHigh(64)
				t.Cleanup(free)
				raw := unsafe.Slice((*byte)(p), 64)
				for i := range raw {
					raw[i] = 0xa7
				}
				copy(raw, seExtraBytes40AA70(t, &in, in.Roles[role], 0, 53))
				identities[in.Roles[role]] = p
				saved := append([]byte(nil), raw...)
				t.Cleanup(func() {
					if !bytes.Equal(raw, saved) {
						t.Error("whole foreign info buffer/canary changed")
					}
				})
			}
			pointer := func(token uint32) unsafe.Pointer {
				if token == 0 {
					return nil
				}
				p, ok := identities[token]
				if !ok {
					t.Fatalf("observed undeclared native pointer token %08x", token)
				}
				return p
			}
			cell := settingEligibilityNativeCell40AA70()
			seExtraPointer40AA70(t, cell)
			oldGlobal := *(*uint32)(cell)
			*(*uint32)(cell) = in.InitialWords["005d5348"]
			defer func() { *(*uint32)(cell) = oldGlobal }()
			layout := settingEligibilityFixtureCLayout40AA70()
			named := [3]uintptr{unsafe.Offsetof(server.Player{}.Field3680), unsafe.Offsetof(server.Player{}.Field2068), unsafe.Offsetof(server.Player{}.Field2140)}
			for i, f := range layout {
				if f.Offset != named[i] || f.Width != 4 {
					t.Fatal("actual C/Go named DWORD layout")
				}
			}
			seenGC := map[int]bool{}
			gcs := 0
			var reads []settingEligibilityNativeOperand40AA70
			var services []seExtraObservedService40AA70
			counts := map[uint32]int{}
			observer := &settingEligibilityNativeObserver40AA70{}
			observer.Operand = func(op settingEligibilityNativeOperand40AA70) {
				if op.Field >= 1 && op.Field <= 3 {
					if op.Offset != named[op.Field-1] || op.Width != 4 || !op.AddressValid || op.Address != unsafe.Add(op.Base, op.Offset) {
						t.Fatal("actual named whole DWORD read")
					}
					if !op.Complete && !seenGC[op.Field] {
						seenGC[op.Field] = true
						for i := 0; i < 3; i++ {
							runtime.GC()
							gcs++
						}
						runtime.KeepAlive(players)
					}
				}
				if op.Complete {
					reads = append(reads, op)
				}
			}
			observer.Service = func(site uint32, args []settingEligibilityNativeArg40AA70) (settingEligibilityNativeReply40AA70, bool) {
				s, ok := in.Services[fmt.Sprintf("%08x", site)]
				if !ok || len(args) != s.Argc || !s.Repeat || len(s.Responses) == 0 {
					t.Fatal("actual site/arity independent input schedule")
				}
				n := counts[site]
				counts[site] = n + 1
				if n >= len(s.Responses) {
					n = len(s.Responses) - 1
				}
				packet := s.Responses[n]
				reply := settingEligibilityNativeReply40AA70{Word: packet.Registers["eax"], ECX: packet.Registers["ecx"], EDX: packet.Registers["edx"]}
				if site == 0x40aa72 || site == 0x40aaed || site == 0x40ab94 || site == 0x40aba8 {
					reply.Pointer = pointer(reply.Word)
				}
				services = append(services, seExtraObservedService40AA70{site, append([]settingEligibilityNativeArg40AA70(nil), args...), reply})
				return reply, true
			}
			pop := pushSettingEligibilityNativeObserver40AA70(observer)
			got := func() int32 { defer pop(); return seExtraCall40AA70(t, backend, pointer(in.Args[0])) }()
			if settingEligibilityNativeObserverCurrent40AA70() != nil {
				t.Fatal("serial LIFO observer stack restored")
			}
			refs := seExtraRefsAfterCall40AA70(t, &in, got)
			for _, r := range refs {
				if len(r.Services) != len(services) {
					t.Fatal("actual service count vs authentic original")
				}
				for i, s := range r.Services {
					actual := services[i]
					if actual.Site != s.Site || len(actual.Args) != len(s.Args) || actual.Reply.Word != s.Packet.Registers["eax"] || actual.Reply.ECX != s.Packet.Registers["ecx"] || actual.Reply.EDX != s.Packet.Registers["edx"] {
						t.Fatal("ordered actual packet/cached service")
					}
					if s.Site == 0x40aa72 || s.Site == 0x40aaed || s.Site == 0x40ab94 || s.Site == 0x40aba8 {
						if actual.Reply.Pointer != pointer(s.Packet.Registers["eax"]) {
							t.Fatal("actual full native pointer reply identity")
						}
					}
					for j, word := range s.Args {
						if s.Site == 0x40aba8 {
							if !actual.Args[j].IsPointer || actual.Args[j].Pointer != pointer(word) {
								t.Fatal("actual full cached player argument")
							}
						} else if actual.Args[j].IsPointer || actual.Args[j].Word != word {
							t.Fatal("actual raw DWORD service argument")
						}
					}
				}
				var expected []struct {
					pc    uint32
					width int
					word  uint32
				}
				for _, e := range r.Events {
					if !e.Complete || e.Kind != "read" {
						continue
					}
					if e.PC != 0x40aaa8 && e.PC != 0x40aab3 && e.PC != 0x40aadf && e.PC != 0x40aafd && e.PC != 0x40ab9d {
						continue
					}
					var text string
					if err := json.Unmarshal(e.Value, &text); err != nil {
						t.Fatal(err)
					}
					raw, err := hex.DecodeString(text)
					if err != nil || len(raw) != e.Width {
						t.Fatal("original actual read width")
					}
					var word uint32
					if len(raw) == 1 {
						word = uint32(raw[0])
					} else {
						word = binary.LittleEndian.Uint32(raw)
					}
					expected = append(expected, struct {
						pc    uint32
						width int
						word  uint32
					}{e.PC, e.Width, word})
				}
				if len(reads) != len(expected) {
					t.Fatal("every actual selected read")
				}
				for i, want := range expected {
					if reads[i].Site != want.pc || int(reads[i].Width) != want.width || !reads[i].ValueValid || reads[i].ValueWord != want.word {
						t.Fatal("ordered actual BYTE/DWORD read vs original")
					}
				}
			}
			if gcs != 9 || len(seenGC) != 3 {
				t.Fatal("three named fields with three actual in-call GCs each")
			}
			for i, p := range players {
				if !bytes.Equal(before[i], unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))) {
					t.Fatal("whole pinned typed player allocation changed")
				}
			}
			if *(*uint32)(cell) != in.InitialWords["005d5348"] {
				t.Fatal("actual read-only production global changed")
			}
			runtime.KeepAlive(players)
			runtime.KeepAlive(identities)
			t.Logf("SETTING_ELIGIBILITY_PINNED_GO_OK backend=%s individually-pinned-players=3 actual-C-pointer-witnesses=3 named-DWORD-fields=3 in-call-GC=9 whole-typed-allocations=3 original-return-and-observations-CW-comparisons=2 helper-original-body-credit=0 physical-register-stack-FPU-credit=0", backend)
		})
	}
}
