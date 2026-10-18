package opennox

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opennox/libs/strman"
	"github.com/opennox/noxscript/ns/asm"

	"github.com/opennox/opennox/v1/server"
)

func scriptAPIServerR317(t *testing.T) *Server {
	t.Helper()
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = oldServer })
	return s
}

func scriptAPIFuncsR317(t *testing.T, s *Server) []server.ScriptFunc {
	t.Helper()
	var raw bytes.Buffer
	word := func(v uint32) {
		if err := binary.Write(&raw, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	raw.WriteString("SCRIPT03STRG")
	word(0)
	raw.WriteString("CODE")
	word(3)
	for i, name := range []string{"GLOBAL", "Probe", "Empty"} {
		raw.WriteString("FUNC")
		word(uint32(len(name)))
		raw.WriteString(name)
		args, returns := uint32(0), uint32(0)
		var sizes []uint32
		if i == 0 {
			sizes = []uint32{2}
		} else if i == 1 {
			args, returns, sizes = 2, 1, []uint32{1, 1, 2}
		}
		word(returns)
		word(args)
		raw.WriteString("SYMB")
		word(uint32(len(sizes)))
		word(0)
		for _, size := range sizes {
			word(size)
		}
		raw.WriteString("DATA")
		word(4)
		word(uint32(asm.OpReturn0))
	}
	raw.WriteString("DONE")
	if err := s.noxScript.ReadScript(bytes.NewReader(raw.Bytes())); err != nil {
		t.Fatal(err)
	}
	funcs := s.noxScript.Funcs()
	if len(funcs) != 3 || len(funcs[0].Values) != 2 || len(funcs[1].Values) != 4 || len(funcs[2].Values) != 0 {
		t.Fatal("unexpected compiled-script fixture layout")
	}
	copy(funcs[0].Values, []uint32{11, 22})
	copy(funcs[1].Values, []uint32{101, 202, 303, 0xffffffff})
	// Set diagnostic metadata explicitly so this API lifetime test does not
	// depend on the separate compiled-name qualifier parser repair.
	funcs[1].FuncDef.Name = "Probe%7%13%-29"
	funcs[1].NamePref, funcs[1].PosOff = "%7", image.Pt(13, -29)
	funcs[2].PosOff = image.Pt(-3, 9)
	return funcs
}

func TestAPIScriptSnapshotR317CanceledRegisteredRoute(t *testing.T) {
	s := scriptAPIServerR317(t)
	funcs := scriptAPIFuncsR317(t, s)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/debug/nox/script/legacy", nil).WithContext(ctx))
	if w.Body.Len() != 0 || len(w.Header()) != 0 || funcs[1].Values[0] != 101 || funcs[1].Name() != "Probe%7%13%-29" {
		t.Fatalf("canceled script response accessed/changed live state: headers=%v body=%q", w.Header(), w.Body.String())
	}
}

func TestAPIScriptSnapshotR317OwnsValuesAndMetadataBeforeHTTP(t *testing.T) {
	s := scriptAPIServerR317(t)
	funcs := scriptAPIFuncsR317(t, s)
	const want = `[
	{
		"ind": 0,
		"name": "GLOBAL",
		"locals": [
			11,
			22
		]
	},
	{
		"ind": 1,
		"name": "Probe%7%13%-29",
		"args": [
			101,
			202
		],
		"locals": [
			303,
			4294967295
		],
		"returns": 1,
		"suff": "%7",
		"dpos": {
			"X": 13,
			"Y": -29
		}
	},
	{
		"ind": 2,
		"name": "Empty",
		"dpos": {
			"X": -3,
			"Y": 9
		}
	}
]
`
	retired := false
	w := &retirementHeaderWriterR316{ResponseRecorder: httptest.NewRecorder(), retire: func() {
		retired = true
		for i := range funcs {
			clear(funcs[i].Values)
			funcs[i].FuncDef = asm.FuncDef{Name: "Retired"}
		}
		// Retire only the VM table; the wrapper also resets mapgen globals.
		s.Server.NoxScriptVM.Reset()
	}}
	serveAPIWorldRouteR316(t, s, w, "/debug/nox/script/legacy")
	if !retired || len(s.noxScript.Funcs()) != 0 || w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json" || w.Body.String() != want {
		t.Fatalf("script lifetime or original JSON format changed: retired=%t funcs=%d status=%d headers=%v\ngot=%s\nwant=%s", retired, len(s.noxScript.Funcs()), w.Code, w.Header(), w.Body, want)
	}
}

func TestAPIScriptSnapshotR317EmptyVMFormat(t *testing.T) {
	s := scriptAPIServerR317(t)
	w := runAPIObjectsRequestR316(t, s, "/debug/nox/script/legacy")
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json" || w.Body.String() != "null\n" {
		t.Fatalf("empty script JSON changed: status=%d headers=%v body=%q", w.Code, w.Header(), w.Body.String())
	}
}
