package committee_record

import (
	"testing"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

func TestCommitteeRecord(t *testing.T) {
	keyPair := ed25519.GenerateKeyPair()
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	rec := NewCommitteeRecord(addr, "node:111", "node:333")
	recBack, err := CommitteeRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.Address.Equals(recBack.Address))
	require.EqualValues(t, rec.Nodes, recBack.Nodes)

	t.Logf("%s", rec)
}
