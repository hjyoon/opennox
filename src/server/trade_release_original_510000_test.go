package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"
)

//go:embed testdata/trade_release_original_v2.json.gz
var tradeReleaseStock510000 []byte

type originalTradeReleaseEvent struct {
	Kind   string
	Record *string
	Head   *string
}

type originalTradeReleaseCase struct {
	Input struct {
		Name      string
		Entry     string
		Placement string
		Stock     int
		Offers    [2]int
		Gold      [2]bool
		Clobber   bool
	}
	Observed struct {
		Events        []originalTradeReleaseEvent
		Head          *string
		NewerNext     *string `json:"newer_next"`
		OlderPrevious *string `json:"older_previous"`
	}
}

func TestTradeReleaseOrderOriginal510000(t *testing.T) {
	compressed, err := gzip.NewReader(bytes.NewReader(tradeReleaseStock510000))
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	raw, err := io.ReadAll(compressed)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "1f58c1245dacdd238a8edcb523c894ab5e0b6c2aaefe1db3954727a2b5256dea" {
		t.Fatalf("fixture digest %s", got)
	}
	var stock struct{ Rows []originalTradeReleaseCase }
	if err := json.Unmarshal(raw, &stock); err != nil {
		t.Fatal(err)
	}
	if len(stock.Rows) != 72 {
		t.Fatalf("fixture cases %d", len(stock.Rows))
	}
	var releases int
	for _, row := range stock.Rows {
		if row.Input.Entry == "release_trade" {
			t.Run(row.Input.Name, func(t *testing.T) { checkTradeReleaseOrder510000(t, row) })
			releases++
		}
	}
	if releases != 60 {
		t.Fatalf("release cases %d", releases)
	}
}

func checkTradeReleaseOrder510000(t *testing.T, row originalTradeReleaseCase) {
	in, want := row.Input, row.Observed
	target, newer, older, other := &TradeSession{}, &TradeSession{}, &TradeSession{}, &TradeSession{}
	head := map[string]*TradeSession{"only": target, "head": target, "middle": newer, "tail": newer, "detached": other}[in.Placement]
	label := func(name string) *string { return &name }
	labels := map[*TradeSession]*string{
		target: label("target"), newer: label("newer"), older: label("older"), other: label("other"),
	}
	if in.Placement == "middle" || in.Placement == "tail" {
		target.Field60, newer.Field56 = newer, target
	}
	if in.Placement == "middle" || in.Placement == "head" {
		target.Field56, older.Field60 = older, target
	}

	allocation := &nativeTradeSessionAllocation{items: make(map[*TradeItem]nativeTradeItemAllocation)}
	state := serverTradeNativeState{head: head, sessions: map[*TradeSession]*nativeTradeSessionAllocation{target: allocation}}
	var events []originalTradeReleaseEvent
	record := func(kind string, role *string) {
		events = append(events, originalTradeReleaseEvent{Kind: kind, Record: role, Head: labels[state.head]})
	}
	makeList := func(count int, name string, freeObject bool) *TradeItem {
		var head *TradeItem
		for i := count - 1; i >= 0; i-- {
			node := &TradeItem{Item0: &Object{}, Field8: head}
			if head != nil {
				head.Field12 = node
			}
			nodeLabel := label(fmt.Sprintf("%s_node%d", name, i))
			item := nativeTradeItemAllocation{freeNode: func() {
				record("free_item_node", nodeLabel)
				if in.Clobber {
					*node = TradeItem{}
				}
			}}
			if freeObject {
				objectLabel := label(fmt.Sprintf("%s_object%d", name, i))
				item.freeObject = func() { record("free_object", objectLabel) }
			}
			allocation.items[node] = item
			head = node
		}
		return head
	}
	target.Field20 = makeList(in.Stock, "stock", true)
	target.Field32 = makeList(in.Offers[0], "offer0", false)
	target.Field36 = makeList(in.Offers[1], "offer1", false)
	for i, present := range in.Gold {
		var name *string
		if present {
			name = label(fmt.Sprintf("gold%d", i))
			if i == 0 {
				target.Field48 = &Object{}
			} else {
				target.Field52 = &Object{}
			}
		}
		allocation.freeGold[i] = func() { record("free_object", name) }
	}
	allocation.freeSession = func() { record("free_session", labels[target]) }
	if !state.release(target) {
		t.Fatal("registered session did not release")
	}
	if !reflect.DeepEqual(events, want.Events) {
		actual, _ := json.Marshal(events)
		expected, _ := json.Marshal(want.Events)
		t.Fatalf("release events %s, want %s", actual, expected)
	}
	if !reflect.DeepEqual(labels[state.head], want.Head) {
		t.Fatal("incorrect session-list head")
	}
	for _, links := range []struct {
		got, want *string
	}{{labels[newer.Field56], want.NewerNext}, {labels[older.Field60], want.OlderPrevious}} {
		if links.want != nil && *links.want == "unused" {
			continue
		}
		if !reflect.DeepEqual(links.got, links.want) {
			t.Fatal("incorrect neighbor link after release")
		}
	}
	if len(allocation.items) != 0 || state.release(target) {
		t.Fatal("released allocations remained registered")
	}
}
