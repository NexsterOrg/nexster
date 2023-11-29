package math

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
