package countdown

import (
	"fmt"
	"runtime"
	"sync"
)

func Solve(target int, numbers ...int) chan *Expression {
	exprs := make([]*Expression, len(numbers))
	for i, v := range numbers {
		exprs[i] = numberExpression(v)
	}
	return solve(target, exprs)
}

var processCount = runtime.GOMAXPROCS(0)

func solve(target int, numbers []*Expression) chan *Expression {
	combs := combinations(numbers)
	accumulator := make(chan *Expression, processCount)
	findBest := evaluator(target)
	var wg sync.WaitGroup
	wg.Add(processCount)
	for i := processCount; i > 0; i-- {
		go func() {
			defer wg.Done()
			for e := range findBest(combs) {
				accumulator <- e
			}
		}()
	}
	go func() {
		wg.Wait()
		close(accumulator)
	}()
	return findBest(accumulator)
}

type Priority int

const (
	low Priority = iota
	high
	atomic
)

type operation struct {
	symbol      string
	priority    Priority
	commutative bool
	evaluator   func(int, int) int
}

var (
	addition = operation{
		symbol:      "+",
		priority:    low,
		commutative: true,
		evaluator: func(a, b int) int {
			return a + b
		},
	}
	subtraction = operation{
		symbol:      "-",
		priority:    low,
		commutative: false,
		evaluator: func(a, b int) int {
			return a - b
		},
	}
	multiplication = operation{
		symbol:      "*",
		priority:    high,
		commutative: true,
		evaluator: func(a, b int) int {
			return a * b
		},
	}
	division = operation{
		symbol:      "/",
		priority:    high,
		commutative: false,
		evaluator: func(a, b int) int {
			return a / b
		},
	}
)

type Expression struct {
	value    int
	print    func() string
	priority Priority
	numbers  []int
}

func (e *Expression) String() string {
	return fmt.Sprintf("%v = %v", e.print(), e.value)
}

func numberExpression(n int) *Expression {
	return &Expression{
		value: n,
		print: func() string {
			return fmt.Sprintf("%d", n)
		},
		priority: atomic,
		numbers:  []int{n},
	}
}

func arithmeticExpression(leftOperand *Expression, operator operation, rightOperand *Expression) *Expression {
	return &Expression{
		value: operator.evaluator(leftOperand.value, rightOperand.value),
		print: func() string {
			return printExpression(leftOperand, operator, rightOperand)
		},
		priority: operator.priority,
		numbers:  concatIntSlices(leftOperand.numbers, rightOperand.numbers),
	}
}

func concatIntSlices(slices ...[]int) []int {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	answer := make([]int, 0, length)
	for _, slice := range slices {
		answer = append(answer, slice...)
	}
	return answer
}

func printExpression(leftOperand *Expression, operator operation, rightOperand *Expression) string {
	return fmt.Sprintf("%v %v %v",
		withParensIfNecessary(leftOperand.print(), leftOperand.priority < operator.priority),
		operator.symbol,
		withParensIfNecessary(rightOperand.print(),
			rightOperand.priority < operator.priority ||
				(rightOperand.priority == operator.priority && !operator.commutative)))
}

func withParensIfNecessary(s string, parensNeeded bool) string {
	if parensNeeded {
		return fmt.Sprintf("(%v)", s)
	}
	return s
}

type combiner func(*Expression) (*Expression, bool)

type combinerCreator func(*Expression) (combiner, bool)

func addCombinerCreator(left *Expression) (combiner, bool) {
	return func(right *Expression) (*Expression, bool) {
		return arithmeticExpression(left, addition, right), true
	}, true
}

func subtractCombinerCreator(left *Expression) (combiner, bool) {
	if left.value < 3 {
		return nil, false
	}
	return func(right *Expression) (*Expression, bool) {
		if left.value <= right.value || left.value == right.value*2 {
			return nil, false
		}
		return arithmeticExpression(left, subtraction, right), true
	}, true
}

func multiplyCombinerCreator(left *Expression) (combiner, bool) {
	if left.value == 1 {
		return nil, false
	}
	return func(right *Expression) (*Expression, bool) {
		if right.value == 1 {
			return nil, false
		}
		return arithmeticExpression(left, multiplication, right), true
	}, true
}

func divideCombinerCreator(left *Expression) (combiner, bool) {
	if left.value == 1 {
		return nil, false
	}
	return func(right *Expression) (*Expression, bool) {
		if right.value == 1 || left.value%right.value != 0 || left.value == right.value*right.value {
			return nil, false
		}
		return arithmeticExpression(left, division, right), true
	}, true
}

var combinerCreators = []combinerCreator{
	addCombinerCreator,
	subtractCombinerCreator,
	multiplyCombinerCreator,
	divideCombinerCreator,
}

func combinations(exprs []*Expression) chan *Expression {
	answer := make(chan *Expression, processCount)
	go func() {
		defer close(answer)
		for p := range permute(exprs) {
			for c := range combine(p) {
				answer <- c
			}
		}
	}()
	return answer
}

func permute(exprs []*Expression) chan []*Expression {
	answer := make(chan []*Expression, processCount)
	go func() {
		defer close(answer)
		if count := len(exprs); count == 1 {
			answer <- exprs
		} else {
			used := usedTracker()
			for i, expr := range exprs {
				if !used(expr.value) {
					currentExpr := exprs[i : i+1]
					answer <- currentExpr
					for es := range permute(concatExpressionSlices(exprs[:i], exprs[i+1:])) {
						answer <- concatExpressionSlices(currentExpr, es)
					}
				}
			}
		}
	}()
	return answer
}

func usedTracker() func(int) bool {
	used := map[int]interface{}{}
	return func(v int) bool {
		_, answer := used[v]
		if !answer {
			used[v] = nil
		}
		return answer
	}
}

func concatExpressionSlices(slices ...[]*Expression) []*Expression {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	answer := make([]*Expression, 0, length)
	for _, s := range slices {
		answer = append(answer, s...)
	}
	return answer
}

func combinersUsing(left *Expression) []combiner {
	answer := make([]combiner, 0, len(combinerCreators))
	for _, cc := range combinerCreators {
		if c, ok := cc(left); ok {
			answer = append(answer, c)
		}
	}
	return answer
}

func combine(exprs []*Expression) chan *Expression {
	answer := make(chan *Expression, processCount)
	go func() {
		defer close(answer)
		if size := len(exprs); size == 1 {
			answer <- exprs[0]
		} else {
			used := usedTracker()
			for i := 1; i < size; i++ {
				for left := range combine(exprs[:i]) {
					combiners := combinersUsing(left)
					for right := range combine(exprs[i:]) {
						for _, comb := range combiners {
							if expr, ok := comb(right); ok && !used(expr.value) {
								answer <- expr
							}
						}
					}
				}
			}
		}
	}()
	return answer
}

func evaluator(target int) func(chan *Expression) chan *Expression {
	differenceFromTarget := differenceFrom(target)
	return func(exprs chan *Expression) chan *Expression {
		answer := make(chan *Expression, processCount)
		go func() {
			defer close(answer)
			bestSoFar := struct {
				diff  int
				count int
			}{diff: 11}
			for e := range exprs {
				switch diff := differenceFromTarget(e); {
				case diff > 10 || diff > bestSoFar.diff:
				case diff < bestSoFar.diff:
					answer <- e
					bestSoFar.diff, bestSoFar.count = diff, len(e.numbers)
				default:
					if count := len(e.numbers); count < bestSoFar.count {
						answer <- e
						bestSoFar.diff, bestSoFar.count = diff, count
					}
				}
			}
		}()
		return answer
	}
}

func differenceFrom(target int) func(*Expression) int {
	return func(e *Expression) int {
		if diff := target - e.value; diff < 0 {
			return diff * -1
		} else {
			return diff
		}
	}
}
