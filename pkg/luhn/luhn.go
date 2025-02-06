package luhn

import (
	"strconv"
	"strings"
)

const limitToReplace = 9

func Check(number string) bool {
	number = strings.ReplaceAll(number, " ", "")
	if number == "" {
		return false
	}

	sum := 0
	nDigits := len(number)
	parity := nDigits % 2

	for i := 0; i < nDigits; i++ {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if i%2 == parity {
			digit *= 2
			if digit > limitToReplace {
				digit -= limitToReplace
			}
		}

		sum += digit
	}

	return sum%10 == 0
}
