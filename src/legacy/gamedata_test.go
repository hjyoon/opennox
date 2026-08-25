package legacy

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/stretchr/testify/require"

	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/server"
)

func newTestMonsterBinStore(freed *int) *monsterBinStore {
	return newMonsterBinStore(func() (*server.MonsterDef, func()) {
		return &server.MonsterDef{}, func() { *freed++ }
	})
}

func monsterDefName(def *server.MonsterDef) string {
	return GoStringS(def.Name0[:])
}

func TestMonsterBinStoreModesAndFields(t *testing.T) {
	const data = `
// the original reader ignores whole-line comments
Goblin
EXPERIENCE 42
HEALTH 100
QUESTHEALTH 175
SPEED 32
RETREAT_RATIO 0.25
RESUME_RATIO 0.75
FLEE_RANGE 90.5
STATUS CAN_BLOCK+CAN_RUN
RUN_MULTIPLIER 1.5
MOVE_SOUND_FRAME_A 4
MOVE_SOUND_FRAME_B 8
MELEE_ATTACK_FRAME 10
MELEE_ATTACK_RANGE 12.5
MELEE_ATTACK_DAMAGE 14
MELEE_ATTACK_IMPACT 2.25
MELEE_ATTACK_DAMAGE_TYPE DAMAGE_BLADE
MELEE_ATTACK_MIN_DELAY 20
MELEE_ATTACK_MAX_DELAY 30
MELEE_ATTACK_POISON_CHANCE 40
MELEE_ATTACK_POISON_STRENGTH 50
MELEE_ATTACK_POISON_MAX 60
MISSILE_NAME NULL
MISSILE_ATTACK_RANGE 100.5
MISSILE_ATTACK_FRAME 11
MISSILE_ATTACK_MIN_DELAY 21
MISSILE_ATTACK_MAX_DELAY 31
MELEE_STRIKE_FUNCTION NULL
DIE_FUNCTION NULL
DEAD_FUNCTION NULL
SOLO HEALTH 200
ARENA HEALTH 300
END
`

	for _, tc := range []struct {
		name       string
		mode       monsterBinMode
		wantHealth uint32
	}{
		{name: "solo", mode: monsterBinMode{solo: true}, wantHealth: 200},
		{name: "arena", mode: monsterBinMode{arena: true}, wantHealth: 300},
	} {
		t.Run(tc.name, func(t *testing.T) {
			freed := 0
			store := newTestMonsterBinStore(&freed)
			require.NoError(t, store.load(strings.NewReader(data), tc.mode))
			def := store.head
			require.NotNil(t, def)
			require.Equal(t, "Goblin", monsterDefName(def))
			require.Equal(t, uint32(42), def.Experience64)
			require.Equal(t, tc.wantHealth, def.Health68)
			require.Equal(t, uint32(175), def.HealthQuest72)
			require.Equal(t, uint32(32), def.Speed76)
			require.Equal(t, float32(0.25), def.RetreatRatio80)
			require.Equal(t, float32(0.75), def.ResumeRatio84)
			require.Equal(t, float32(90.5), def.FleeRange88)
			require.Equal(t, object.MonStatusCanBlock|object.MonStatusCanRun, def.StatusFlags92)
			require.Equal(t, float32(1.5), def.RunMultiplier96)
			require.Equal(t, uint32(4), def.MoveSndFrameA100)
			require.Equal(t, uint32(8), def.MoveSndFrameB104)
			require.Equal(t, uint32(10), def.MeleeAttackFrame108)
			require.Equal(t, float32(12.5), def.MeleeAttackRange112)
			require.Equal(t, uint32(14), def.MeleeAttackDamage116)
			require.Equal(t, float32(2.25), def.MeleeAttackImpact120)
			require.Equal(t, uint32(object.DamageBlade), def.MeleeAttackDamageType124)
			require.Equal(t, uint32(20), def.MeleeAttackDelayMin128)
			require.Equal(t, uint32(30), def.MeleeAttackDelayMax132)
			require.Equal(t, uint32(40), def.MeleeAttackPoisonChange136)
			require.Equal(t, uint32(50), def.MeleeAttackPoisonStrength140)
			require.Equal(t, uint32(60), def.MeleeAttackPoisonMax144)
			require.Empty(t, GoStringS(def.MissileName148[:]))
			require.Equal(t, float32(100.5), def.MissileAttackRange212)
			require.Equal(t, uint32(11), def.MissileAttackFrame216)
			require.Equal(t, uint32(21), def.MissileAttackDelayMin220)
			require.Equal(t, uint32(31), def.MissileAttackDelayMax224)
			require.Nil(t, def.MeleeStrikeFunc236)
			require.Nil(t, def.DieFunc228)
			require.Nil(t, def.DeadFunc232)

			def.TypeInd240 = 77
			require.Same(t, def, store.findByType(77))
			require.Nil(t, store.findByType(78))
			store.clear()
			require.Nil(t, store.head)
			require.Equal(t, 1, freed)
		})
	}
}

func TestMonsterBinStoreRejectsUnknownFieldWithoutLeaks(t *testing.T) {
	freed := 0
	store := newTestMonsterBinStore(&freed)
	err := store.load(strings.NewReader("Goblin HEALTH 10 END Ogre UNKNOWN 1 END"), monsterBinMode{solo: true})
	require.ErrorContains(t, err, `monster "Ogre" field "UNKNOWN": unknown field`)
	require.Nil(t, store.head)
	require.Equal(t, 2, freed)
}

func TestMonsterBinStorePreservesNullDefinitionName(t *testing.T) {
	freed := 0
	store := newTestMonsterBinStore(&freed)
	require.NoError(t, store.load(strings.NewReader("NULL END"), monsterBinMode{arena: true}))
	require.Equal(t, "NULL", monsterDefName(store.head))
	store.clear()
	require.Equal(t, 1, freed)
}

func TestMonsterTokenReaderRestartsAfterComment(t *testing.T) {
	r := &monsterTokenReader{r: bufioNewReaderForMonsterTest("discard// comment\nHEALTH 10")}
	tok, err := r.next()
	require.NoError(t, err)
	require.Equal(t, "HEALTH", tok)
	tok, err = r.next()
	require.NoError(t, err)
	require.Equal(t, "10", tok)
}

func TestMonsterTokenReaderStopsAtNULTerminator(t *testing.T) {
	r := &monsterTokenReader{r: bufioNewReaderForMonsterTest("Goblin END \x00\x00")}
	tok, err := r.next()
	require.NoError(t, err)
	require.Equal(t, "Goblin", tok)
	tok, err = r.next()
	require.NoError(t, err)
	require.Equal(t, "END", tok)
	_, err = r.next()
	require.ErrorIs(t, err, io.EOF)
}

func bufioNewReaderForMonsterTest(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

func TestMonsterBinOracleFile(t *testing.T) {
	path := os.Getenv("NOX_MONSTER_BIN_TEST")
	if path == "" {
		t.Skip("set NOX_MONSTER_BIN_TEST to validate an installed monster.bin")
	}
	for _, mode := range []monsterBinMode{{solo: true}, {arena: true}} {
		f, err := binfile.BinfileOpen(path, binfile.ReadOnly)
		require.NoError(t, err)
		require.NoError(t, f.SetKey(23))
		freed := 0
		store := newTestMonsterBinStore(&freed)
		require.NoError(t, store.load(f, mode))
		require.NotNil(t, store.head)
		count := 0
		for it := store.head; it != nil; it = it.Next244 {
			count++
			require.NotEmptyf(t, monsterDefName(it), "definition %d has raw name %q", count, it.Name0)
		}
		require.Greater(t, count, 1)
		store.clear()
		require.Equal(t, count, freed)
		require.NoError(t, f.Close())
	}
}
