package roundedcensus

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.vocdoni.io/dvote/util"
)

// openCensus opens a census file and returns a list of participants.
func openCensus(filename string) ([]*Participant, error) {
	testFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	testData := map[string]string{}
	if err := json.Unmarshal(testFile, &testData); err != nil {
		panic(err)
	}
	census := []*Participant{}
	for k, v := range testData {
		value, ok := new(big.Int).SetString(v, 10)
		if !ok {
			panic("Invalid value")
		}
		census = append(census, &Participant{Address: k, Balance: value})
	}
	return census, nil
}

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

	var err error
	var census []*Participant
	if testCensus := os.Getenv("TEST_CENSUS"); testCensus != "" {
		if census, err = openCensus(testCensus); err != nil {
			t.Fatalf("Error opening census: %v", err)
		}
	} else {
		census = generateRandomCensus(censusSize, maxBalance)
	}

	privacyThreshold, err := strconv.Atoi(os.Getenv("PRIVACY_THRESHOLD"))
	if err != nil {
		t.Fatalf("Error parsing PRIVACY_THRESHOLD: %v", err)
	}
	groupBalanceDiff, err := strconv.Atoi(os.Getenv("GROUP_BALANCE_DIFF"))
	if err != nil {
		t.Fatalf("Error parsing GROUP_BALANCE_DIFF: %v", err)
	}
	roundedCensus := groupAndRoundCensus(census, privacyThreshold, big.NewInt(int64(groupBalanceDiff)))
	accuracy := calculateAccuracy(census, roundedCensus)
	distinctBalances := []string{}
	groupsCounters := map[string]int{}
	for _, p := range roundedCensus {
		if _, exists := groupsCounters[p.Balance.String()]; !exists {
			groupsCounters[p.Balance.String()] = 1
			distinctBalances = append(distinctBalances, p.Balance.String())
		} else {
			groupsCounters[p.Balance.String()]++
		}
	}
	t.Logf("Privacy Threshold: %d, Accuracy: %.2f%%, Groups: %d, Holders: %d\n",
		privacyThreshold, accuracy, len(distinctBalances), len(census))
}

func TestAutoRoundingAlgorithm(t *testing.T) {
	censusSize := 10000 // Configurable number of addresses
	maxBalance := int64(10000000)

	var err error
	var census []*Participant
	output_census := "./census_rounded.json"
	if testCensus := os.Getenv("TEST_CENSUS"); testCensus != "" {
		if census, err = openCensus(testCensus); err != nil {
			t.Fatalf("Error opening census: %v", err)
		}
		census_filename := filepath.Base(testCensus)
		census_folder := strings.TrimSuffix(testCensus, census_filename)
		census_ext := filepath.Ext(census_filename)
		census_base := strings.TrimSuffix(census_filename, census_ext)
		output_census = fmt.Sprintf("%s%s_rounded.json", census_folder, census_base)
	} else {
		census = generateRandomCensus(censusSize, maxBalance)
	}

	minPrivacyThreshold, err := strconv.Atoi(os.Getenv("MIN_PRIVACY_THRESHOLD"))
	if err != nil {
		t.Fatalf("Error parsing MIN_PRIVACY_THRESHOLD: %v", err)
	}
	minAccuracy, err := strconv.ParseFloat(os.Getenv("MIN_ACCURACY"), 64)
	if err != nil {
		t.Fatalf("Error parsing MIN_ACCURACY: %v", err)
	}
	groupBalanceDiff, err := strconv.Atoi(os.Getenv("GROUP_BALANCE_DIFF"))
	if err != nil {
		t.Fatalf("Error parsing GROUP_BALANCE_DIFF: %v", err)
	}

	roundedCensus, accuracy, err := GroupAndRoundCensus(census, minPrivacyThreshold, big.NewInt(int64(groupBalanceDiff)), minAccuracy)
	if err != nil {
		t.Fatalf("Error rounding census: %v", err)
	}
	groupsCounters := map[string]int{}
	for _, p := range roundedCensus {
		if _, exists := groupsCounters[p.Balance.String()]; !exists {
			groupsCounters[p.Balance.String()] = 1
		} else {
			groupsCounters[p.Balance.String()]++
		}
	}
	fd, err := os.Create(output_census)
	if err != nil {
		t.Fatalf("Error creating file: %v", err)
	}
	defer fd.Close()
	jsonCensus := map[string]string{}
	for _, p := range roundedCensus {
		jsonCensus[p.Address] = p.Balance.String()
	}

	jsonData, err := json.Marshal(jsonCensus)
	if err != nil {
		t.Fatalf("Error marshalling data: %v", err)
	}
	fd.Write(jsonData)
	t.Logf("Min Privacy Threshold: %d, MinAccuracy: %.2f%%, Accuracy: %.2f%%, Groups: %d, Holders: %d\n",
		minPrivacyThreshold, minAccuracy, accuracy, len(groupsCounters), len(census))
}
