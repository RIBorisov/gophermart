package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	asciiZero = 48
	asciiTen  = 57
)

// Generate returns a valid Luhn number of the provided length
func Generate(length int) string {
	rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	var s strings.Builder
	for i := 0; i < length-1; i++ {
		s.WriteString(strconv.Itoa(rand.Intn(9)))
	}

	_, res, _ := Calculate(s.String())
	return res
}

func calculateLuhnSum(number string, parity int) (int64, error) {
	var sum int64
	for i, d := range number {
		if d < asciiZero || d > asciiTen {
			return 0, errors.New("invalid digit")
		}

		d = d - asciiZero
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

func Calculate(number string) (string, string, error) {
	p := (len(number) + 1) % 2
	sum, err := calculateLuhnSum(number, p)
	if err != nil {
		return "", "", nil
	}

	luhn := sum % 10
	if luhn != 0 {
		luhn = 10 - luhn
	}

	return strconv.FormatInt(luhn, 10), fmt.Sprintf("%s%d", number, luhn), nil
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
