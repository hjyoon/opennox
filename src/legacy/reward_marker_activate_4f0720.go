package legacy

import "github.com/opennox/opennox/v1/server"

func rewardMarkerActivateCall4F0720(
	s *server.Server,
	marker *server.Object,
	stage uint32,
	runtime server.RewardMarkerActivateRuntime4F0720,
) *server.Object {
	return s.RewardMarkerActivate4F0720(marker, stage, runtime)
}
