package math

import (
	"math/rand"
	"time"
)

type Number interface {
	int | float32
}

func Max[T Number](num1, num2 T) T {
	if num1 >= num2 {
		return num1
	}
	return num2
}

func Min[T Number](num1, num2 T) T {
	if num1 >= num2 {
		return num2
	}
	return num1
}

func GenRandomNumber() int {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(9000) + 1000
}
