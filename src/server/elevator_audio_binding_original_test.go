package server

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestElevatorAudioBindingOriginal(t *testing.T) {
	flags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(flags) })
	s := newElevatorStateServer(t)
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	for id := 249; id <= 260; id++ {
		s.Audio.bySound[id].Field12 = 1
	}
	var owner *Object
	var got []uint32
	s.Audio.OnSound(func(id sound.ID, kind int, p *Object, pos types.Pointf) {
		if p != owner || kind != 0 || pos != owner.PosVec {
			t.Fatal("audio target/kind/position changed")
		}
		got = append(got, uint32(id))
	})
	count := 0
	for _, row := range elevatorPassengerOriginalRows(t) {
		if row.Mode != "actual-startup" || row.Input.Kind != "audio" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			p, free := alloc.New(Object{})
			t.Cleanup(free)
			owner = p
			*p = Object{Material: row.Input.Material, ObjSubClass: object.SubClass(row.Input.Subclass), PosVec: rigidPointOriginal(row.Input.Source)}
			s.Audio.Reset()
			got = make([]uint32, 0)
			want := make([]uint32, 0)
			for _, e := range row.Events {
				if e.Service == "audio" {
					want = append(want, e.Value)
				}
			}
			s.ElevatorAudio53B490(p, row.Input.Direction != 0)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("native audio IDs=%v original=%v", got, want)
			}
			queued := make([]uint32, 0)
			s.Audio.EachEvent(func(e *AudioEvent) {
				if e.Obj != p || e.Kind != 0 || e.Code != 0 {
					t.Fatal("audio queue target/kind/code changed")
				}
				queued = append(queued, uint32(e.Sound))
			})
			if !reflect.DeepEqual(queued, want) {
				t.Fatalf("native audio queue=%v original=%v", queued, want)
			}
		})
	}
	if count != 100 {
		t.Fatalf("native audio conditions=%d", count)
	}
}
