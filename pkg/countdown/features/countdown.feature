Feature: Countdown numbers solver

  Scenario Outline: Exact solution found
    When I call the solver with target number <target> and numbers <numbers>
    Then a solution is found whose value equals the target number and which uses <count> numbers

  Examples:
    | target | numbers        | count |
    | 834    | 10,9,8,7,6,5   | 5     |
    | 378    | 50,7,4,3,2,1   | 3     |
    | 493    | 50,25,4,3,2,4  | 6     |
    | 803    | 50,4,9,6,6,1   | 6     |
    | 827    | 25,8,5,8,1,2   | 6     |

  Scenario: Non-exact solution found
    When I call the solver with target number 954 and numbers 50, 75, 25, 100, 5, 8
    Then a solution is found whose value equals 955 and which uses 5 numbers

  Scenario: No solution found
    When I call the solver with target number 999 and numbers 1,2,3,4,5,6
    Then no solution is found
