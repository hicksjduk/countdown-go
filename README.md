# countdown-go

A Go implementation of a solver for the Countdown numbers game.

The game involves making an arithmetic expression using some or all of the six
given numbers, together with the basic arithmetic operators (add, subtract, 
multiply, divide) and as many pairs of parentheses as necessary, in order to get
as close to the given target number as possible.

The target number has three digits (that is, it is in the range 100 to 999
inclusive). The numbers for use are taken from the 'big' numbers 25, 50, 75 and
100 (up to one occurrence of each) and the 'small' numbers 1 to 10 inclusive
(up to two occurrences of each).

In order to be a valid solution, an expression must differ from the target
number by 10 or less. One solution is better than another if it differs by less
from the target number by less. This solver also treats one solution as better
than another if it differs from the target number by the same amount, but uses
fewer numbers (although this is not part of the game).

Typically, the solver outputs multiple expressions, showing a progression
towards the optimum solution. It outputs the first expression it finds that
differs from the target number by 10 or less, then each expression that is found
to be a better solution than the previous one.

## Some notes on implementation

There are three main stages in the solution, each of which builds on the previous
one:

1. Make all permutations of the input numbers - that is, all possible unique
sequences that contain one or more of the numbers. Order is significant when it comes
to uniqueness - two sequences that contain the same numbers, but in a different
order, are unique.

1. For each permutation, make all the expressions that can be derived from the
numbers, in the order given, using the four arithmetic operators and as many pairs of
parentheses as necessary to adjust the order of evaluation.

1. Evaluate each expression against the best solution found so far, and if it is
better output it, and retain it as the new best solution.

### Generators

The solver makes extensive use of a "generator" pattern, to generate and 
process potentially very large sets of data in an efficient way. The term "generator"
is derived from Python, where most iterations are done using generators and there is
explicit language support for creating them. 

The difference between a generator and more traditional ways of generating data sets
(returning arrays or collections) is that each item of data returned by a generator
is created on demand, when it is needed. This avoids both the up-front cost of 
generating the data before it is used, and the memory footprint consumed by the
structure that contains it.

There is no explicit language support for generators in Go, but they can easily be
simulated using channels and goroutines. A typical pattern for a generator function 
is:

* The function takes whatever parameters are needed to generate its output, and
returns a channel of the appropriate output type.
* The body of the function:
   * Creates the channel through which the output is to be returned. The channel is
     typically unbuffered.
   * Runs a goroutine which generates the output items and puts them in the channel,
     and then (crucially) closes the channel.
    * Returns the channel.

This behaves as a generator, generating data values only on demand, because 
when an attempt is made to insert a value in a channel, the attempt
blocks if the channel is unbuffered, or its buffer is full, and nothing is waiting
to retrieve a value from the channel. This block remains until a retrieval request 
is made.

An example of a generator function that returns a potentially very large sequence
of integers is:

```golang
func int_sequence_generator(max int) chan int {
    answer := make(chan int)
    go func() {
        defer close(answer)
        for i := 1; i <= max; i++ {
            answer <- i
        }
    }()
    return answer
}
```

This can be invoked, and its results consumed, using code similar to the
following, which ranges over the channel returned by the generator:

```golang
for i := range big_sequence_generator(10000000000) {
	// process the data item
}
```

## Testing notes

The tests in this repository are BDD (behaviour-driven design) tests written in
[the Gherkin language](https://cucumber.io/docs/gherkin/reference/). 
These tests are run using the [godog](https://github.com/cucumber/godog) package.

The simplest steps (followed in this repo) to get BDD-style tests discovered and
run using godog are:
* In the package where the code resides that you want to test, add:
   * a subdirectory called `features` which contains your Gherkin feature file(s)
   * a file whose name ends with `_test.go` which contains your test definitions.
* In the root of your project, run the command `go mod init`. This is required
because test discovery and running only works properly in a Go module.

Godog tests can be integrated with the Go test infrastructure using the `TestMain`
function which is supported from version 1.14 onwards. The `countdown_test_go`
file in this repo contains the following functions:
* `TestMain`, which creates and runs a `godog.TestSuite` that is initialised using
the `InitializeScenario` function.
* A number of functions that implement the various test steps. Each of these functions
returns an `error` if any expectations are not met, otherwise it returns `nil`.
* `InitializeScenario`, which associates the test step implementation functions
with the regular expressions that are used to match step definitions
in the feature file. Each association is defined by calling `godog.Step`, passing
a regular expression and a reference to a step implementation function.

The easiest way to pass data from one step of a test to another is to define a type
which contains the data that needs to be passed, and then to make all
the step implementation functions (or at least those that need to pass or receive
data) methods of that type. Then, in the `InitializeScenario` function, an instance
of that type is created, and each step implementation is defined as a method of that
instance.

For example, in this test suite all the step implementations are methods of type `testContext`
which is a struct containing a number of fields. Then the `InitializeScenario`
function is defined thus:
```go
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
```
The first line of this function creates an instance of `testContext`, and the remaining lines
associate its methods with the regular expressions used in parsing the feature file.