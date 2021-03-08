package countdown

import (
	"fmt"
	"sync"
)

func Solve(target int, numbers ...int) chan Expression {
	exprs := make([]Expression, len(numbers))
	for i, v := range numbers {
		exprs[i] = numberExpression(v)
	}
	return solve(target, exprs)
}

const processCount = 5

func solve(target int, numbers []Expression) chan Expression {
	combs := combinations(numbers)
	accumulator := make(chan Expression)
	go func() {
		var wg sync.WaitGroup
		wg.Add(processCount)
		defer func() {
			wg.Wait()
			close(accumulator)
		}()
		for i := processCount; i > 0; i-- {
			go func() {
				defer wg.Done()
				for e := range findBest(target, combs) {
					accumulator <- e
				}
			}()
		}
	}()
	return findBest(target, accumulator)
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
	operations = []operation{addition, subtraction, multiplication, division}
)

type Expression struct {
	value       int
	print       string
	priority    Priority
	numberCount int
}

func (e Expression) String() string {
	return fmt.Sprintf("%v = %v", e.print, e.value)
}

func numberExpression(n int) Expression {
	return Expression{
		value:       n,
		print:       fmt.Sprintf("%d", n),
		priority:    atomic,
		numberCount: 1,
	}
}

func arithmeticExpression(leftOperand Expression, operator operation, rightOperand Expression) Expression {
	return Expression{
		value:       operator.evaluator(leftOperand.value, rightOperand.value),
		print:       printExpression(leftOperand, operator, rightOperand),
		priority:    operator.priority,
		numberCount: leftOperand.numberCount + rightOperand.numberCount,
	}
}

func printExpression(leftOperand Expression, operator operation, rightOperand Expression) string {
	return fmt.Sprintf("%v %v %v",
		withParensIfNecessary(leftOperand.print, func() bool {
			return leftOperand.priority < operator.priority
		}),
		operator.symbol,
		withParensIfNecessary(rightOperand.print, func() bool {
			return rightOperand.priority < operator.priority ||
				(rightOperand.priority == operator.priority && !operator.commutative)
		}))
}

func withParensIfNecessary(s string, parensNeeded func() bool) string {
	if parensNeeded() {
		return fmt.Sprintf("(%v)", s)
	}
	return s
}

type combiner func(Expression, Expression) (Expression, bool)

func addCombiner(left, right Expression) (Expression, bool) {
	return arithmeticExpression(left, addition, right), true
}

func subtractCombiner(left, right Expression) (Expression, bool) {
	if left.value <= right.value {
		return Expression{}, false
	}
	if left.value == right.value*2 {
		return Expression{}, false
	}
	return arithmeticExpression(left, subtraction, right), true
}

func multiplyCombiner(left, right Expression) (Expression, bool) {
	if left.value == 1 || right.value == 1 {
		return Expression{}, false
	}
	return arithmeticExpression(left, multiplication, right), true
}

func divideCombiner(left, right Expression) (Expression, bool) {
	if left.value == 1 || right.value == 1 {
		return Expression{}, false
	}
	if left.value%right.value != 0 || left.value == right.value*right.value {
		return Expression{}, false
	}
	return arithmeticExpression(left, division, right), true
}

var combiners = []combiner{addCombiner, subtractCombiner, multiplyCombiner, divideCombiner}

func combinations(exprs []Expression) chan Expression {
	answer := make(chan Expression)
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

func permute(exprs []Expression) chan []Expression {
	answer := make(chan []Expression)
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
					for es := range permute(copyAppend(exprs[:i], exprs[i+1:])) {
						answer <- copyAppend(currentExpr, es)
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

func copyAppend(arrs ...[]Expression) []Expression {
	length := 0
	for _, arr := range arrs {
		length += len(arr)
	}
	answer := make([]Expression, 0, length)
	for _, arr := range arrs {
		answer = append(answer, arr...)
	}
	return answer
}

func combine(exprs []Expression) chan Expression {
	answer := make(chan Expression)
	go func() {
		defer close(answer)
		if size := len(exprs); size == 1 {
			answer <- exprs[0]
		} else {
			for i := 1; i < size; i++ {
				for left := range combine(exprs[:i]) {
					for right := range combine(exprs[i:]) {
						for _, comb := range combiners {
							if expr, ok := comb(left, right); ok {
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

func findBest(target int, exprs chan Expression) chan Expression {
	answer := make(chan Expression)
	differenceFromTarget := differenceFrom(target)
	go func() {
		defer close(answer)
		bestSoFar, bestDiff := Expression{}, 11
		for e := range exprs {
			switch diff := differenceFromTarget(e); {
			case diff > 10 || diff > bestDiff:
			case diff == bestDiff && e.numberCount >= bestSoFar.numberCount:
			default:
				answer <- e
				bestSoFar, bestDiff = e, diff
			}
		}
	}()
	return answer
}

func differenceFrom(target int) func(Expression) int {
	return func(e Expression) int {
		if diff := target - e.value; diff < 0 {
			return diff * -1
		} else {
			return diff
		}
	}
}
