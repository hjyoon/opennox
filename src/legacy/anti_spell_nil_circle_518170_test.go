package legacy

import "testing"

func TestAntiSpellCircleNilCallbackOriginal518170(t *testing.T) {
	old := GetServer
	GetServer = func() Server {
		t.Fatal("original nil callback must return before reading the origin or world")
		return nil
	}
	defer func() { GetServer = old }()
	antiSpellNilCircleCProbe518170()
}
