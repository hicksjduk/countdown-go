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