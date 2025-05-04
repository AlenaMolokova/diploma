package utils

func LuhnCheck(number string) bool {
	if number == "" {
		return false
	}

	var sum int
	isEven := false

	for i := len(number) - 1; i >= 0; i-- {
		if number[i] < '0' || number[i] > '9' {
			return false
		}
		digit := int(number[i] - '0')
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
