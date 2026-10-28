package server_test

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/internal/porttest"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectsVelocityOriginal5118A0(t *testing.T) {
	porttest.RunVelocityOriginal5118A0(t, "testdata/objects_velocity_original_5118a0.json", func(s *porttest.VelocityScene5118A0) int32 {
		return server.UpdateObjectsVelocity5118A0(math.Float32frombits(s.Input.Step), s.Runtime())
	})
}
