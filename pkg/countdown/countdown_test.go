package countdown

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/cucumber/godog"
)

var opts = godog.Options{
	Format:   "progress",
	NoColors: true,
}

func TestMain(m *testing.M) {
	godog.TestSuite{
		Name:                "Countdown",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()
}

type testContext struct {
	target  int
	numbers []int
	result  *Expression
}

func (tc *testContext) callSolver(target int, numbers string) error {
	tc.target = target
	tc.numbers = toNumberArray(numbers)
	tc.result = Solve(target, tc.numbers...)
	return nil
}

func toNumberArray(str string) []int {
	splitRe := regexp.MustCompile(`\D+`)
	items := splitRe.Split(str, -1)
	nums := make([]int, len(items))
	for i, item := range items {
		nums[i], _ = strconv.Atoi(item)
	}
	return nums
}

func (tc *testContext) checkExactSolution(expectedCount int) error {
	return tc.checkSolution(tc.target, expectedCount)
}

func (tc *testContext) checkSolution(expectedValue, expectedCount int) error {
	if expected, actual := expectedValue, tc.result.value; expected != actual {
		return fmt.Errorf("Expected solution with value %v but got %v", expected, actual)
	}
	if expected, actual := expectedCount, len(tc.result.numbers); expected != actual {
		return fmt.Errorf("Expected solution that uses %v number(s) but got %v", expected, actual)
	}
	countVal := countValidator(tc.numbers)
	for _, actual := range tc.result.numbers {
		if err := countVal(actual); err != nil {
			return err
		}
	}
	return nil
}

func countValidator[K comparable](values []K) func(K) error {
	expected := map[K]int{}
	for _, n := range values {
		expected[n] += 1
	}
	actual := map[K]int{}
	return func(v K) error {
		actual[v] += 1
		if actual[v] > expected[v] {
			return fmt.Errorf("Solution has too many occurrences of %v", v)
		}
		return nil
	}
}

func (tc *testContext) checkNoSolution() error {
	if tc.result != nil {
		return fmt.Errorf("Expected no solution but there was one")
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := &testContext{}
	ctx.Step(`^I call the solver with target number (\d+) and numbers (\d+(?:\s*,\s*\d+)*)$`,
		tc.callSolver)
	ctx.Step(`^a solution is found whose value equals the target number and which uses (\d+) numbers$`,
		tc.checkExactSolution)
	ctx.Step(`^a solution is found whose value equals (\d+) and which uses (\d+) numbers$`,
		tc.checkSolution)
	ctx.Step(`^no solution is found$`, tc.checkNoSolution)
}
