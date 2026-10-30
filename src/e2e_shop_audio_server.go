//go:build server

package opennox

import (
	"fmt"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
)

func e2eObserveShopAudioPacket(netmsg.Op, sound.ID, int, int, bool)            {}
func e2eObserveShopAudioVoiceReuseLocked(*nativeAudioEffectsState, ail.Sample) {}

func (sc *e2eScenario) CheckShopAudio(_, _ int, name string) {
	sc.add(0, name, func() { e2eError(fmt.Errorf("shop audio playback audit requires a client build")) })
}
