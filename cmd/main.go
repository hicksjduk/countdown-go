package main

import (
	"fmt"
	"os"
	"strconv"

	"countdown/pkg/countdown"
)

func main() {
	args := os.Args[1:]
	if len(args) != 7 {
		fmt.Println("Argument list must contain 7 integers, the first of which is the target number")
		os.Exit(1)
	}
	intArgs := convertToInt(args)
	target, numbers := intArgs[0], intArgs[1:]
	if target < 100 || target > 999 {
		fmt.Println("Target number must be in the range 100 to 999 inclusive")
		os.Exit(1)
	}
	validateNumbersToUse(numbers)
	countdown.Solve(target, numbers...)
}

func convertToInt(args []string) []int {
	answer := make([]int, len(args))
	for i, v := range args {
		if n, err := strconv.Atoi(v); err != nil {
			fmt.Println("Argument list must contain 7 integers, the first of which is the target number")
			os.Exit(1)
		} else {
			answer[i] = n
		}
	}
	return answer
}

func validateNumbersToUse(numbers []int) {
	maxCounts := map[int]int{}
	for i := 1; i <= 10; i++ {
		maxCounts[i] = 2
		if i <= 4 {
			maxCounts[i*25] = 1
		}
	}
	for _, v := range numbers {
		if count, ok := maxCounts[v]; !ok {
			fmt.Println("Numbers to use must be in the range 1 to 10 inclusive, or 25, 50, 75 or 100")
			os.Exit(1)
		} else if count == 0 {
			fmt.Printf("Too many instances of %v - small numbers can appear twice, big numbers only once\n", v)
			os.Exit(1)
		} else {
			maxCounts[v]--
		}
	}
}
