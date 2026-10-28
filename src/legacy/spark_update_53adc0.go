package legacy

import "github.com/opennox/opennox/v1/server"

func sparkUpdateCall53ADC0(obj *server.Object) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	server.SparkUpdateRaw53ADC0(obj, sparkTrailPushRawRuntime().DelayedDelete)
}
