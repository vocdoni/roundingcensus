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
		lowestBalance := roundToFirstCommonDigit(group)
		// lowestBalance := group[0].Balance
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

// groupAndRoundCensus groups the cleanedParticipants and rounds their balances.
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

// GroupAndRoundCensus groups the participants and rounds their balances. It
// rounds the balances of the participants with the highest accuracy possible
// while maintaining a minimum privacy threshold. It discards outliers from the
// rounding process but returns them in the final list of participants.
func GroupAndRoundCensus(participants []*Participant, minPrivacyThreshold int, groupBalanceDiff *big.Int, minAccuracy float64) ([]*Participant, float64, int, error) {
	// cleanedParticipants, outliers := detectLowerOutliers(participants, 5.0)
	cleanedParticipants, outliers := zScore(participants, 2.0)

	maxPrivacyThreshold := len(participants) / minPrivacyThreshold
	currentPrivacyThreshold := minPrivacyThreshold
	maxAccuracy := 0.0
	maxAccuracyPrivacyThreshold := currentPrivacyThreshold
	for currentPrivacyThreshold <= maxPrivacyThreshold {
		_, lastAccuracy := groupAndRoundCensus(cleanedParticipants, currentPrivacyThreshold, groupBalanceDiff)
		if lastAccuracy > maxAccuracy {
			maxAccuracy = lastAccuracy
			maxAccuracyPrivacyThreshold = currentPrivacyThreshold
		}
		gap := currentPrivacyThreshold / 33
		if gap < 1 {
			gap = 1
		}
		currentPrivacyThreshold += gap
	}
	roundCensus, finalAccuracy := groupAndRoundCensus(cleanedParticipants, maxAccuracyPrivacyThreshold, groupBalanceDiff)
	roundCensus = append(roundCensus, outliers...)
	if finalAccuracy < minAccuracy {
		return roundCensus, finalAccuracy, maxAccuracyPrivacyThreshold, fmt.Errorf("could not find a privacy threshold that satisfies the minimum accuracy")
	}
	return roundCensus, finalAccuracy, maxAccuracyPrivacyThreshold, nil
}

// zScore identifies and returns outliers based on a specified z-score
// threshold. The z-score is the number of standard deviations from the mean a
// data point is. For example, a z-score of 2 means the data point is 2 standard
// deviations above the mean. A z-score of -2 means it is 2 standard deviations
// below the mean. A number is considered an outlier if its z-score is greater
// than the threshold.
func zScore(participants []*Participant, threshold float64) ([]*Participant, []*Participant) {
	// calculate mean and standard deviation
	// mean = sum of all values / number of values
	// standard deviation = sqrt(sum of all (value - mean)^2 / number of values)
	mean := new(big.Float)
	stdDev := new(big.Float)
	n := new(big.Float).SetInt64(int64(len(participants)))
	fBalances := make([]*big.Float, len(participants))
	for i, p := range participants {
		fBalance := new(big.Float).SetInt(p.Balance)
		fBalances[i] = fBalance
		mean = new(big.Float).Add(mean, fBalance)
	}
	mean = new(big.Float).Quo(mean, n)
	for _, balance := range fBalances {
		diff := new(big.Float).Sub(balance, mean)
		stdDev = new(big.Float).Add(stdDev, new(big.Float).Mul(diff, diff))
	}
	stdDev = new(big.Float).Quo(stdDev, n)
	stdDev = new(big.Float).Sqrt(stdDev)
	// calculate z-score for each value to determine outliers
	outliers := make([]*Participant, 0)
	newParticipants := make([]*Participant, 0)
	for _, p := range participants {
		// z-score = (value - mean) / standard deviation
		fBalance := new(big.Float).SetInt(p.Balance)
		diff := new(big.Float).Sub(fBalance, mean)
		zScore := new(big.Float).Quo(diff, stdDev)
		// fmt.Println("zScore:", zScore, "isOutlier:", zScore.Cmp(big.NewFloat(threshold)) > 0, "value:", p.Balance, "mean:", mean, "stdDev:", stdDev, "address:", p.Address)
		// if z-score is greater than threshold, it is an outlier
		if zScore.Abs(zScore).Cmp(big.NewFloat(threshold)) > 0 {
			outliers = append(outliers, p)
		} else {
			newParticipants = append(newParticipants, p)
		}
	}
	return newParticipants, outliers
}

// detectLowerOutliers identifies and returns lower outliers based on a specified lower percentile.
func detectLowerOutliers(participants []*Participant, lowerPercentile float64) ([]*Participant, []*Participant) {
	sort.Sort(ByBalance(participants))
	thresholdIndex := int(lowerPercentile / 100 * float64(len(participants)))
	thresholdValue := participants[thresholdIndex].Balance

	newParticipants := []*Participant{}
	outliers := []*Participant{}
	for _, p := range participants {
		if p.Balance.Cmp(thresholdValue) < 0 {
			outliers = append(outliers, p)
		} else {
			newParticipants = append(newParticipants, p)
		}
	}
	return newParticipants, outliers
}

func roundToFirstCommonDigit(participants []*Participant) *big.Int {
	// check if at least two numbers is provided
	if len(participants) == 0 {
		return big.NewInt(0)
	}
	if len(participants) == 1 {
		return participants[0].Balance
	}
	// get the minimun length of any number
	sBalances := []string{}
	minBalance := participants[0].Balance
	minLenght := int64(len(participants[0].Balance.String()))
	for _, n := range participants {
		sBalances = append(sBalances, n.Balance.String())
		if l := int64(len(n.Balance.String())); l < minLenght {
			minLenght = l
			minBalance = n.Balance
		}
	}

	firstCommonByte := int64(minLenght - 1)
	for firstCommonByte >= 0 {
		commonNumber := true
		currentNumber := sBalances[0][firstCommonByte]
		for _, n := range sBalances[1:] {
			if n[firstCommonByte] != currentNumber {
				commonNumber = false
				break
			}
		}
		if commonNumber {
			firstCommonByte++
			padding := new(big.Int).Exp(big.NewInt(10), big.NewInt(minLenght-firstCommonByte), nil)
			rounded := new(big.Int)
			rounded.SetString(sBalances[0][:firstCommonByte], 10)
			return rounded.Mul(rounded, padding)
		}
		firstCommonByte--
	}
	return minBalance
}
