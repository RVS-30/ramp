package cmd

import (
	"fmt"
	"strconv"
)

func Add(first string, second string) (result string) {
    num1, err := strconv.ParseFloat(first, 64)
	if err != nil {
		fmt.Println("Error parsing first number:", err)
		return "cannot be performed"
	}

	nums2, err := strconv.ParseFloat(second, 64)
	if err != nil {
		fmt.Println("Error parsing second number:", err)
		return "cannot be performed"
	}

	return fmt.Sprintf("%f", num1+nums2)
}

func Subtract(first string, second string) (result string) {
	num1, err := strconv.ParseFloat(first, 64)
	if err != nil {
		fmt.Println("Error parsing first number:", err)
		return
	}

	num2, err := strconv.ParseFloat(second, 64)
	if err != nil {
		fmt.Println("Error parsing second number:", err)
		return
	}

	return fmt.Sprintf("%f", num1-num2)
}