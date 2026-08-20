package server

import (
	"fmt"
	"reflect"
	"testing"
)

var _ func(uint32) = SageNoop4EF4F0

func TestSageNoop4EF4F0IgnoresWholeDword(t *testing.T) {
	for _, value := range []uint32{
		0,
		1,
		0x7fffffff,
		0x80000000,
		0xffffffff,
	} {
		SageNoop4EF4F0(value)
	}
}

type sageCommandWorld4EF4F0 struct {
	quest  bool
	events []string
	fault  int
}

func (w *sageCommandWorld4EF4F0) record(event string) {
	w.events = append(w.events, event)
	if w.fault != 0 && len(w.events) == w.fault {
		panic(event)
	}
}

func (w *sageCommandWorld4EF4F0) runtime() SageCommandRuntime4EF4F0 {
	return SageCommandRuntime4EF4F0{
		QuestMode: func() bool {
			w.record(fmt.Sprintf("quest:%t", w.quest))
			return w.quest
		},
		SetSage: func(value uint32) {
			w.record(fmt.Sprintf("set:%08x", value))
			SageNoop4EF4F0(value)
		},
		LoadString: func(key string) string {
			w.record("load:" + key)
			return "message:" + key
		},
		Print: func(message string) {
			w.record("print:" + message)
		},
	}
}

func TestSageCommand4EF4F0SetOrder(t *testing.T) {
	w := &sageCommandWorld4EF4F0{}
	SageCommand4EF4F0(true, w.runtime())
	want := []string{
		"quest:false",
		"set:00000001",
		"load:sageset",
		"print:message:sageset",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestSageCommand4EF4F0SetQuestIsQuiet(t *testing.T) {
	w := &sageCommandWorld4EF4F0{quest: true}
	SageCommand4EF4F0(true, w.runtime())
	want := []string{"quest:true"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestSageCommand4EF4F0UnsetSkipsQuest(t *testing.T) {
	w := &sageCommandWorld4EF4F0{quest: true}
	runtime := w.runtime()
	runtime.QuestMode = func() bool {
		t.Fatal("unset queried Quest mode")
		return true
	}
	SageCommand4EF4F0(false, runtime)
	want := []string{
		"set:00000000",
		"load:sageunset",
		"print:message:sageunset",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestSageCommand4EF4F0FaultPrefixes(t *testing.T) {
	tests := []struct {
		name   string
		enable bool
		want   []string
	}{
		{
			name:   "set",
			enable: true,
			want: []string{
				"quest:false",
				"set:00000001",
				"load:sageset",
				"print:message:sageset",
			},
		},
		{
			name:   "unset",
			enable: false,
			want: []string{
				"set:00000000",
				"load:sageunset",
				"print:message:sageunset",
			},
		},
	}

	for _, test := range tests {
		for fault := 1; fault <= len(test.want); fault++ {
			t.Run(fmt.Sprintf("%s/event-%d", test.name, fault), func(t *testing.T) {
				w := &sageCommandWorld4EF4F0{fault: fault}
				defer func() {
					if got := recover(); got != test.want[fault-1] {
						t.Fatalf("panic = %v, want %q", got, test.want[fault-1])
					}
					if want := test.want[:fault]; !reflect.DeepEqual(w.events, want) {
						t.Fatalf("events = %v, want %v", w.events, want)
					}
				}()
				SageCommand4EF4F0(test.enable, w.runtime())
			})
		}
	}
}
