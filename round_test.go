package roundedcensus

import (
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"go.vocdoni.io/dvote/util"
)

// generateRandomCensus generates a random census of a given size.
func generateRandomCensus(size int, maxBalance int64) []*Participant {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	participants := []*Participant{}
	for i := 0; i < size; i++ {
		address := fmt.Sprintf("%x", util.RandomBytes(20))
		balance := big.NewInt(r.Int63n(maxBalance)) // Random balance up to 100,000
		participants = append(participants, &Participant{Address: address, Balance: balance})
	}
	return participants
}

// TestRoundingAlgorithm tests the rounding algorithm with different privacy thresholds.
func TestRoundingAlgorithm(t *testing.T) {
	censusSize := 10000 // Configurable number of addresses
	maxBalance := int64(10000000)
	census := generateRandomCensus(censusSize, maxBalance)

	privacyThreshold, err := strconv.Atoi(os.Getenv("PRIVACY_THRESHOLD"))
	if err != nil {
		t.Fatalf("Error parsing PRIVACY_THRESHOLD: %v", err)
	}
	groupBalanceDiff, err := strconv.Atoi(os.Getenv("GROUP_BALANCE_DIFF"))
	if err != nil {
		t.Fatalf("Error parsing GROUP_BALANCE_DIFF: %v", err)
	}
	_, accuracy := groupAndRoundCensus(census, privacyThreshold, big.NewInt(int64(groupBalanceDiff)))
	t.Logf("Privacy Threshold: %d, Accuracy: %.2f%%\n", privacyThreshold, accuracy)
	//	for _, p := range roundedCensus {
	//		t.Logf("%s: %d\n", p.Address, p.Balance)
	//	}
}
