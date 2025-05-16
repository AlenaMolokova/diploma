package validation

import (
	"regexp"
)

type OrderValidator interface {
	ValidateOrderNumber(orderNumber string) bool
}

type LuhnValidator struct {
	digitRegex *regexp.Regexp
}

func NewLuhnValidator() *LuhnValidator {
	return &LuhnValidator{
		digitRegex: regexp.MustCompile(`^\d+$`),
	}
}

func (v *LuhnValidator) ValidateOrderNumber(orderNumber string) bool {
	if orderNumber == "" || !v.digitRegex.MatchString(orderNumber) {
		return false
	}

	var sum int
	isEven := false

	for i := len(orderNumber) - 1; i >= 0; i-- {
		digit := int(orderNumber[i] - '0')
		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}
