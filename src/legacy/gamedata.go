package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type monsterBinMode struct {
	solo  bool
	arena bool
}

type monsterBinStore struct {
	head  *server.MonsterDef
	free  map[*server.MonsterDef]func()
	alloc func() (*server.MonsterDef, func())
}

func newMonsterBinStore(allocDef func() (*server.MonsterDef, func())) *monsterBinStore {
	return &monsterBinStore{
		free:  make(map[*server.MonsterDef]func()),
		alloc: allocDef,
	}
}

var loadedMonsterDefs = newMonsterBinStore(func() (*server.MonsterDef, func()) {
	return alloc.New(server.MonsterDef{})
})

func (s *monsterBinStore) clear() {
	for it := s.head; it != nil; {
		next := it.Next244
		if free := s.free[it]; free != nil {
			free()
		}
		it = next
	}
	s.head = nil
	clear(s.free)
}

func (s *monsterBinStore) prepend(def *server.MonsterDef, free func()) {
	def.Next244 = s.head
	s.head = def
	s.free[def] = free
}

func (s *monsterBinStore) findByType(typ uint32) *server.MonsterDef {
	for it := s.head; it != nil; it = it.Next244 {
		if it.TypeInd240 == typ {
			return it
		}
	}
	return nil
}

func (s *monsterBinStore) load(r io.Reader, mode monsterBinMode) error {
	s.clear()
	tr := &monsterTokenReader{r: bufio.NewReader(r)}
	for {
		name, err := tr.next()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			s.clear()
			return err
		}
		if err := s.readDefinition(tr, name, mode); err != nil {
			s.clear()
			return err
		}
	}
}

func (s *monsterBinStore) readDefinition(tr *monsterTokenReader, name string, mode monsterBinMode) (retErr error) {
	def, free := s.alloc()
	if def == nil {
		return errors.New("monster definition allocation failed")
	}
	defer func() {
		if retErr != nil {
			free()
		}
	}()
	if err := copyMonsterName(def.Name0[:], name); err != nil {
		return fmt.Errorf("monster %q: %w", name, err)
	}
	for {
		field, err := tr.next()
		if errors.Is(err, io.EOF) || strings.EqualFold(field, "END") {
			s.prepend(def, free)
			return nil
		} else if err != nil {
			return fmt.Errorf("monster %q: %w", name, err)
		}
		if mode.solo {
			switch {
			case strings.EqualFold(field, "ARENA"):
				if err := tr.skipLine(); err != nil && !errors.Is(err, io.EOF) {
					return fmt.Errorf("monster %q: %w", name, err)
				}
				continue
			case strings.EqualFold(field, "SOLO"):
				continue
			}
		} else if mode.arena {
			switch {
			case strings.EqualFold(field, "SOLO"):
				if err := tr.skipLine(); err != nil && !errors.Is(err, io.EOF) {
					return fmt.Errorf("monster %q: %w", name, err)
				}
				continue
			case strings.EqualFold(field, "ARENA"):
				continue
			}
		}
		value, err := tr.next()
		if err != nil {
			return fmt.Errorf("monster %q field %q: missing value: %w", name, field, err)
		}
		if err := setMonsterDefField(def, field, value); err != nil {
			return fmt.Errorf("monster %q field %q: %w", name, field, err)
		}
	}
}

type monsterTokenReader struct {
	r *bufio.Reader
}

func (r *monsterTokenReader) next() (string, error) {
	var token []byte
	for {
		b, err := r.r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) && len(token) != 0 {
				return string(token), nil
			}
			return "", err
		}
		// Encrypted Nox text files are block-padded with NUL bytes. GAME.EXE's
		// 00517090 reader reaches physical EOF without returning that partial
		// token, so the padding terminates the logical token stream.
		if b == 0 {
			return "", io.EOF
		}
		if b == '/' {
			next, err := r.r.Peek(1)
			if err == nil && next[0] == '/' {
				_, _ = r.r.ReadByte()
				if err := r.skipLine(); err != nil && !errors.Is(err, io.EOF) {
					return "", err
				}
				token = token[:0]
				continue
			}
		}
		if isMonsterSpace(b) {
			if len(token) == 0 {
				continue
			}
			return string(token), nil
		}
		if len(token) == 255 {
			return "", errors.New("monster.bin token exceeds 255 bytes")
		}
		token = append(token, b)
	}
}

func (r *monsterTokenReader) skipLine() error {
	for {
		b, err := r.r.ReadByte()
		if err != nil {
			return err
		}
		if b == '\n' {
			return nil
		}
	}
}

func isMonsterSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

func copyMonsterString(dst []byte, value string) error {
	if strings.EqualFold(value, "NULL") {
		value = ""
	}
	return copyMonsterName(dst, value)
}

func copyMonsterName(dst []byte, value string) error {
	if len(value) >= len(dst) {
		return fmt.Errorf("value exceeds %d bytes", len(dst)-1)
	}
	clear(dst)
	copy(dst, value)
	return nil
}

func parseMonsterUint32(value string) (uint32, error) {
	v, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(int32(v)), nil
}

func parseMonsterFloat32(value string) (float32, error) {
	v, err := strconv.ParseFloat(value, 32)
	return float32(v), err
}

func setMonsterDefField(def *server.MonsterDef, field, value string) error {
	setUint := func(dst *uint32) error {
		v, err := parseMonsterUint32(value)
		if err == nil {
			*dst = v
		}
		return err
	}
	setFloat := func(dst *float32) error {
		v, err := parseMonsterFloat32(value)
		if err == nil {
			*dst = v
		}
		return err
	}
	switch strings.ToUpper(field) {
	case "EXPERIENCE":
		return setUint(&def.Experience64)
	case "HEALTH":
		return setUint(&def.Health68)
	case "QUESTHEALTH":
		return setUint(&def.HealthQuest72)
	case "SPEED":
		return setUint(&def.Speed76)
	case "RETREAT_RATIO":
		return setFloat(&def.RetreatRatio80)
	case "RESUME_RATIO":
		return setFloat(&def.ResumeRatio84)
	case "FLEE_RANGE":
		return setFloat(&def.FleeRange88)
	case "STATUS":
		status, _ := object.ParseMonsterStatusSet(value)
		def.StatusFlags92 = status
		return nil
	case "RUN_MULTIPLIER":
		return setFloat(&def.RunMultiplier96)
	case "MOVE_SOUND_FRAME_A":
		return setUint(&def.MoveSndFrameA100)
	case "MOVE_SOUND_FRAME_B":
		return setUint(&def.MoveSndFrameB104)
	case "MELEE_ATTACK_FRAME":
		return setUint(&def.MeleeAttackFrame108)
	case "MELEE_ATTACK_RANGE":
		return setFloat(&def.MeleeAttackRange112)
	case "MELEE_ATTACK_DAMAGE":
		return setUint(&def.MeleeAttackDamage116)
	case "MELEE_ATTACK_IMPACT":
		return setFloat(&def.MeleeAttackImpact120)
	case "MELEE_ATTACK_DAMAGE_TYPE":
		typ, err := object.ParseDamageType(value)
		if err != nil {
			return err
		}
		def.MeleeAttackDamageType124 = uint32(typ)
		return nil
	case "MELEE_ATTACK_MIN_DELAY":
		return setUint(&def.MeleeAttackDelayMin128)
	case "MELEE_ATTACK_MAX_DELAY":
		return setUint(&def.MeleeAttackDelayMax132)
	case "MELEE_ATTACK_POISON_CHANCE":
		return setUint(&def.MeleeAttackPoisonChange136)
	case "MELEE_ATTACK_POISON_STRENGTH":
		return setUint(&def.MeleeAttackPoisonStrength140)
	case "MELEE_ATTACK_POISON_MAX":
		return setUint(&def.MeleeAttackPoisonMax144)
	case "MISSILE_NAME":
		return copyMonsterString(def.MissileName148[:], value)
	case "MISSILE_ATTACK_RANGE":
		return setFloat(&def.MissileAttackRange212)
	case "MISSILE_ATTACK_FRAME":
		return setUint(&def.MissileAttackFrame216)
	case "MISSILE_ATTACK_MIN_DELAY":
		return setUint(&def.MissileAttackDelayMin220)
	case "MISSILE_ATTACK_MAX_DELAY":
		return setUint(&def.MissileAttackDelayMax224)
	case "MELEE_STRIKE_FUNCTION":
		fn, ok := monsterStrikeFunctions[strings.ToUpper(value)]
		if !ok {
			return fmt.Errorf("unknown strike function %q", value)
		}
		def.MeleeStrikeFunc236 = fn
		return nil
	case "DIE_FUNCTION":
		fn, ok := monsterDieFunctions[strings.ToUpper(value)]
		if !ok {
			return fmt.Errorf("unknown die function %q", value)
		}
		def.DieFunc228 = fn
		return nil
	case "DEAD_FUNCTION":
		fn, ok := monsterDeadFunctions[strings.ToUpper(value)]
		if !ok {
			return fmt.Errorf("unknown dead function %q", value)
		}
		def.DeadFunc232 = fn
		return nil
	default:
		return fmt.Errorf("unknown field")
	}
}

var monsterStrikeFunctions = map[string]unsafe.Pointer{
	"NULL":                 nil,
	"OGRESTRIKE":           unsafe.Pointer(C.nox_xxx_strikeOgre_549220),
	"SCORPIONSTRIKE":       unsafe.Pointer(C.nox_xxx_strikeScorpion_5495B0),
	"VILEZOMBIESTRIKE":     unsafe.Pointer(C.nox_xxx_strikeVileZombie_549700),
	"STONEGOLEMSTRIKE":     unsafe.Pointer(C.nox_xxx_strikeStoneGolem_5497E0),
	"MECHGOLEMSTRIKE":      unsafe.Pointer(C.nox_xxx_strikeMechGolem_549960),
	"WASPSTRIKE":           unsafe.Pointer(C.nox_xxx_strikeWasp_549980),
	"SPIDERSTRIKE":         unsafe.Pointer(C.nox_xxx_strikeSpider_549BC0),
	"SPITTINGSPIDERSTRIKE": unsafe.Pointer(C.nox_xxx_strikeSpittingSpider_549CA0),
	"GHOSTSTRIKE":          unsafe.Pointer(C.nox_xxx_strikeGhost_549A60),
	"BOMBERSTRIKE":         unsafe.Pointer(C.nox_xxx_strikeBomber_549BB0),
	"MONSTERSTRIKE":        unsafe.Pointer(C.nox_xxx_strikeMonsterDefault_549380),
}

var monsterDeadFunctions = map[string]unsafe.Pointer{
	"NULL":             nil,
	"EMBERDEMONDEAD":   unsafe.Pointer(C.sub_549D80),
	"DEMONDEAD":        unsafe.Pointer(C.sub_549E00),
	"IMPDEAD":          unsafe.Pointer(C.sub_549E70),
	"MECHGOLEMDEAD":    unsafe.Pointer(C.sub_549E90),
	"GOLEMDEAD":        unsafe.Pointer(C.sub_549FA0),
	"BOMBERDEAD":       unsafe.Pointer(C.nox_bomberDead_54A150),
	"SPIDERDEAD":       unsafe.Pointer(C.sub_54A250),
	"SKELETONDEAD":     unsafe.Pointer(C.sub_54A310),
	"SKELETONLORDDEAD": unsafe.Pointer(C.sub_54A750),
	"TROLLDEAD":        unsafe.Pointer(C.nox_xxx_monsterDeadTroll_54A270),
}

var monsterDieFunctions = map[string]unsafe.Pointer{
	"NULL":            nil,
	"ARCHERDIE":       unsafe.Pointer(C.sub_54A890),
	"OGREDIE":         unsafe.Pointer(C.sub_54A900),
	"SWORDSMANDIE":    unsafe.Pointer(C.sub_54A7D0),
	"URCHINSHAMANDIE": unsafe.Pointer(C.sub_54A850),
	"OGREWARLORDDIE":  unsafe.Pointer(C.sub_54A950),
}

func Nox_xxx_loadMonsterBin_517010() int {
	loadedMonsterDefs.clear()
	f, err := binfile.BinfileOpen("monster.bin", binfile.ReadOnly)
	if err != nil {
		gameLog.Printf("cannot open monster.bin: %v", err)
		return 0
	}
	defer f.Close()
	if err := f.SetKey(23); err != nil {
		gameLog.Printf("cannot decrypt monster.bin: %v", err)
		return 0
	}
	mode := monsterBinMode{
		solo:  noxflags.HasGame(noxflags.GameModeCoop) || noxflags.HasGame(noxflags.GameFlag22),
		arena: noxflags.HasGame(noxflags.GameOnline),
	}
	if err := loadedMonsterDefs.load(f, mode); err != nil {
		gameLog.Printf("cannot parse monster.bin: %v", err)
		return 0
	}
	return 1
}

func nox_xxx_monsterListFree_5174F0_native() {
	loadedMonsterDefs.clear()
}

func nox_xxx_monsterList_517520_native() int {
	for it := loadedMonsterDefs.head; it != nil; it = it.Next244 {
		name := GoStringS(it.Name0[:])
		it.TypeInd240 = uint32(GetServer().S().Types.IndByID(name))
		if it.TypeInd240 == 0 {
			gameLog.Printf("cannot resolve monster object type %q", name)
			loadedMonsterDefs.clear()
			return 0
		}
	}
	return 1
}

func nox_xxx_monsterDefByTT_517560_native(typ int) *server.MonsterDef {
	return loadedMonsterDefs.findByType(uint32(typ))
}

//export nox_xxx_loadMonsterBin_517010_go
func nox_xxx_loadMonsterBin_517010_go() C.int {
	return C.int(Nox_xxx_loadMonsterBin_517010())
}

//export nox_xxx_monsterListFree_5174F0_go
func nox_xxx_monsterListFree_5174F0_go() unsafe.Pointer {
	nox_xxx_monsterListFree_5174F0_native()
	return nil
}

//export nox_xxx_monsterList_517520_go
func nox_xxx_monsterList_517520_go() C.int {
	return C.int(nox_xxx_monsterList_517520_native())
}

//export nox_xxx_monsterDefByTT_517560_go
func nox_xxx_monsterDefByTT_517560_go(typ C.int) unsafe.Pointer {
	return unsafe.Pointer(nox_xxx_monsterDefByTT_517560_native(int(typ)))
}

//export nox_xxx_gamedataGetFloat_419D40
func nox_xxx_gamedataGetFloat_419D40(k *C.char) C.double {
	key := GoString(k)
	val := C.double(GetServer().S().Balance.Float(key))
	return val
}

//export nox_xxx_gamedataGetFloatTable_419D70
func nox_xxx_gamedataGetFloatTable_419D70(k *C.char, i_cgo int32) C.double {
	i := int(i_cgo)
	key := GoString(k)
	val := C.double(GetServer().S().Balance.FloatInd(key, i))
	return val
}
