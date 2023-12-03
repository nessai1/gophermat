package order

import (
	"errors"
	"strconv"
)

var ErrInvalidOrderNumber = errors.New("invalid order number")

type DataSource interface {
	Begin() (*Transaction, error)
}

type Transaction interface {
	Commit() error
	Rollback() error
}

func IsOrderNumberCorrect(orderNumber string) bool {
	sum := 0

	numSize := len(orderNumber)
	for i := 0; i < numSize; i++ {
		num, err := strconv.Atoi(string(orderNumber[numSize-i-1]))
		if err != nil {
			return false
		}

		if i%2 == 1 {
			num *= 2

			if num >= 10 {
				buf := strconv.Itoa(num)
				v1, _ := strconv.Atoi(string(buf[0]))
				v2, _ := strconv.Atoi(string(buf[1]))
				num = v1 + v2
			}
		}

		sum += num
	}

	return sum%10 == 0
}
