//go:build server

package opennox

import (
	"fmt"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
)

func e2eObserveDoorAudioPacket(netmsg.Op, sound.ID, int, bool)                 {}
func e2eObserveDoorAudioVoiceReuseLocked(*nativeAudioEffectsState, ail.Sample) {}

func (sc *e2eScenario) CheckDoorAudio(name string) {
	sc.add(0, name, func() { e2eError(fmt.Errorf("door playback audit requires a client build")) })
}
