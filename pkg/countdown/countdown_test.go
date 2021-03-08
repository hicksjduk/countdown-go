package countdown

import (
	"fmt"
	"testing"
)

func TestPermuteAllDifferent(t *testing.T) {
	testPermute(t, []int{1, 2, 3, 4, 5, 6}, 1956)
}

func TestPermuteAllPairsOfDuplicates(t *testing.T) {
	testPermute(t, []int{1, 2, 2, 3, 1, 3}, 270)
}

func testPermute(t *testing.T, input []int, expectedCount int) {
	exprs := make([]Expression, len(input))
	for i, v := range input {
		exprs[i] = numberExpression(v)
	}
	results := map[string]interface{}{}
	for v := range permute(exprs) {
		str := fmt.Sprintf("%v", v)
		if _, ok := results[str]; ok {
			t.Errorf("Duplicate result: %v", str)
		} else {
			results[str] = nil
		}
	}
	if actualCount := len(results); expectedCount != actualCount {
		t.Errorf("Expected %d result(s) but got %d", expectedCount, actualCount)
	}
}

func TestSolveWithSolution(t *testing.T) {
	solution := getSolution(834, 10, 9, 8, 7, 6, 5)
	validateSolution(t, solution, 834)
}

func TestSolveWithSolution2(t *testing.T) {
	solution := getSolution(378, 50, 7, 4, 3, 2, 1)
	validateSolution(t, solution, 378)
}

func TestSolveWithSolution3(t *testing.T) {
	solution := getSolution(493, 50, 25, 4, 3, 2, 4)
	validateSolution(t, solution, 493)
}

func TestSolveWithSolution4(t *testing.T) {
	solution := getSolution(803, 50, 4, 9, 6, 6, 1)
	validateSolution(t, solution, 803)
}

func TestSolveWithNonExactSolution(t *testing.T) {
	solution := getSolution(954, 50, 75, 25, 100, 5, 8)
	validateSolution(t, solution, 955)
}

func TestSolveWithoutSolution(t *testing.T) {
	solution := getSolution(999, 1, 2, 3, 4, 5, 6)
	validateSolution(t, solution, 0)
}

func validateSolution(t *testing.T, solution Expression, expectedValue int) {
	if expectedValue != solution.value {
		t.Errorf("Expected solution with value %v but got %v", expectedValue, solution.value)
	}
}

func getSolution(target int, numbers ...int) Expression {
	var solution Expression
	for solution = range Solve(target, numbers...) {
		fmt.Printf("%v = %v\n", solution.print, solution.value)
	}
	return solution
}
