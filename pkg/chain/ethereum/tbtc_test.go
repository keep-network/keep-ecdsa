package ethereum

import (
	"testing"
)

func TestParseUtxoOutpoint(t *testing.T) {
	testData := map[string]struct {
		utxoOutpoint            []uint8
		expectedTransactionHash string
		expectedOutputIndex     uint32
	}{
		"0-index": {
			[]uint8{
				84, 121, 222, 219, 15, 205, 170, 204, 170, 166,
				190, 232, 164, 145, 34, 53, 160, 239, 22, 193,
				165, 36, 226, 24, 247, 34, 199, 29, 229, 182,
				169, 95, 0, 0, 0, 0,
			},
			"5fa9b6e51dc722f718e224a5c116efa0352291a4e8bea6aaccaacd0fdbde7954",
			0,
		},
		"14-index": {
			[]uint8{
				222, 45, 61, 151, 209, 42, 39, 106, 9, 18,
				136, 250, 70, 152, 112, 235, 71, 222, 108, 253,
				210, 110, 210, 180, 214, 54, 49, 222, 146, 177,
				226, 76, 14, 0, 0, 0,
			},
			"4ce2b192de3136d6b4d26ed2fd6cde47eb709846fa8812096a272ad1973d2dde",
			14,
		},
		"300-index": {
			[]uint8{
				194, 93, 69, 179, 7, 146, 227, 106, 188, 242,
				116, 164, 222, 159, 208, 248, 123, 116, 86, 54,
				92, 157, 139, 246, 61, 161, 250, 45, 205, 230,
				140, 175, 44, 1, 0, 0,
			},
			"af8ce6cd2dfaa13df68b9d5c3656747bf8d09fdea474f2bc6ae39207b3455dc2",
			300,
		},
		"66333-index": {
			[]uint8{
				113, 173, 108, 198, 64, 153, 226, 149, 75, 126,
				163, 248, 117, 187, 34, 19, 238, 59, 82, 18,
				122, 29, 135, 136, 179, 34, 15, 29, 32, 16,
				241, 213, 29, 3, 1, 0,
			},
			"d5f110201d0f22b388871d7a12523bee1322bb75f8a37e4b95e29940c66cad71",
			66333,
		},
		"67175197-index": {
			[]uint8{
				118, 11, 174, 10, 36, 188, 103, 172, 237, 141,
				123, 53, 103, 227, 100, 132, 14, 191, 93, 253,
				103, 95, 152, 91, 214, 187, 60, 4, 110, 48,
				161, 239, 29, 3, 1, 4,
			},
			"efa1306e043cbbd65b985f67fd5dbf0e8464e367357b8dedac67bc240aae0b76",
			67175197,
		},
	}
	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			transactionHash, outputIndex := parseUtxoOutpoint(testData.utxoOutpoint)
			if transactionHash != testData.expectedTransactionHash {
				t.Errorf("unexpected transaction hash\nexpected: %s\nactual:   %s", testData.expectedTransactionHash, transactionHash)
			}
			if outputIndex != testData.expectedOutputIndex {
				t.Errorf("unexpected output index\nexpected: %d\nactual:   %d", testData.expectedOutputIndex, outputIndex)
			}
		})
	}
}
