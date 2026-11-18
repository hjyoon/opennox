package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type monsterSpellQueryServer540CE0 struct {
	Server
	server *server.Server
}

func (s *monsterSpellQueryServer540CE0) S() *server.Server { return s.server }

func TestMonsterSpellQueriesOriginalNativeC540CE0(t *testing.T) {
	data, err := os.ReadFile("testdata/monster_spell_query_original_v1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	const expectedHash = "dac1630223ef169919aa1fcb267ebe41b4a3feea482e70b863cc207dd9b472c9"
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != expectedHash {
		t.Fatalf("original spell query corpus hash = %s", got)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var corpus struct {
		Cases   int
		Startup string `json:"startup_control_word"`
		Guards  bool   `json:"unchanged_scratch_and_register_guards"`
		Entered map[string]int
		Rows    []struct {
			Input struct {
				Entry   string
				Spell   int32
				Buff    uint32
				Owner   int
				Records [][2]int32
			}
			Result int
		}
	}
	if err := json.NewDecoder(reader).Decode(&corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Cases != 305 || len(corpus.Rows) != corpus.Cases || corpus.Startup != "0xe3f" || !corpus.Guards ||
		corpus.Entered["spell-enchant"] != 132 || corpus.Entered["summon-id"] != 194 || corpus.Entered["active-summon"] != 30 ||
		corpus.Entered["enchant-table"] != 2316 || corpus.Entered["buff-test"] != 116 ||
		corpus.Entered["duration-first"] != 30 || corpus.Entered["duration-next"] != 39 {
		t.Fatal("unexpected original spell query proof")
	}

	a := game4ConsumerValue(t, server.Object{})
	b := game4ConsumerValue(t, server.Object{})
	objects := []*server.Object{nil, a, b}
	s := &monsterSpellQueryServer540CE0{server: new(server.Server)}
	durations := &s.server.Spells.Dur
	oldServer := GetServer
	GetServer = func() Server { return s }
	t.Cleanup(func() { GetServer = oldServer })

	for i, row := range corpus.Rows {
		in := row.Input
		t.Run(fmt.Sprintf("%s-%03d", in.Entry, i), func(t *testing.T) {
			a.Buffs, b.Buffs = in.Buff, ^in.Buff
			unit := objects[in.Owner]
			var result int
			switch in.Entry {
			case "spell-enchant":
				result = monsterSpellEnchantFixture540CE0(unit, in.Spell)
			case "summon-id":
				result = monsterSpellIDFixture540D20(in.Spell)
			case "active-summon":
				var next *server.DurSpell
				records := make([]*server.DurSpell, len(in.Records))
				values := make([]server.DurSpell, len(records))
				for j := len(records) - 1; j >= 0; j-- {
					rec := in.Records[j]
					values[j] = server.DurSpell{Spell: uint32(rec[0]), Caster16: objects[rec[1]], Next: next}
					records[j] = game4ConsumerValue(t, values[j])
					next = records[j]
				}
				durations.List = next
				defer func() { durations.List = nil }()
				result = monsterSpellOwnerFixture540D40(unit)
				if durations.List != next {
					t.Fatal("duration query changed the head")
				}
				for j, record := range records {
					if *record != values[j] {
						t.Fatalf("duration query modified record %d", j)
					}
				}
			default:
				t.Fatalf("unexpected entry %q", in.Entry)
			}
			if result != row.Result || a.Buffs != in.Buff || b.Buffs != ^in.Buff {
				t.Fatalf("entry %s result/buffs = %d/%#x/%#x, want %d/%#x/%#x",
					in.Entry, result, a.Buffs, b.Buffs, row.Result, in.Buff, ^in.Buff)
			}
		})
	}
}
