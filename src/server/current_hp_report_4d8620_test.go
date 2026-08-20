package server

import (
	"fmt"
	"reflect"
	"testing"
)

type currentHPReportWorld4D8620 struct {
	object         string
	healthByObject map[string]string
	current        map[string]uint16
	netCode        uint32
	recipient      int32
	sendResult     int32
	healthLoads    int
	events         []string
	faultAt        int
	afterEvent     map[string]func()
}

func newCurrentHPReportWorld4D8620() *currentHPReportWorld4D8620 {
	return &currentHPReportWorld4D8620{
		healthByObject: make(map[string]string),
		current:        make(map[string]uint16),
		afterEvent:     make(map[string]func()),
	}
}

func (w *currentHPReportWorld4D8620) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func (w *currentHPReportWorld4D8620) hooks() currentHPReportHooks4D8620[string, string] {
	return currentHPReportHooks4D8620[string, string]{
		loadObjectArg: func() string {
			obj := w.object
			w.record("load-object=" + obj)
			return obj
		},
		loadHealth: func(obj string) string {
			w.healthLoads++
			health := w.healthByObject[obj]
			w.record(fmt.Sprintf("load-health-%d:%s=%s", w.healthLoads, obj, health))
			if obj == "" {
				panic("nil-object-health")
			}
			return health
		},
		getUnitNetCode: func(obj string) uint32 {
			value := w.netCode
			w.record(fmt.Sprintf("net-code:%s=%08x", obj, value))
			return value
		},
		loadCurrent: func(health string) uint16 {
			value := w.current[health]
			w.record(fmt.Sprintf("load-current:%s=%d", health, value))
			if health == "" {
				panic("nil-health-current")
			}
			return value
		},
		loadRecipient: func() int32 {
			value := w.recipient
			w.record(fmt.Sprintf("load-recipient=%d", value))
			return value
		},
		sendReliable: func(recipient int32, packet [4]byte, related string, remove int32) int32 {
			result := w.sendResult
			w.record(fmt.Sprintf(
				"send:%d:%02x%02x%02x%02x:related=%s:remove=%d:result=%d",
				recipient, packet[0], packet[1], packet[2], packet[3], related, remove, result,
			))
			return result
		},
	}
}

func TestCurrentHPReport4D8620ObjectAndInitialHealthOrder(t *testing.T) {
	w := newCurrentHPReportWorld4D8620()
	defer func() {
		if got := recover(); got != "nil-object-health" {
			t.Fatalf("panic = %v, want nil-object-health", got)
		}
		want := []string{"load-object=", "load-health-1:="}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	currentHPReport4D8620(w.hooks())
}

func TestCurrentHPReport4D8620NullInitialHealthReturnsZero(t *testing.T) {
	w := newCurrentHPReportWorld4D8620()
	w.object = "object"
	w.recipient = 9
	w.sendResult = -1
	if got := currentHPReport4D8620(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"load-object=object", "load-health-1:object="}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestCurrentHPReport4D8620ReloadsHealthAndBuildsExactPacket(t *testing.T) {
	w := newCurrentHPReportWorld4D8620()
	w.object = "object"
	w.healthByObject["object"] = "health-1"
	w.current["health-2"] = 0xffff
	w.netCode = 0x12345678
	w.recipient = 7
	w.sendResult = -1 << 31
	w.afterEvent["net-code:object=12345678"] = func() {
		w.healthByObject["object"] = "health-2"
	}
	w.afterEvent["load-current:health-2=65535"] = func() {
		w.current["health-2"] = 0
		w.recipient = 11
	}
	w.afterEvent["load-recipient=11"] = func() {
		w.recipient = 99
	}

	if got := currentHPReport4D8620(w.hooks()); got != -1<<31 {
		t.Fatalf("result = %d, want MinInt32", got)
	}
	want := []string{
		"load-object=object",
		"load-health-1:object=health-1",
		"net-code:object=12345678",
		"load-health-2:object=health-2",
		"load-current:health-2=65535",
		"load-recipient=11",
		"send:11:417856ff:related=:remove=1:result=-2147483648",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestCurrentHPReport4D8620ReloadedNullHealthFaultsBeforeRecipient(t *testing.T) {
	w := newCurrentHPReportWorld4D8620()
	w.object = "object"
	w.healthByObject["object"] = "health"
	w.afterEvent["net-code:object=00000000"] = func() {
		w.healthByObject["object"] = ""
	}
	defer func() {
		if got := recover(); got != "nil-health-current" {
			t.Fatalf("panic = %v, want nil-health-current", got)
		}
		want := []string{
			"load-object=object",
			"load-health-1:object=health",
			"net-code:object=00000000",
			"load-health-2:object=",
			"load-current:=0",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	currentHPReport4D8620(w.hooks())
}

func TestCurrentHPReport4D8620AllFaultPrefixes(t *testing.T) {
	want := []string{
		"load-object=object",
		"load-health-1:object=health",
		"net-code:object=12345678",
		"load-health-2:object=health",
		"load-current:health=4660",
		"load-recipient=7",
		"send:7:4178561a:related=:remove=1:result=-7",
	}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newCurrentHPReportWorld4D8620()
			w.object = "object"
			w.healthByObject["object"] = "health"
			w.current["health"] = 0x1234
			w.netCode = 0x12345678
			w.recipient = 7
			w.sendResult = -7
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want %v", w.events, prefix)
				}
			}()
			currentHPReport4D8620(w.hooks())
		})
	}
}
