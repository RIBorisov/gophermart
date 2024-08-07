package service

import (
	"errors"
)

const asciiTen = 57

func calculateLuhnSum(number string, parity int) (int64, error) {
	var sum int64
	for i, d := range number {
		if d < '0' || d > asciiTen {
			return 0, errors.New("invalid digit")
		}

		d -= '0'
		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += int64(d)
	}

	return sum, nil
}

func ValidateLuhn(number string) error {
	p := len(number) % 2
	sum, err := calculateLuhnSum(number, p)
	if err != nil {
		return err
	}

	// If the total modulo 10 is not equal to 0, then the number is invalid.
	if sum%10 != 0 {
		return errors.New("invalid number")
	}

	return nil
}
