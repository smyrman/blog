# Go generics beyond the playground [DRAFT]

While Go as of version 1.16 does not support Generics, there is an accepted [language proposal][lang-prop] for it, and we can _probably_ expect it to arrive in Go 1.18 or later. But as it turns out, we don't need to wait for it to start experimenting with the new feature. We can start _now_ to write small snippets of code on the [go2go playground][go2go-play], or even full Go packages that you [develop locally][go2go-readme].

By far the easiest way to test out Go generics, is to use the playground. And there is nobody saying that the playground isn't awesome. However awesome though, there is clear limits to how much you can reasonably try out in the playground alone. What if you have enough code to start splitting into files? What if you want to write _unit-tests_? How would a full package look like with generics? In my view, the _best_ way to try out a new feature, is to actually do something useful. And to do this with generics, we need to venture out of the safety and comfort of the playground.

> My hope is that this will _inspire_ you to do your own experiments with Go generics _beyond the playground_ and write something _potentially useful_. Only then, can we truly see if generics itself, is going to be useful in Go.

In this article, I will go through how I [re-wrote][subx] a [test matching library][subtest] from scratch with generics as part of the tool-box. My hope is that this will _inspire_ you to do your own experiments with Go generics _beyond the playground_ and write something _potentially useful_. Only then, can we truly see if generics itself, is going to be useful in Go. If you want, you could use the library in this article for _testing_ your experiments; or you can extend the library and do a pull request.

Before we start: a warning. I will assume basic knowledge about the generics proposal in Go and what the most important design concepts are. If you don't have this knowledge, I would recommend that you first acquire it. E.g. by reading [Go 2 generics in 5 minutes][go-generics-5min], by reading the [updated design draft][design-draft], or by doing your own experiments in the [go2go playground][go2go-play] first. In the meantime, keep this article on your TO-READ list. With that said, because we are trying to do something _potentially useful_, this article will be more about package design then about generics itself. This is the _side-effect_ of simply adding generics to our toolbox, rather then trying to write everything with generics (like we did with channels back before the release of Go v1). Be reassured though, generics will definitely still be an _essential_ part of our (re-)design.

## A problem to solve (again) with generics

Starting out with Go generics, in order to do something _useful_, we need a problem to solve. The problem I have picked for this article is one that I have tried to solve before when designing the [test matcher/assertion library][subtest] that we use to test the [Clarify][clarify] back-end at [Searis][searis]. But first, you probably have a question: with all the great test matcher libraries we have in Go, why on earth would we want to write a new one? To answer that, it's worth having a closer look at at one of the existing matcher libraries. Does it solve it's _mission_ in a useful way, and with the best possible package design?

In the Go world, by far the most popular Go matcher library still appear to be the `assert` package from the [stretchr/testify][testify] repo. It's an old, good and stable library. However, because it's old, and because most (old) Go libraries strived to keep full backwards compatibility, it's also an excellent library to demonstrate some _problems_ that can perhaps be solved by a different design. For this article, we will be content with considering the following code, testing an imaginary function `mypkg.Sum`:

```go
func TestSum(t *testing.T) {
	a := mypkg.Vector{1,0,3}
	b := mypkg.Vector{0,1,-2}
	expect := mypkg.Vector{1,1,1}

	result, err := mypkg.Sum(a, b)

	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, result, expect, "unexpected result")
}
```

At first glance, all of this might look fine. What's wrong with it, you might think. Before we get into that, there is one minor alteration I want to do to the code, and you would probably know this if you have ever used the `assert` library. Because the descriptive text in the assertion is optional, you would _probably_ write them like this:

```go
assert.NoError(t, err)
assert.Equal(t, result, expect)
```

Which may give the following output in the case of a failure:

```txt
--- FAIL: TestSum_assert (0.00s)
    /Users/smyrman/Code/blog/2021-03-generics-beyond-the-playground/mypkg/subtest_sum_test.go:35:
        	Error Trace:	subtest_sum_test.go:35
        	Error:      	Not equal:
        	            	expected: mypkg.Vector{0, 1, -2}
        	            	actual  : mypkg.Vector{1, 1, 1}

        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1,5 +1,5 @@
        	            	 (mypkg.Vector) (len=3) {
        	            	- (float64) 0,
        	            	  (float64) 1,
        	            	- (float64) -2
        	            	+ (float64) 1,
        	            	+ (float64) 1
        	            	 }
        	Test:       	TestSum_assert
FAIL
FAIL	github.com/smyrman/blog/2021-03-generics-beyond-the-playground/mypkg	0.130s
FAIL
```

So, to the problems. Problems in design are often subtle ones. The ones you can't quite put your finger on. Once you know about them though, they are hard not to se. In this short snippet of code and failure output, there is actually as much as _six problems_ that I want to bring to attention:

1. We have to pass in the `t` parameter. It's a minor inconvenience, but enough for Dave Cheney to write a quite [interesting package][pkg-expect] for solving it. Not recommended for production use, I might add.
2. The "got" and "want" parameter order is hard to get right. If you are observant, you might have noticed that the code above gets it _wrong_. This again makes the failure output tell a _lie_. In my own experience, such lies can lead to significant confusion when debugging a problem. As a rule of thumb if you use the `assert` package a lot, and tend to continue doing so, the "expected" parameter almost always come _first_. This is of course not a general rule in Go. E.g. the standard library tend to report the "got" parameter first, and the "want" parameter second.
3. From simply _starring_ at the code (which is what code reviewers generally do) it's not obviously clear what's meant by _Equal_. It's well enough _documented_, but from the name alone, it's not clear if it's going to use the equal comparison (`==`), the `reflect.DeepEqual` method, or if it can handel equal comparison of two `time.Time` instances with a different time-zone. At some point, subtile details in how equality is implemented, might come back and bite us. Especially so if any of these details don't match the comparison we _usually do_ in our programs.
4. The descriptive text for what went wrong is _optional_; and thus easily omitted. This can make debugging failing tests hard, as there is little information to exactly _what_ went wrong (other than two values did not compare equal). Especially so if we have multiple assert statements. Is it the result from the function under test that's wrong? Is this just a sanity check before running the main test?
5. For what the output potentially lack of _useful_ information, it makes up for in _redundant_ information. Why do we need a diff for finding the difference between these simple struct instances? Given the diff is printed, why do we also get the "actual" and "expected" one-liners? Why is the filename and line trace information repeated? Why is the _test name_ repeated?
6. Many of the library assertion, `assert.DeepEqual` included, lack compile-time type safety. Of course, one might say, but if type-safety is beneficial in the _rest_ of our programs, would it not also be beneficial in tests?

In our test-matcher for Clarify (called `subtest`), we have managed to to solve many of these problems. In order to explain how, let's start by rewriting the test function for `mypkg.Sum`:

```go
func TestSum(t *testing.T) {
	a := mypkg.Vector{1, 0, 3}
	b := mypkg.Vector{0, 1, -2}
	expect := mypkg.Vector{1, 1, 1}

	result, err := mypkg.Sum(a, b)

	t.Run("Expect no error", subtest.Value(err).NoError())
	t.Run("Expect correct sum", subtest.Value(result).DeepEqual(expect))
}
```

Which in case of a failing test, may give the output:

```txt
--- FAIL: TestSum (0.00s)
    --- FAIL: TestSum/Expect_correct_sum (0.00s)
        /Users/smyrman/Code/blog/2021-03-generics-beyond-the-playground/mypkg/subtest_sum_test.go:46: not deep equal
            got: mypkg.Vector
                [0 1 -2]
            want: mypkg.Vector
                [1 1 1]
FAIL
FAIL	github.com/smyrman/blog/2021-03-generics-beyond-the-playground/mypkg	0.106s
FAIL
```

So, let's go through which problems we managed to solve by this, and which one we didn't.

First out, you might note that we do not pass in a `t` parameter. That is because what `subtest` does, is to _return a test function_. This test function can then be run as a Go _sub-test_, omitting the need for the user to pass in a `t` parameter. Cleaver? I certainly thought so when I first had the idea.

Next up we solved the ordering issue of "got" and "want" by wrapping them in each their method call. We can see that we first call `Value` on the result. This actually return what we call a `ValueFunc` or _value initializer_. It is not important now to explain why it's a function, but we can perhaps get back to that. The important part now is that this type has _methods_ that let's us initialize tests. In this particular instance, we call the method `DeepEqual` with the expect parameter. Here we have also attempted to solve the clarity issue of what `Equal` does, by giving the method a more descriptive name.

Next up, as we pass the `func(*testing.T)` instance returned by `DeepEqual` to `t.Run`, this method require us to supply a _name_. Thus a description for the check is required. In fact, as we are using Go sub-tests, there is now even a way for us to re-run a particular _check_ in the test for better focus during debugging. This can be particular useful for cases when there are many failing sub-tests and many failing checks in a test.

```sh
test -name '^TestSum/Expect_correct_sum'
```

Next up, we kept our test-output _brief_ without repeating any information. We should note that there definitively is some cases where we would not get _enough_ information from this brief output format. While the library do allow some output customization, this is one of those trade-offs that could probably be improved.

As for the final problem though, compile time type-safety, `subtest` falls short. We can still pass pretty much anything into the `subtest.Value` and various test initializer methods, and it won't detect any type mismatches for us before the test run. Besides, does it really make sense to have the same test initializer methods available for all Go types? Should not a `subtest.Value(time.Time{})` provide methods that are useful for comparing times, such as analogies to the `time.Time` methods `.Equal`, `.Before` and `Time{}.After`?

If we do the count, we gather that `subtest` appear to solve five out of the six problems we identified with the assert library. At this point though, it's important to note that at the time when the `assert` package was designed, the sub-test feature in Go did not yet exist. Therefore it would have been impossible for that library to embed it into it's design. This is also true for when [Gomega][gomega] and [Ginko][ginko] where designed. If these test frameworks where created _now_, then most likely some parts of their design would have been done differently. What I am trying to say is that with even the slightest change in the Go language and standard library, completely new ways of designing programs become possible. Especially for _new_ packages without any legacy use-cases to consider. And this brings us to generics.

With generics added to our tool-box, can we manage to solve all six problems by _starting over_?

## But why do we even need a matcher library?

Before we start of re-designing a tool, it might be worth while asking ourselves the question, what _exactly_ is the tool trying to solve? And why would anyone ever want to use it? In this case, why would anyone need a matcher library in Go?

And to be fair, you often don't. Go famously don't include any assert functionality, because the Go developers believe that it's better to do checks in tests the same way you do checks in your normal programs. I.e. if you do `if err != nil` in your program, that's the syntax for doing the same check in _your tests_. This way reading Go tests, becomes no different then reading any other go code, and you are less likely to do _mistakes_.

I believe that perhaps a common misconception could be that a matcher library is needed because _checking_ results is hard. But this simply isn't true. At least not for simple checks. We can prove this by rewriting our example test to not use a matcher library at all:

```go
func TestSum(t *testing.T) {
	a := mypkg.Vector{1, 0, 3}
	b := mypkg.Vector{0, 1, -2}
	expect := mypkg.Vector{1, 1, 1}

	result, err := mypkg.Sum(a, b)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !reflect.DeepEqual(result, expect) {
		t.Errorf("got: %v, want: %v", result, expect)
	}
}
```

As we can see, the check part is quite straight-forward. So why then would we need a matcher library?

My own opinion is that the part of writing tests that might make you _consider_ a matcher library, is the disciplinary challenge of letting the output on failures be both _useful_ and _consistent_. Especially so, if you have a project with multiple developers. How much team and review discipline do you need to consistently order the "got" and "want" parameter, for instance? Again, if you are observant, you might have noticed that the code above have already failed that test. While the test above won't tell _lies_, consistent wording, ordering and reporting of test failures, can be important tools in streamlining the debugging process. And while some teams, like the Go team, are extremely good at getting this right, other teams might prefer to just use a library.

To finish up, perhaps we can agree that the _right_ mission statement for a matcher library, might be more about providing consistent and useful information on failure than it is about making complicated checks easy; although it might do that too.

## Exploring the original design

We will soon move on to designing, or rather re-designing, our test matcher library with generics. But to do that we must understand a little bit more about the core design of the `subtest` library.

So, remember this code?

```go
t.Run("Expect correct sum", subtest.Value(result).DeepEqual(expect))
```

We mentioned the code generates a test, and we mentioned something about lack of type-safety, but we didn't give to much detail around exactly what we mean by this. Actually this code is just a _short-hand_ syntax. As it happens, the _long_ syntax for the same operation reveals the underlying design much better.

```go
t.Run("Expect correct sum", subtest.Value(result).Test(subtest.DeepEqual(expect)))
```

Or to break the code up even further:

```go
// Step 1: create a "value function" a.k.a. a value initializer:
valueFunc := subtest.Value(result)

// Step 2: create a check:
check := subtest.DeepEqual(expect)

// Step 3: Combine the value initializer and the check into a
// test function:
testFunc := valueFunc.Test(check)

// Step 4: Run the test function as a Go sub-test:
t.Run("Expect correct sum", testFunc)
```

Obviously, the check portion here can be _replaced_. E.g. instead of `subtest.DeepEqual`, we could use [`subtest.NumericEqual`](https://pkg.go.dev/github.com/searis/subtest#NumericEqual), [`subtest.Before`](https://pkg.go.dev/github.com/searis/subtest#Before) or even a user defined check. This is powered by a combination of Go interfaces and first-class function support.

While the subtest package has several checks and some formatting helpers, the core design can actually be summarized in very few lines of code:

```go

// The interface for a check.
type Check interface{
    Check(ValueFunc) error
}

// An adapter that lets a normal function implement a check.
type CheckFunc func(interface{}) error

func (f checkFunc) Check(vf ValueFunc) error {
    v, err := vf()
    if err != nil {
        return fmt.Errorf("failed to initialize value: %w", err)
    }
    return f(V)
}

// A value initializer function.
type ValueFunc func() (interface{}, error)

// An adapter to convert a plain value to a value initializer
// function.
func Value(v interface{}) ValueFunc {
    return func() (interface{}, error) {
        return v, nil
    }
}

// A method that constructs and returns a Go (sub-)test function
// from the combination of a value initializer and a check.
func (vf) Test(c Check) func(*testing.T) {
    return func(t *testing.T) {
        if err := c.Check(vf); err != nil {
            t.Fatal(err.Error())
        }
    }
}
```

So comes the problem to solve with generics. Because what's the point of passing in a `NumericEqual` check (expecting some numeric value) or a `Before` check (expecting a `time.Time`) to a value initializer returning a `mypkg.Vector` type? Can we, by use of generics, make this fail compilation?

## The generic re-design

Finally, the re-design. With _generics_, or _type parameterization_, which the proposed Go generics is more accurately called, we can enforce type-safety for the check and value initializer combination into tests. To demonstrate how, here is a re-write of the subtest core-design:

```go
// The check function type.
type CheckFunc[T any]func(func() T) error

// An adapter for converting a value to a value initializer
// function.
func Value[T any](v T) func() T {
	return func() T{
		return v
	}
}

// A function for combining a compatible value initializer and a
// check function into a Go (sub-)test.
func Test[T any](vf func() T, cf CheckFunc[T]) func(t *testing.T) {
	return func(t *testing.T) {
		if err := c.Check(vf); err != nil {
			t.Fatal(err.Error())
		}
	}
}
```

Each individual check implementation, or check initializer implementation to be exact, can be declared with various degrees of type parameterization:

```go
// A type parameterized check initializer that can be initialized
// for any type.
func DeepEqual[T any](w T) CheckFunc[T] {...}

// A type parameterized check initializer that can only be
// initialized with a comparable type.
func CompareEqual[T comparable](v T) CheckFunc[T] {...}

// A type parameterized check without a compare value.
func ReflectNil[T any]() CheckFunc[T]

// A check initializer that is not type parameterized.
func TimeBefore(t time.Time) CheckFunc[time.Time] {...}
```

The trick here is that the type used to initialize the _check_ and the type returned by the _value initializer_ needs to match. If we try to combine a `subx.TimeBefore` check with a `mypkg.Vector` value initializer, compilation would fail. If we try to combine a `mypkg.CompareEqual` with the same initializer, it will work if `mypkg.Vector` is implemented as a comparable type. E.g. if `mypkg.Vector` is implemented as a struct with three fields, that are all comparable. then it will work:

```go
type Vector struct{
	X, Y, Z float64
}
```

If however it is implemented as `type Vector []float64`, then a compilation with `CompareEqual` would fail. Assuming that `CompareEqual` won't work, we can now rewrite our test. In order to explain what's happening, we will first show this code _without_ any type inference:

```go
func TestSum(t *testing.T) {
	a := mypkg.Vector{1, 0, 3}
	b := mypkg.Vector{0, 1, -2}
	expect := mypkg.Vector{1, 1, 1}

	result, err := mypkg.Sum(a, b)

	t.Run("Expect no error", subx.Test(subx.Value[error](err), subx.CompareEqual[error](nil))
	t.Run("Expect correct sum", subx.Test(subx.Value[mypkg.Vector](result), subx.DeepEqual[mypkg.Vector](expect)))
}
```

Looking at this code, you might argue that the `subx.Test` function has reintroduced the ordering issue of the "got" and "want" parameters. However, it has not. this is because ordering this parameters wrong would lead to a _compile-time error_. We mentioned that using `CompareEqual` would fail compilation for the slice implementation of `mypkg.Vector`. Actually, even `subx.DeepEqual[*mypkg.Vector]` will cause a compilation error. This type-safety can be a useful tool to prevent simple programming mistakes.

If you read the design proposal, you would know all about type inference, and when it can and cannot be used; or perhaps you discovered how through some trial and error in the playground. With the current type-inference implemented in the `go2go` tool, the `[T]` syntax can be omitted everywhere in our example except for when comparing an error to nil. This again makes the code more readable.

```go
func TestSum(t *testing.T) {
	a := mypkg.Vector{1, 0, 3}
	b := mypkg.Vector{0, 1, -2}
	expect := mypkg.Vector{1, 1, 1}

	result, err := mypkg.Sum(a, b)

	t.Run("Expect no error", subx.Test(subx.Value(err), subx.CompareEqual[error](nil))
	t.Run("Expect correct sum", subx.Test(subx.Value(result), subx.DeepEqual(expect)))
}
```

## Some cool things we can do with subx

While not part of the core design, we define syntactic sugar that allows different short-hand methods to be placed on different value-initializer functions similar to subtest. E.g.:

```go
// Long syntax:
result := mypkg.Sum(2, 3)
t.Run("Expect correct sum", subtest.Test(subtest.Value(result), subtest.CompareEqual(5)))

// Short-hand syntax:
result := mypkg.Sum(2, 3)
t.Run("Expect correct sum", subtest.Number(result).Equal(5))
```

If we have a function that isn't reliable, we can repeat a check:

```go
vf := func() int {
	return mypkg.Sum(2, 2, 1)
}

t.Run("Expect stabler results", subx.Test(vf,
	subx.AllOf(subx.Repeat(1000, subx.CompareEqual(5))...),
))
```

If we want to run a check outside of a test, we can do that as well:

```go
result := mypkg.Sum(2, 3)
cf := subx.CompareEqual(5)
fmt.Println("CHECK sum error:", cf(subx.Value(result)))
```

## Challenge: go beyond the playground

Now I have told you how I went beyond the playground to re-design a package with Go generics. Now, it's your turn!

I hereby challenge you to find something _useful_ to do with generics. Re-design a package. Write an ORM with generics. What ever you find that you want to do, following the README instructions of [subx][subx] should be enough to get you started.

Fair warning: As of the time of writing, writing anything with generics is a bit of a time-travel in terms of editor support. We are now quite spoiled by the features of `gopls`, `gofmt`, `goimports` and more. Not to mention, syntax highlighting. At the time of writing, none of this works for Go generics; at least not out of the box. You have to remove your own blank lines, insert your own imports, and align your own struct tags. Now that you are warned; God speed!

[searis]: https://searis.no
[lang-prop]: https://github.com/golang/go/issues/43651
[design-draft]: https://go.googlesource.com/proposal/+/refs/heads/master/design/go2draft-type-parameters.md
[go2go-readme]: https://go.googlesource.com/go/+/refs/heads/dev.go2go/README.go2go.md
[updated-draft]: https://blog.golang.org/generics-next-step
[go2go-play]: https://go2goplay.golang.org/
[subtest]: https://github.com/searis/subtest
[subx]: https://github.com/smyrman/subx
[clarify]: https://clarify.us
[testify]: https://github.com/stretchr/testify
[go-generics-5min]: https://dev.to/threkk/go-2-generics-in-5-minutes-1fjf
[pkg-expect]: https://github.com/pkg/expect
[gomega]: https://onsi.github.io/gomega/
[ginko]: http://onsi.github.io/ginkgo/
