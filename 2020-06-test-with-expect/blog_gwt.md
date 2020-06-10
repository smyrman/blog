# Test-With-Expect: A BDD-style Go naming pattern

_TL;DR: This article demonstrate how to write GWT-inspired tests in plain Go, and how to name them. Skip to the Go TWE heading to see the result._

[GWT][gwt], or "Given-When-Then", is a great naming convention for tests that comes from the [BDD][bdd], or "Behavior-Driven-Development" paradigm. It makes it easy to _plan_ tests as well as the _behavior_ of your feature before you start the detailed implementation.

GWT is composed of three concepts or steps:

- Given: A precondition for the test (context)
- When: The action to perform (action)
- Then: A result to expect (check)

Each of these steps can be nested, or sometimes skipped.

The challenge of this article is not to write GWT or BDD style tests in Go, this has been [demonstrated][gwt-demo-1] many times [before][gwt-demo-2], but an exploration into how we can do this without a third-party test framework and DSL. There are also some benefits associated with relying on the the default test-runner as well as good-old (yes they are 4 years now!) [subtests][go-subtests] that we will look into.

## Code under test

In this article, we will imagine that we are going to write a generic `Sum` function. Generic is a loaded term in the programming world. In our case, we plan to write a function that accepts any slice or array of numeric values and returns the sum. Because _type parameterization_ is not (yet) possible in Go, we will allow the function to return an error when it receives invalid input.

Let's define the interface we want this function to have:

```go
package mypkg

import (
)

// Sum accepts any kind of slice or array holding only
// numeric values, and returns the sum.
func Sum(v interface{}) (float64, error) {
    return -1, errors.New("NYI")
}
```

Following the sprit of BDD and TDD, we will wait with the actual implementation until _after_ we are done with the tests. In fact, since this article is about writing and naming tests, we will leave the _entire implementation_ as an exercise for the reader.

## Planning our tests

For the scope of this article, let's hash out how we want this function to behave for **integer slice** (`[]int`) input in particular. By the power of GWT, we plan our tests in plain text first:

```txt
TestSum:
  Given a non-empty int slice S:
    When calling Sum(S):
      Then it should not fail
      Then it should return the correct sum
      Then S should be unchanged

  Given an empty int slice S:
    When calling Sum(S):
      Then it should not fail
      Then it should return zero
      Then S should be unchanged
```

Great, we have specified our tests, but how does it look once we turn it into Go code? And what kind of output can we expect?

## Writing the tests in Go

Some popular BDD frameworks, such as [Ginko][ginko], define their own test runner and re-implement Go sub-tests (although it predates them, to be fair) by structuring their library to implement a form of DSL (Domain Specific Language). This framework can render pretty output for tests, especially if run in a a terminal that supports color. However, if you wish to focus on a failing sub-test, you can not rely on `go test -run` to do it; you need to do it the "Ginko way". Because of this, you can also not rely on editor or IDE integrations in the same way you can with other Go tests.

Contrary to popular belief however, it is actually possible to write GWT-style tests in Go _without_ using a BDD-style framework or DSL; GWT itself is just a naming convention, and we can use it with normal Go sub-tests. The most importantly benefit of doing this, is that your GWT-style tests will behave **consistently to other Go tests**, and can thus be **treated equally** by both humans and tools. This means you can spend less time in training humans, CIs, JUnit parsers, IDEs, etc., and more time _writing tests_. Especially so if your team is Go-centric anyways.

So let's write it!

```go
func TestSum(t *testing.T) {
	t.Run("Given a non-empty int-slice S", func(t *testing.T) {
		s := []int{1, 2, 3}
		t.Run("When calling Sum(S)", func(t *testing.T) {
			i, err := mypkg.Sum(s)
			t.Run("Then it should not fail", subtest.Value(err).NoError())
			t.Run("Then it should return the correct sum", subtest.Value(i).NumericEqual(6))
			t.Run("Then S should be unchanged", subtest.Value(s).DeepEqual([]int{1, 2, 3}))
		})
	})
	t.Run("Given an empty int-slice S", func(t *testing.T) {
		s := []int{}
		t.Run("When calling Sum(S)", func(t *testing.T) {
			i, err := mypkg.Sum(s)
			t.Run("Then it should not fail", subtest.Value(err).NoError())
			t.Run("Then it should return zero", subtest.Value(i).NumericEqual(0))
			t.Run("Then S should be unchanged", subtest.Value(s).DeepEqual([]int{}))
		})
	})
}
```

To shorten the implementation somewhat, we are using our own **experimental** matching library [searis/subtest][subtest]. Subtest works like other matcher libraries, but instead of taking `t *testing.T` as a parameter, like [testify/assert][testify-assert], or initialize a matcher instance, like [Gomega can do][gomega-xunit], it returns a test-function (a.k.a. sub-test) that we can pass directly to `t.Run`. A convenient side-effect of this is that `subtest` allows focus on _individual checks_ via `go test -run '^ParentTestName/SubTestName/CheckName$'`.

PS! I want to underline that I don't view a matcher library to be a test framework. `library != framework`. If you are not convinced, it is possible to write tests without a matcher library as well, and in fact, this what the Go team does. It's not _difficult_ to do the checks, and there are good arguments for doing the checks yourself, but it requires a very high discipline and use of boiler-plate to ensure consistently styled failure output.

But we are diverging... The code reads well now, but **there is a problem**!

## Long test-names and duplicated information

These are the full test names that was generated by our code above:

```txt
TestSum/Given_a_non-empty_int-slice_S/When_calling_Sum(S)/Then_it_should_not_fail
TestSum/Given_a_non-empty_int-slice_S/When_calling_Sum(S)/Then_it_should_return_the_correct_sum
TestSum/Given_a_non-empty_int-slice_S/When_calling_Sum(S)/Then_S_should_be_unchanged
TestSum/Given_an_empty_int-slice_S/When_calling_Sum(S)/Then_it_should_not_fail
TestSum/Given_an_empty_int-slice_S/When_calling_Sum(S)/Then_it_should_return_zero
TestSum/Given_an_empty_int-slice_S/When_calling_Sum(S)/Then_S_should_be_unchanged
```

Once you mange to grok them, they make sene, but they are _long_, and it _stutters_. If you find the names _themselves_ intimidating, try scanning them _quickly_ from the `go test` output:

```txt
$ go test github.com/smyrman/blog/2020-06-test-with-expect/mypkg -run ^(TestSum)$
--- FAIL: TestSum (0.00s)
    --- FAIL: TestSum/Given_a_non-empty_int-slice_S (0.00s)
        --- FAIL: TestSum/Given_a_non-empty_int-slice_S/When_calling_Sum(S) (0.00s)
            --- FAIL: TestSum/Given_a_non-empty_int-slice_S/When_calling_Sum(S)/Then_it_should_not_fail (0.00s)
                sum_gwt_test.go:16: error is not nil
                    got: *errors.errorString
                        NYI
            --- FAIL: TestSum/Given_a_non-empty_int-slice_S/When_calling_Sum(S)/Then_it_should_return_the_correct_sum (0.00s)
                sum_gwt_test.go:17: not numeric equal
                    got: float64
                        -1
                    want: float64
                        6
    --- FAIL: TestSum/Given_an_empty_int-slice_S (0.00s)
        --- FAIL: TestSum/Given_an_empty_int-slice_S/When_calling_Sum(S) (0.00s)
            --- FAIL: TestSum/Given_an_empty_int-slice_S/When_calling_Sum(S)/Then_it_should_not_fail (0.00s)
                sum_gwt_test.go:25: error is not nil
                    got: *errors.errorString
                        NYI
            --- FAIL: TestSum/Given_an_empty_int-slice_S/When_calling_Sum(S)/Then_it_should_return_zero (0.00s)
                sum_gwt_test.go:26: not numeric equal
                    got: float64
                        -1
                    want: float64
                        0
FAIL
FAIL	github.com/smyrman/blog/2020-06-test-with-expect/mypkg	0.350s
FAIL
```

To me, one of the most obvious problem with the names is that space (` `) is replaced by underscore (`_`); this makes them hard to scan. The default go test runner also _repeats_ the parent names when printing the sub-test name, contributing to the problem. But perhaps the biggest problem, is actually that the names themselves are _too long_. In fact, the names themselves are duplicating information. In particular `"When_calling_Sum(S)"` is information that can already by understood by reading the top-level test-name `TestSum`. We are testing Sum -- how can we do that without calling it?

Naming a test after the type, function or method that is under test is a pretty common Go convention for unit-tests. And even we are writing this test as BDD, this particular test _is_ a unit-test. If we can keep following this convention, it does makes the test behave _more_ like other Go tests.

## Test-With-Expect

The fundamental concepts that GWT offers, are pretty cool, but the words themselves -- Given, When, Then -- is actually less important. We will look at an alternative wording that fit better for Go in particular, but it can of-course apply elsewhere.

Another aspect wi will attack, is that GWT names are written to be human readable, and as they form near complete "English-like" sentences using what BDD-guys call _natural language_, they are also _naturally_ long. If there is _one_ idiom that is important in Go though, it is that _names_ should be short and precise, rather then long and windy. That's not my words. Here is an extract from Rus Cox's famous [naming philosophy][rsc-quote] quote:

> A name's length should not exceed its information content. (...) Global names must convey relatively more information, because they appear in a larger variety of contexts. Even so, a short, precise name can say more than a long-winded one: compare acquire and take_ownership. Make every name tell.

Other advice and information that relate to Go names, include:

- Names in Go have [semantic effect](https://golang.org/doc/effective_go.html#names).
- Avoid redundancy in names, E.g. package names + global names.
- Short and concise is more important than grammatical correctness. E.g. a constant named `StatusFail` read just as well as `StatusFailure`.

The first restriction to note here, is that names have semantic effect. Relevant to our case, all test function names in Go _have_ to start with the word `Test`. Taking the consequence of this, we might as well include that as our first word. The next two words follow relatively naturally from that restriction:

- Test: Type or function to test (subject).
- With: Configuration or input that are some-how passed to subject (configuration)
- Expect: What to expect afterwards (check).

To sum it up (no pun intended), these are our new test-names:

```txt
TestSum/With_non-empty_int_slice/Expect_no_error
TestSum/With_non-empty_int_slice/Expect_correct_sum
TestSum/With_non-empty_int_slice/Expect_input_is_unchanged
TestSum/With_empty_int_slice/Expect_no_error
TestSum/With_empty_int_slice/Expect_zero
TestSum/With_empty_int_slice/Expect_input_is_unchanged
```

Notice that we write just `Expect_correct_sum` over the more _correct_ `Expect_the_correct_sum_to_be_returned`. or the previous `Then_it_should_return_the_correct_sum`. This is just an application of the Go naming philosophy to BDD-style _natural language_. Keep it Short, Precise, and Happily Sacrifice Some English Grammar -- or KISPAHSSEG to make an abbreviation that you wil defiantly remember!

PS! KISPAHSSEG, is a very inclusive version of English, especially for the non-native speaker.

## Go TWE

Finally, here is the code for our tests in Test-With-Expect format:

```go
package mypkg_test

import (
  "github.com/smyrman/mypkg"
)

func TestSum(t *testing.T) {
	t.Run("With non-empty int slice", func(t *testing.T) {
		s := []int{1, 2, 3}
		i, err := mypkg.Sum(s)
		t.Run("Expect no error", subtest.Value(err).NoError())
		t.Run("Expect correct sum", subtest.Value(i).NumericEqual(6))
		t.Run("Expect input is unchanged", subtest.Value(s).DeepEqual([]int{1, 2, 3}))
	}
	t.Run("With empty int slice", func(t *testing.T) {
		s := []int{}
		i, err := mypkg.Sum(s)
		t.Run("Expect no error", subtest.Value(err).NoError())
		t.Run("Expect zero", subtest.Value(i).NumericEqual(0))
		t.Run("Expect input is unchanged", subtest.Value(s).DeepEqual([]int{}))
	}
}
```

And the co-responding test output:

```txt
$ go test github.com/smyrman/blog/2020-06-test-with-expect/mypkg -run ^(TestSum)$
--- FAIL: TestSum (0.00s)
    --- FAIL: TestSum/With_non-empty_int_slice (0.00s)
        --- FAIL: TestSum/With_non-empty_int_slice/Expect_no_error (0.00s)
            sum_twe_test.go:14: error is not nil
                got: *errors.errorString
                    NYI
        --- FAIL: TestSum/With_non-empty_int_slice/Expect_correct_sum (0.00s)
            sum_twe_test.go:15: not numeric equal
                got: float64
                    -1
                want: float64
                    6
    --- FAIL: TestSum/With_empty_int_slice (0.00s)
        --- FAIL: TestSum/With_empty_int_slice/Expect_no_error (0.00s)
            sum_twe_test.go:21: error is not nil
                got: *errors.errorString
                    NYI
        --- FAIL: TestSum/With_empty_int_slice/Expect_zero (0.00s)
            sum_twe_test.go:22: not numeric equal
                got: float64
                    -1
                want: float64
                    0
FAIL
FAIL	github.com/smyrman/blog/2020-06-test-with-expect/mypkg	0.064s
FAIL
```

## Extensions

Test-With-Expect is our _base_, but these three words are not always enough. Maybe you need more words? One additional word I have used myself, is `After`. It is not useful in this example, but what if you are writing tests that utilize a for-loop and do a check for each iteration? Other starting words could be called for in specific contexts, but be sure to limit the number and usage to ensure consistency.

```txt
TestX/With_Y/After_N_repetitions/Expect_no_error.
```

You can also include information for failing tests _without_ putting information in the name; just make a call to `t.Log`/`t.Logf`. By default, this output only appear if your (sub-)test actually fails.

The following code shows both a log statement and a useful setup/teardown pattern:

```go
func TestResourceFind(t *testing.T) {
    setup := func(t *testing.T, cnt int) (r Resource, teardown func()) {
        t.Logf("Resource R set-up with %d records", cnt)
        // setup r with records ...
        return r
    }
    t.Run("With Query={Limit:5,Offset:32}", func(t *testing.T) {
        r, teardown := setup(t, 1000)
        defer teardown()
        // test r.Find ...
    })
    // ...
}
```

## Improving test-runner output

So to the problems that we have not resolved. Can we print space instead of underscore? Can we avoid printing the parent names? Can we add color?

To answer all of those question at once, I would like to quote Bob the Builder:

> Can we fix it? Yes we can!

We can even fix it without writing or using a separate _test-runner_; All we need is a separate _test-formatter_ that can handle the output of `go test -json`. But this is an exercise for another blog post.

Besides, having the full test name printed in-tact _does_ hold value; it can be copy-pasted and inserted into `go test -run` to _focus_ on individual tests or groups of tests. Perhaps in the future our IDEs and editors can also insert links to re-run them?

## Conclusion

As for a final conclusion, I want to leave this up to the reader. Do _you_ think Given-When-Expect is a meaningful way to name sub-tests in Go? Would you do it yourself, or would you prefer the DSL from one of the main frameworks?

[go-subtests]: https://blog.golang.org/subtests
[gwt]: https://www.agilealliance.org/glossary/gwt/
[bdd]: https://www.agilealliance.org/glossary/gwt/
[gomega]: http://onsi.github.io/gomega/
[gomega-xunit]: https://onsi.github.io/gomega/#using-gomega-with-golangs-xunit-style-tests
[ginko]: http://onsi.github.io/ginkgo/
[go-convey]: http://goconvey.co
[godog]: https://github.com/cucumber/godog
[gospec]: https://github.com/luontola/gospec
[testify-assert]: https://pkg.go.dev/github.com/stretchr/testify/assert?tab=doc
[go-names]: https://www.reddit.com/r/golang/comments/8wxwgv/why_does_go_encourage_short_variable_names/
[gwt-demo-1]: https://rollout.io/blog/implementing-a-bdd-workflow-in-go/
[gwt-demo-2]: https://semaphoreci.com/community/tutorials/how-to-use-godog-for-behavior-driven-development-in-go
[subtest]: https://github.com/searis/subtest
[rsc-quote]: https://research.swtch.com/names
