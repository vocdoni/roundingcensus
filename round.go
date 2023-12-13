package roundedcensus

/*
Package roundedcensus provides an algorithm to anonymize participant balances in a voting system
while maintaining a certain level of accuracy. It sorts participants by balance, groups them
based on a privacy threshold and balance differences, rounds their balances, and calculates
lost balance for accuracy measurement.

The main steps of the algorithm are:

1. Sort Participants by Balance:
   - Participants are sorted in ascending order based on their balances.

2. Group Participants:
   - Participants are initially grouped with a size equal to the privacy threshold.
   - The group can extend if consecutive participants have the same balance or if
     the difference in balances between consecutive participants is less than or
     equal to the groupBalanceDiff threshold.

3. Round Group Balances:
   - Each group's balances are rounded down to the lowest value within that group.

4. Output Rounded Balances and Accuracy:
   - The algorithm provides the new list of participants with their rounded balances
     and the calculated accuracy to quantify the balance preservation.
*/

import (
	"fmt"
	"math/big"
	"sort"
)

// Participant represents a participant with an Ethereum address and balance.
type Participant struct {
	Address string
	Balance *big.Int
}

// ByBalance implements sort.Interface for []Participant based on the Balance field.
type ByBalance []*Participant

func (a ByBalance) Len() int           { return len(a) }
func (a ByBalance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByBalance) Less(i, j int) bool { return a[i].Balance.Cmp(a[j].Balance) < 0 }

// roundGroups rounds the balances within each group to the lowest value in the group.
func roundGroups(groups [][]*Participant) []*Participant {
	roundedCensus := []*Participant{}
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		lowestBalance := group[0].Balance
		for _, participant := range group {
			roundedCensus = append(roundedCensus, &Participant{Address: participant.Address, Balance: lowestBalance})
		}
	}
	return roundedCensus
}

// calculateAccuracy computes the accuracy of the rounding process.
func calculateAccuracy(original, rounded []*Participant) float64 {
	var totalOriginal, totalRounded big.Int
	for i := range original {
		totalOriginal.Add(&totalOriginal, original[i].Balance)
		totalRounded.Add(&totalRounded, rounded[i].Balance)
	}
	lostWeight := new(big.Float).Sub(new(big.Float).SetInt(&totalOriginal), new(big.Float).SetInt(&totalRounded))
	totalOriginalFloat := new(big.Float).SetInt(&totalOriginal)
	accuracy, _ := new(big.Float).Quo(lostWeight, totalOriginalFloat).Float64()
	return 100 - (accuracy * 100)
}

// groupAndRoundCensus groups the participants and rounds their balances.
func groupAndRoundCensus(participants []*Participant, privacyThreshold int, groupBalanceDiff *big.Int) ([]*Participant, float64) {
	sort.Sort(ByBalance(participants))

	var groups [][]*Participant
	var currentGroup []*Participant
	for i, participant := range participants {
		if len(currentGroup) == 0 {
			currentGroup = append(currentGroup, participant)
		} else {
			lastParticipant := currentGroup[len(currentGroup)-1]
			balanceDiff := new(big.Int).Abs(new(big.Int).Sub(participant.Balance, lastParticipant.Balance))

			if len(currentGroup) < privacyThreshold || balanceDiff.Cmp(groupBalanceDiff) <= 0 {
				currentGroup = append(currentGroup, participant)
			} else {
				groups = append(groups, currentGroup)
				currentGroup = []*Participant{participant}
			}
		}
		// Ensure the last group is added
		if i == len(participants)-1 {
			groups = append(groups, currentGroup)
		}
	}

	roundedCensus := roundGroups(groups)
	accuracy := calculateAccuracy(participants, roundedCensus)
	return roundedCensus, accuracy
}

func GroupAndRoundCensus(participants []*Participant, minPrivacyThreshold int, groupBalanceDiff *big.Int, minAccuracy float64) ([]*Participant, float64, error) {
	privacyThreshold := len(participants) / 2
	nearAccuracy := minAccuracy * 0.9
	for {
		if privacyThreshold <= minPrivacyThreshold {
			return nil, 0.0, fmt.Errorf("could not find a privacy threshold that satisfies the minimum accuracy")
		}
		_, lastAccuracy := groupAndRoundCensus(participants, privacyThreshold, groupBalanceDiff)
		if lastAccuracy >= nearAccuracy {
			break
		}
		privacyThreshold /= 2
	}

	for privacyThreshold > minPrivacyThreshold {
		finalCensus, accuracy := groupAndRoundCensus(participants, privacyThreshold, groupBalanceDiff)
		if accuracy >= minAccuracy {
			return finalCensus, accuracy, nil
		}
		privacyThreshold--
	}
	return nil, nearAccuracy, fmt.Errorf("could not find a privacy threshold that satisfies the minimum accuracy")
}
