package server

import (
	"testing"
	"unsafe"
)

func TestTriggerUpdateDataLayout(t *testing.T) {
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"size", unsafe.Sizeof(TriggerUpdateData{}), 60},
		{"Field4", unsafe.Offsetof(TriggerUpdateData{}.Field4), 4},
		{"State", unsafe.Offsetof(TriggerUpdateData{}.State), 8},
		{"ScriptCollide", unsafe.Offsetof(TriggerUpdateData{}.ScriptCollide), 12},
		{"ScriptActivate", unsafe.Offsetof(TriggerUpdateData{}.ScriptActivate), 20},
		{"ScriptDeactivate", unsafe.Offsetof(TriggerUpdateData{}.ScriptDeactivate), 28},
		{"SoundActivate", unsafe.Offsetof(TriggerUpdateData{}.SoundActivate), 36},
		{"SoundDeactivate", unsafe.Offsetof(TriggerUpdateData{}.SoundDeactivate), 40},
		{"ClassInclude", unsafe.Offsetof(TriggerUpdateData{}.ClassInclude), 44},
		{"ClassExclude", unsafe.Offsetof(TriggerUpdateData{}.ClassExclude), 48},
		{"TeamInclude", unsafe.Offsetof(TriggerUpdateData{}.TeamInclude), 52},
		{"TeamExclude", unsafe.Offsetof(TriggerUpdateData{}.TeamExclude), 53},
		{"Colors", unsafe.Offsetof(TriggerUpdateData{}.Colors), 54},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}
