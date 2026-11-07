//go:build darwin && arm64

package legacy

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unsafe"
)

func rddNativeArgumentKinds4E0B30(name string) ([]string, error) {
	var kinds []string
	switch name {
	case "game_flags", "gameplay_flags", "sound_count":
		kinds = []string{"word"}
	case "balance_table_float":
		kinds = []string{"string", "word"}
	case "monster_sound_set", "fire_protection", "electric_protection", "owner_chain", "monster_hit_latch", "zombie":
		kinds = []string{"object"}
	case "gameball_on_damage":
		kinds = []string{"object", "object", "word"}
	case "late_defend", "pre_damage":
		kinds = []string{"object", "object", "object", "damage", "word"}
	case "adjust_hp", "damage_clear", "player_state", "buff_membership", "buff_power", "buff_off":
		kinds = []string{"object", "word"}
	case "field_guide":
		kinds = []string{"object", "object", "damage"}
	case "audio_event":
		kinds = []string{"word", "object", "word", "word"}
	case "vamp_fx":
		kinds = []string{"word", "coordinates", "word"}
	case "shield":
		kinds = []string{"object", "damage", "word", "object"}
	case "default_damage_sound", "enemy", "target_damage_sound_callback":
		kinds = []string{"object", "object"}
	case "source_damage_callback":
		kinds = []string{"object", "object", "object", "word", "word"}
	default:
		return nil, fmt.Errorf("undeclared original external service %s", name)
	}
	return kinds, nil
}
func rddPrepareNativeArguments4E0B30(name string, args []uint32, mem *rddMemory4E0B30, step *rddNativeStep4E0B30) error {
	kinds, e := rddNativeArgumentKinds4E0B30(name)
	if e != nil {
		return e
	}
	if len(kinds) != len(args) {
		return fmt.Errorf("exact native outgoing argument count for %s", name)
	}
	step.ArgumentKinds = kinds
	step.Strings = map[int]string{}
	step.Coordinates = map[int][4]uint32{}
	step.DamageBefore = map[int]uint32{}
	for i, kind := range kinds {
		switch kind {
		case "string":
			var text []byte
			terminated := false
			for n := 0; n < 512; n++ {
				value, e := mem.read(uint64(args[i])+uint64(n), 1)
				if e != nil {
					return e
				}
				if value[0] == 0 {
					terminated = true
					break
				}
				text = append(text, value[0])
			}
			if !terminated {
				return fmt.Errorf("bounded original service name missing NUL")
			}
			step.Strings[i] = string(text)
		case "coordinates":
			raw, e := mem.read(uint64(args[i]), 16)
			if e != nil {
				return e
			}
			var words [4]uint32
			for n := range words {
				words[n] = binary.LittleEndian.Uint32(raw[n*4:])
			}
			step.Coordinates[i] = words
		case "damage":
			if uint64(args[i]) != rddCaller4E0B30+16 {
				return fmt.Errorf("mutable original damage is not caller damage slot")
			}
			word, e := mem.word(uint64(args[i]))
			if e != nil {
				return e
			}
			step.DamageBefore[i] = word
		}
	}
	return nil
}
func rddDeclaredEffectForHost4E0B30(c *rddNativeComparison4E0B30, event, service rddEvent4E0B30, returned rddFrame4E0B30, counters map[string]int) (rddEffect4E0B30, []uint32, error) {
	var effects []rddEffect4E0B30
	var args []uint32
	key := ""
	switch event.Reason {
	case "after-completed-original-read":
		if event.StimulusIndex < 0 || event.StimulusIndex >= len(c.Input.AfterReads) {
			return rddEffect4E0B30{}, nil, fmt.Errorf("read input stimulus index")
		}
		effects = c.Input.AfterReads[event.StimulusIndex].Effects
		key = fmt.Sprintf("read/%d", event.StimulusIndex)
	case "after-completed-external-service":
		if event.StimulusIndex < 0 || event.StimulusIndex >= len(c.Input.AfterServices) {
			return rddEffect4E0B30{}, nil, fmt.Errorf("service input stimulus index")
		}
		effects = c.Input.AfterServices[event.StimulusIndex].Effects
		args = returned.Args
		key = fmt.Sprintf("after/%d", event.StimulusIndex)
	case "external-service-input-effect":
		site := fmt.Sprintf("%08x", mustRDDNumber4E0B30(event.ServiceSite))
		schedule, ok := c.Input.Services[site]
		if !ok || service.Site != event.ServiceSite || service.ResponseIndex < 0 || service.ResponseIndex >= len(schedule.Responses) {
			return rddEffect4E0B30{}, nil, fmt.Errorf("explicit service effect response identity")
		}
		var response rddResponse4E0B30
		if e := json.Unmarshal(schedule.Responses[service.ResponseIndex], &response); e != nil {
			return rddEffect4E0B30{}, nil, e
		}
		effects = response.Effects
		args = service.Args
		key = fmt.Sprintf("service/%s/%d", site, service.Occurrence)
	default:
		return rddEffect4E0B30{}, nil, fmt.Errorf("unknown native input effect reason %s", event.Reason)
	}
	index := counters[key]
	if index >= len(effects) {
		return rddEffect4E0B30{}, nil, fmt.Errorf("declared effect exhausted")
	}
	effect := effects[index]
	counters[key]++
	address, data, e := c.Projection.resolve(effect, args)
	if e != nil {
		return effect, nil, e
	}
	if address != mustRDDNumber4E0B30(event.Address) || hex.EncodeToString(data) != event.After {
		return effect, nil, fmt.Errorf("original host event differs from independent declared input effect")
	}
	return effect, append([]uint32(nil), args...), nil
}
func (c *rddNativeComparison4E0B30) service(call rawDefaultDamageNativeService4E0B30) (rawDefaultDamageNativeReply4E0B30, bool) {
	if call.Indirect && !rddDeclaredNativeCallback4E0B30(c.Projection.Graph, call.Callback) {
		return rawDefaultDamageNativeReply4E0B30{}, false
	}
	c.record(rddNativeObserved4E0B30{Kind: "service", Indirect: call.Indirect, Site: call.Site, Callback: uintptr(call.Callback), Arguments: rddObserveArguments4E0B30(call.Arguments, c.Damage)})
	step := c.next()
	event := step.Event
	if step.Kind != "service" || call.Site != uint32(mustRDDNumber4E0B30(event.Site)) || len(call.Arguments) != len(step.ArgumentKinds) {
		c.fail("actual external service order/site/argc")
	}
	spec := c.Bundle.Calls[event.Site]
	if spec.Indirect != call.Indirect {
		c.fail("explicit direct/indirect service kind")
	}
	if spec.Indirect {
		if !c.CallbackPending || c.CallbackSite != call.Site || c.CallbackTarget != call.Callback {
			c.fail("indirect service requires exactly one matching callback attempt")
		}
	} else if c.CallbackPending {
		c.fail("pending callback before direct service")
	}
	if spec.Indirect && call.Callback != c.pointer(uint32(mustRDDNumber4E0B30(spec.Target))) {
		c.fail("cached full-width callback identity")
	}
	for i, arg := range call.Arguments {
		kind := step.ArgumentKinds[i]
		if arg.Kind != kind {
			c.fail("native argument%d semantic type %s != %s", i, arg.Kind, kind)
		}
		switch kind {
		case "word":
			if arg.Word != event.Args[i] {
				c.fail("full scalar service argument%d", i)
			}
		case "object":
			if arg.Pointer != c.pointer(event.Args[i]) {
				c.fail("full object service argument%d", i)
			}
		case "string":
			if arg.String != step.Strings[i] {
				c.fail("raw service-name content")
			}
		case "coordinates":
			if arg.Coordinates != step.Coordinates[i] {
				c.fail("all four unnormalized outgoing coordinate DWORDs")
			}
		case "damage":
			if arg.Damage == nil || c.Damage == nil || arg.Damage != c.Damage {
				c.fail("actual caller damage pointer identity before dereference")
			}
			if uint32(*arg.Damage) != step.DamageBefore[i] {
				c.fail("actual caller damage before service")
			}
			if c.Damage != nil && arg.Damage != c.Damage {
				c.fail("mutable caller damage pointer reloaded from another cell")
			}
			c.Damage = arg.Damage
		}
	}
	schedule := c.Input.Services[fmt.Sprintf("%08x", call.Site)]
	if event.ResponseIndex < 0 || event.ResponseIndex >= len(schedule.Responses) {
		c.fail("declared response index")
	}
	var response rddResponse4E0B30
	if e := json.Unmarshal(schedule.Responses[event.ResponseIndex], &response); e != nil {
		c.fail("explicit input response: %v", e)
	}
	if response.EAX == nil {
		c.fail("explicit response word absent")
	}
	reply := rawDefaultDamageNativeReply4E0B30{Word: *response.EAX}
	if spec.Name == "owner_chain" || spec.Name == "monster_sound_set" {
		reply.Pointer = c.pointer(*response.EAX)
	}
	if response.X87 != nil {
		data, e := hex.DecodeString(response.X87.Hex)
		if e != nil || len(data) != map[string]int{"f32": 4, "f64": 8, "ext80": 10}[response.X87.Format] {
			c.fail("exact native external float input packet")
		}
		reply.FloatFormat = response.X87.Format
		copy(reply.FloatPayload[:], data)
	}
	c.CallbackPending = false
	c.CallbackSite = 0
	c.CallbackTarget = nil
	c.Position++
	c.drainEffects()
	return reply, true
}

// Boundary admission depends ONLY on declared native function-token identities.
// Nil/unmapped/non-function inputs return false before any expected-step lookup;
// the implementation must then execute its actual cached C callback pointer.
func rddDeclaredNativeCallback4E0B30(g *ddiGraph4E0B30, target unsafe.Pointer) bool {
	if target == nil {
		return false
	}
	for _, token := range g.Tokens {
		if (token.Kind == 2 || token.Kind == 3) && token.Pointer == target {
			return true
		}
	}
	return false
}
