package countdown

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func Solve(target int, numbers ...int) *Expression {
	fmt.Println("-----------------------------------")
	fmt.Printf("Target: %v, numbers: %v\n", target, numbers)
	answer, time := doTimed(func() *Expression {
		return solve(target, numbers)
	})
	fmt.Printf("Finished in %vms\n", time)
	if answer == nil {
		fmt.Println("No result found")
	}
	fmt.Println("-----------------------------------")
	return answer
}

func doTimed(f func() *Expression) (*Expression, int64) {
	start := currentTimeMillis()
	answer := f()
	end := currentTimeMillis()
	return answer, end - start
}

func currentTimeMillis() int64 {
	return int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond)
}

var processCount = runtime.GOMAXPROCS(0)

func solve(target int, numbers []int) *Expression {
	exprs := make([]*Expression, len(numbers))
	for i, v := range numbers {
		exprs[i] = numberExpression(v)
	}
	var answer *Expression = nil
	for expr := range solveIt(target, exprs) {
		fmt.Printf("%v\n", expr)
		answer = expr
	}
	return answer
}

func solveIt(target int, numbers []*Expression) chan *Expression {
	combs := combinations(numbers, target)
	evaluate := evaluator(target)
	accumulator := make(chan *Expression, processCount)
	wg := sync.WaitGroup{}
	wg.Add(processCount)
	for i := processCount; i > 0; i-- {
		go func() {
			defer wg.Done()
			evaluate(combs, accumulator)
		}()
	}
	answer := make(chan *Expression, processCount)
	go func() {
		defer close(answer)
		evaluate(accumulator, answer)
	}()
	go func() {
		wg.Wait()
		close(accumulator)
	}()
	return answer
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
	addition = &operation{
		symbol:      "+",
		priority:    low,
		commutative: true,
		evaluator: func(a, b int) int {
			return a + b
		},
	}
	subtraction = &operation{
		symbol:      "-",
		priority:    low,
		commutative: false,
		evaluator: func(a, b int) int {
			return a - b
		},
	}
	multiplication = &operation{
		symbol:      "*",
		priority:    high,
		commutative: true,
		evaluator: func(a, b int) int {
			return a * b
		},
	}
	division = &operation{
		symbol:      "/",
		priority:    high,
		commutative: false,
		evaluator: func(a, b int) int {
			return a / b
		},
	}
)

type Expression struct {
	value       int
	print       func() string
	priority    Priority
	numbers     []int
	parentheses int
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
		priority:    atomic,
		numbers:     []int{n},
		parentheses: 0,
	}
}

func arithmeticExpression(leftOperand *Expression, operator *operation, rightOperand *Expression) *Expression {
	answer := &Expression{
		value: operator.evaluator(leftOperand.value, rightOperand.value),
		print: func() string {
			return printExpression(leftOperand, operator, rightOperand)
		},
		priority:    operator.priority,
		numbers:     append(append([]int{}, leftOperand.numbers...), rightOperand.numbers...),
		parentheses: leftOperand.parentheses + rightOperand.parentheses,
	}
	if parenthesiseLeft(operator, leftOperand) {
		answer.parentheses++
	}
	if parenthesiseRight(operator, rightOperand) {
		answer.parentheses++
	}
	return answer
}

func printExpression(leftOperand *Expression, operator *operation, rightOperand *Expression) string {
	return fmt.Sprintf("%v %v %v",
		withParensIfNecessary(leftOperand.print(), parenthesiseLeft(operator, leftOperand)),
		operator.symbol,
		withParensIfNecessary(rightOperand.print(), parenthesiseRight(operator, rightOperand)))
}

func parenthesiseLeft(operator *operation, leftOperand *Expression) bool {
	return leftOperand.priority < operator.priority
}

func parenthesiseRight(operator *operation, rightOperand *Expression) bool {
	return rightOperand.priority < operator.priority ||
		(rightOperand.priority == operator.priority && !operator.commutative)
}

func withParensIfNecessary(s string, parensNeeded bool) string {
	if parensNeeded {
		return fmt.Sprintf("(%v)", s)
	}
	return s
}

type combiner func(*Expression) *Expression

type combinerCreator func(*Expression) combiner

func addCombinerCreator(left *Expression) combiner {
	return func(right *Expression) *Expression {
		return arithmeticExpression(left, addition, right)
	}
}

func subtractCombinerCreator(left *Expression) combiner {
	if left.value < 3 {
		return nil
	}
	return func(right *Expression) *Expression {
		if left.value <= right.value || left.value == right.value*2 {
			return nil
		}
		return arithmeticExpression(left, subtraction, right)
	}
}

func multiplyCombinerCreator(left *Expression) combiner {
	if left.value == 1 {
		return nil
	}
	return func(right *Expression) *Expression {
		if right.value == 1 {
			return nil
		}
		return arithmeticExpression(left, multiplication, right)
	}
}

func divideCombinerCreator(left *Expression) combiner {
	if left.value == 1 {
		return nil
	}
	return func(right *Expression) *Expression {
		if right.value == 1 || left.value%right.value != 0 || left.value == right.value*right.value {
			return nil
		}
		return arithmeticExpression(left, division, right)
	}
}

var combinerCreators = []combinerCreator{
	addCombinerCreator,
	subtractCombinerCreator,
	multiplyCombinerCreator,
	divideCombinerCreator,
}

func combinations(exprs []*Expression, target int) chan *Expression {
	answer := make(chan *Expression, processCount)
	diff := differenceFrom(target)
	go func() {
		defer close(answer)
		for p := range permute(exprs) {
			for c := range combine(p) {
				if diff(c) <= 10 {
					answer <- c
				}
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
			used := usedTracker[int]()
			for i, expr := range exprs {
				if !used(expr.value) {
					answer <- []*Expression{exprs[i]}
					others := make([]*Expression, 0, count-1)
					for _, subArray := range [][]*Expression{exprs[:i], exprs[i+1:]} {
						others = append(others, subArray...)
					}
					for perm := range permute(others) {
						answer <- append([]*Expression{exprs[i]}, perm...)
					}
				}
			}
		}
	}()
	return answer
}

func usedTracker[K comparable]() func(K) bool {
	used := map[K]interface{}{}
	return func(v K) bool {
		_, answer := used[v]
		if !answer {
			used[v] = nil
		}
		return answer
	}
}

func combinersUsing(left *Expression) []combiner {
	answer := make([]combiner, 0, len(combinerCreators))
	for _, cc := range combinerCreators {
		if c := cc(left); c != nil {
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
			for i := 1; i < size; i++ {
				for left := range combine(exprs[:i]) {
					combiners := combinersUsing(left)
					for right := range combine(exprs[i:]) {
						for _, comb := range combiners {
							if expr := comb(right); expr != nil {
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

func evaluator(target int) func(chan *Expression, chan *Expression) {
	compare := better(target)
	return func(exprs, output chan *Expression) {
		var bestSoFar *Expression
		for e := range exprs {
			if bestSoFar == nil || compare(bestSoFar, e) == e {
				bestSoFar = e
				output <- e
			}
		}
	}
}

func better(target int) func(*Expression, *Expression) *Expression {
	extractors := []func(*Expression) int{
		differenceFrom(target),
		func(e *Expression) int {
			return len(e.numbers)
		},
		func(e *Expression) int {
			return e.parentheses
		},
	}
	return func(e1, e2 *Expression) *Expression {
		for _, f := range extractors {
			v1, v2 := f(e1), f(e2)
			if v1 < v2 {
				return e1
			}
			if v1 > v2 {
				return e2
			}
		}
		return e1
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
