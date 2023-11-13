package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func IsBinary(contents []byte) bool {
	for _, ch := range contents {
		if ch == 0 {
			return true
		}
	}
	return false
}

func RollDice(diceRoll string) (int, error) {
	diceParts := strings.Split(strings.ToLower(diceRoll), "d")
	if len(diceParts) != 2 {
		return 0, fmt.Errorf("invalid diceRoll format: %s", diceRoll)
	}

	diceNum, err := strconv.Atoi(diceParts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid number of dices: %s", diceParts[0])
	}

	diceSides, err := strconv.Atoi(diceParts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid dice sides: %s", diceParts[1])
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	sum := 0
	for i := 0; i < diceNum; i++ {
		sum += rand.Intn(diceSides) + 1
	}

	return sum, nil
}
