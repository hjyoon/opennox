package blobs

import (
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadMemmap(t *testing.T) {
	blobTestFixture(t)
	// Parse the real mapping, but write the result only into the fixture.
	data, err := os.ReadFile("../../common/memmap/nox/noxmap.go")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(Path(memmapGo2), data, 0644))
	m, err := ReadMemmap()
	require.NoError(t, err)
	const varCnt = 1396
	require.GreaterOrEqual(t, len(m.Vars), varCnt)
	err = m.Write()
	require.NoError(t, err)
	written, err := ReadMemmap()
	require.NoError(t, err)
	require.Len(t, written.Vars, len(m.Vars))
	sort.Slice(m.Vars, func(i, j int) bool {
		return m.Vars[i].Size > m.Vars[j].Size && m.Vars[i].Off > 0
	})
	for _, v := range m.Vars[:10] {
		t.Logf("0x%X, %d [%d]", v.Blob, v.Off, v.Size)
	}
}
