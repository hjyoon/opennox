package legacy

import "github.com/opennox/opennox/v1/server"

func sparkUpdateCall53ADC0(obj *server.Object) {
	outer := GetServer()
	outer.S().SparkUpdate53ADC0(obj, server.SparkUpdateRuntime53ADC0{
		DelayedDelete: outer.DelayedDelete,
	})
}
