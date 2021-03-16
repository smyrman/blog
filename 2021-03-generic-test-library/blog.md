# Go generics beyond the playground

While Go as of version 1.16 does not support Generics, there is an accepted [language proposal][lang-prop] for it, and we can _probably_ expect it to arrive in Go 1.18 or later. But as it turns out, we don't need to wait for it to start experimenting with the new feature. We can start _now_ to write small snippets of code on the [go2go playground][go2go-play], or even full Go packages that you [develop locally][go2go-readme].

By far the easiest way to test out Go generics, is to use the playground. And there is nobody saying that the playground isn't awesome. However awesome though, there is clear limits to how much you can reasonably try out in the playground alone. What if you have enough code to start splitting into files? What if you want to write _unit-tests_? How would a full package look like with generics? In my view, the _best_ way to try out a new feature, is to actually do something useful. And to do this with generics, we need to venture out of the safety and comfort of the playground.

> My hope is that this will _inspire_ you to do your own experiments with Go generics _beyond the playground_, by writing something _potentially useful_. Only then, can we truly see if generics itself, is going to be useful in Go.

In this article, I will go through how I [re-wrote][subx] a [test matching library][subtest] from scratch with generics as part of the tool-box. My hope is that this will _inspire_ you to do your own experiments with Go generics _beyond the playground_ and write something _potentially useful_. Only then, can we truly see if generics itself, is going to be useful in Go. If you want, you could use the library in this article for _testing_ your experiments; or you can extend it with more checks and do a pull request.

Before we start: a warning. I will assume basic knowledge about the generics proposal in Go, and what's the most important key design concepts. If you don't have this knowledge, I would recommend that you first acquire it. E.g. by reading [Go 2 generics in 5 minutes][go-generics-5min], by reading the [updated design draft][design-draft], or by doing your own experiments in the [go2go playground][go2go-play] first. In the meantime, keep this article on your TO-READ list. With that said, because we are trying to do something _potentially useful_, this article will also sometimes be more about package design then on generics itself. This is the _side-effect_ of simply adding generics to your toolbox, rather then trying to write everything with generics (like we did with channels back before the release of Go v1). Be reassured though, generics will definitely still be an _essential_ part of our (re-)design.

## A problem to solve (again) with generics

Starting out with Go generics, in order to do something _useful_, we need a problem to solve. The problem I have picked for this article is one that I have solved before when I designed the [test matcher/assertion library][subtest] that we use to test the [Clarify][clarify] back-end at [Searis][searis]. But first, you probably have a question: with all the great test matcher libraries you have in Go, why on earth would you want to write a new one? And to answer that, it's worth taking a look at at one of the existing matcher libraries. Does it solve it's _mission_ in a useful way, and with the best possible package design?

In the Go world, by far the most popular Go matcher library still appear to be the `assert` package from the [stretchr/testify][testify] repo. It's an old, good and stable library. However, because it's old, and because most (old) Go libraries strived to keep full backwards compatibility, it's also an excellent library to demonstrate some _problems_ that can be worth solving if designing a new library from scratch. Let's for instance, consider the following code, testing an imaginary function `mypkg.Sum`:

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

But let's be honest though, and you would know this if you use the assert library, the last two lines you would _probably_ write like this:

```go
assert.NoError(t, err)
assert.Equal(t, result, expect)
```

Which may give the following output in the case of a failure:

```txt
--- FAIL: TestSum (0.00s)
    /Users/smyrman/Code/blog/2021-01-subtest/mypkg/sum_test.go:18:
        	Error Trace:	sum_test.go:18
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
        	Test:       	TestSum
FAIL
FAIL	github.com/smyrman/blog/2021-01-subtest/mypkg	0.125s
FAIL
```

At first glance, all of this might look fine. What's wrong with it, you might think. Well, in this _short_ snippet of code, there is actually as much as _six problems_ that I want to bring to attention:

1. We have to pass in the `t` parameter. It's a minor inconvenience, but enough for Dave Cheney to write a quite hacky but definitively [interesting package][pkg-expect] for solving it. Not recommended for production use, I might add.
2. The "got" and "want" parameter order is hard to get right. If you are observant, you might have noticed that the code above gets it _wrong_; expected is supposed to come first (for most checks in assert, with a few exceptions).
3. From simply _starring_ at the code (which is what code reviewers generally do) it's not obviously clear what's meant by _Equal_. It's well enough _documented_, but from the name alone, it's not clear if it's going to use the equal comparison (`==`), if it using `reflect.DeepEqual`, or if it can, for instance, handel equal comparison of two `time.Time` instances with a different time-zone. At some point, subtile details in how equality is implemented, might come back to bite you. Especially so if any of these details don't _match_ the comparison you _usually do_ in your programs.
4. The descriptive text for what went wrong is _optional_; and thus generally omitted. This can make debugging failing tests hard, as there is little information to exactly _what_ when wrong (other than two values did not compare equal). Especially so if you have multiple assert statements. Is it the result from the function under test that's wrong? Is this just a sanity check before running the main test?
5. For what the output potentially lack of _useful_ information, it makes up for in _redundant_ information. Why do I need a diff for finding the difference between these simple struct instances? Given the diff is printed, why do I also get the "actual" and "expected" one-liners? Why is the filename and line trace information repeated? Why is the _test name_ repeated?
6. There is no compile-time type safety for the various checks. Of course, one might say, but if type-safety is beneficial in the _rest_ of your programs, would it not also be beneficial in tests?

In our test-matcher for Clarify (called subtest), we have managed to to solve many of these problems. But before we get into how, let's first look at some code:

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
        /Users/smyrman/Code/blog/2021-01-subtest/mypkg/sum_test.go:30: not deep equal
            got: mypkg.Vector
                [0 1 -2]
            want: mypkg.Vector
                [1 1 1]
FAIL
FAIL	github.com/smyrman/blog/2021-01-subtest/mypkg	0.174s
FAIL
```

So, let's go through which problems we managed to solve by this, and which one we didn't.

First out, you might note that we do not pass in a `t` parameter. That is because what `subtest` does, is to _return a test function_. This test function can then be run as a Go _sub-test_, omitting the need to manually pass in a `t` parameter. Cleaver? I certainly thought so when I first had the idea.

Next up we solved the ordering issue of "got" and "want". You can see that we first call `Value` on the result. This actually return what we call a `ValueFunc` or _value initializer_. It is not important now to explain why it's a function, but we can perhaps get back to that. The important part now is that this type has _methods_ that let's us initialize tests. In this particular instance, we where call the method `DeepEqual` which we call with the expect parameter. Here we have also attempted to solve the clarity issue of what `Equal` does, by giving the method a more descriptive name.

Next up, as we pass the `func(*testing.T)` instance returned by `DeepEqual` to `t.Run`, this method require you to supply a _name_. Thus a description for the check is required. In fact, as we are using Go sub-tests, there is now even a way for you to re-run a particular _check_ in your test for better focus during de-bugging:

```sh
test -name '^TestSum/Expect_correct_sum'
```

Next up, we kept our test-output _brief_ without repeating any information. We should note that there definitively is some cases where you would not get _enough_ information from this brief output format. So this is defiantly just one of those trade-offs that could be improved.

As for the final problem though, compile time type-safety, `subtest` definitively falls short. You can still pass pretty much anything into the `subtest.Value` and various test initializer methods, and it won't detect any inconsistencies for you before the test run. Besides, does it really make sense to have the same test initializer methods available for all Go types? Should not a `subtest.Value(time.Time{})` provide methods that are useful for comparing times, such as a time specific `Time{}.Equal`, `Time{}.Before` and `Time{}.After` method?

If we do the count, we gather that `subtest` appear to solve five out of the six problems we identified with the assert library. At this point though, it's important to note that at the point the `assert` package was designed, the sub-test feature in Go did not yet exist. Therefore it would have been impossible for that library to embed it into it's design. This is also true for [Gomega][gomega] and [Ginko][ginko], which have been to some inspiration for subtest. This proves to demonstrate that with new features in the Go language and standard library, new ways of designing programs become possible. And this brings us to generics.

With generics added to our tool-box, can we manage to solve all six problems by _starting over_?
## But why do we even need a matcher library?

Before we start of re-designing a tool, it might be worth while asking yourself the question, what _exactly_ is the tool trying to solve? And why would you ever want to use it? In this case, why do I need a matcher library in Go?

And to be fair, you don't _need_ a matcher library in Go. Go famously don't include any assert functionality, because the Go developers believe that it's better to do checks in tests the same way you do checks in your normal programs. I.e. if you do `if err != nil` in your program, that's the syntax for doing the same check in _your tests_. This way reading Go tests, becomes no different then reading any other go code, and you are less likely to do _mistakes_.

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

As you can see, the check part is quite straight-forward. So why then would you need a matcher library?

My own opinion is that the difficult part in writing tests that should make you consider a matcher library, is the disciplinary challenge of letting the output on failures be both _useful_ and _consistent_. Especially so, if you have a project with multiple developers. How much team and review discipline do you need to consistently order the "got" and "want" parameter, for instance?

So to finish up, the _mission statement_ for any matcher library, might be more about providing consistent and useful failure information than it is about doing complicated checks easy; although it might do that too.

## Exploring the original design

We will soon move on to designing, or rather re-designing, our test matcher library with generics. But to do that we must understand a little bit more about the core design of the `subtest` library.

So, remember this code?

```go
t.Run("Expect correct sum", subtest.Value(result).DeepEqual(expect))
```

We mentioned the code generates a test, and we mentioned something about lack of type-safety, but we didn't give to much detail around exactly what we meant by this. Well, actually this code is just a _short-hand_ syntax. As it happens, the _long_ syntax for the same operation reveals the underlying design much better.

```go
t.Run("Expect correct sum", subetst.Value(result).Test(subets.DeepEqual(expect)))
```

Or to break the code up even further:

```go
// Step 1: create a "value function" a.k.a. a value initializer:
valueFunc := subtest.Value(result)

// Step 2: create a check:
check := subtest.DeepEqual(expect)

// Step 3: Combine the value initializer and the check into a test function:
testFunc := valueFunc.Test(check)

// Step 4: Run the test function as a Go sub-test:
t.Run("Expect correct sum", testFunc)
```

Obviously, the check portion here can be _replaced_. E.g. instead of `subtest.DeepEqual`, we could use [`subtest.NumericEqual`](https://pkg.go.dev/github.com/searis/subtest#NumericEqual), [`subtest.Before`](https://pkg.go.dev/github.com/searis/subtest#Before) or even a user defined check. This  is powered by a combination of Go interfaces and first-class functions.

While the subtest package has lots of checks and some formatting helpers, the core design is actually implemented in very few lines of code, summarized here:

```go

type Check interface{
	Check(ValueFunc) error
}

type CheckFunc func(interface{}) error

func (f checkFunc) Check(vf ValueFunc) error {
	v, err := vf()
	if err != nil {
		return fmt.Errorf("failed to initialize value: %w", err)
	}
	return f(V)
}


type ValueFunc func() (interface{}, error)

func Value(v interface{}) ValueFunc {
	return func() (interface{}, error) {
		return v, nil
	}
}

func (vf) Test(c Check) func(*testing.T) {
	return func(t *testing.T) {
		if err := c.Check(vf); err != nil {
			t.Fatal(err.Error())
		}
	}
}

```

So comes the problem to solve with generics. Because what's the point of passing in a `NumericEqual` check (expecting some numeric value) or a `Before` check (expecting a `time.Time`) to a value initializer returning a `mypkg.Vector` type? Can we, by use of generics avoid that?

## The generic re-design

Finally, the re-design. With _generics_, or _type parameterization_, which the proposed Go generics is more accurately called, we can enforce type-safety for the check and value initializer combination into tests. To demonstrate how, here is a re-write of the subtest core-design:

```go
type CheckFunc[T any]func(func() T) error

func Value[T any](v T) func() T {
	return func() T{
		return v
	}
}

func Test(vf func() T, cf CheckFunc[T]) func(t *testing.T) {
	return func(t *testing.T) {
		if err := c.Check(vf); err != nil {
			t.Fatal(err.Error())
		}
	}
}
```

Each individual check can either rely on type parameterization or not. E.g.:

```go
// A type parameterized check that works for comparable values.
func CompareEqual[T comparable](v T) CheckFunc[T] {...}

// A check that works for time values.
func TimeBefore(t time.Time) CheckFunc[time.Time] {...}

// A type parameterized check that works on any value, but compares against an
// interface instead of the specific type.
func DeepEqual[T any](w interface{}) CheckFunc[T] {...}

// A type parameterized check that works on any value, without a compare value.
func ReflectNil[T any]() CheckFunc[T]
```

The trick here is that the type used to initialize the _check_ and the type returned by the _value initializer_ needs to match. If we try to combine a `subx.TimeBefore` check with a `mypkg.Vector` value initializer, compilation would fail. If we try to combine a `mypkg.CompareEqual` with the same initializer, it will work if `mypkg.Vector` is implemented as a comparable type. E.g. if `mypkg.Vector` is implemented as a struct with three fields, that are all comparable. then it will work:

```go
type Vector struct{
	X, Y, Z float64
}
```

If however it is implemented as `type Vector []float64`, then a compilation with `CompareEqual` would fail. Assuming that `CompareEqual` won't work, we can now rewrite our test:

```go
func TestSum(t *testing.T) {
	a := mypkg.Vector{1, 0, 3}
	b := mypkg.Vector{0, 1, -2}
	expect := mypkg.Vector{1, 1, 1}

	result, err := mypkg.Sum(a, b)

	t.Run("Expect no error", subx.Test(subx.Value(err), subx.CompareEqual(nil))
	t.Run("Expect correct sum", subx.Test(subx.Value(result), subx.DeepEqual[mypkg.Vector](expect)))
}
```

Note that due to Go generic type inference, you don't always need to explicitly specify that type (using `[T]` syntax) when using a generic type.

## Some cool things you can do with subx

While not part of the core design, we define syntactic sugar with different short-hand methods based on which value-initializer you used. E.g.:

```go
// Instead Of:
result := mypkg.Sum(2, 3)
t.Run("Expect correct sum", subtest.Test(subtest.Value(result), subtest.CompareEqual(5)))

// You can write:
result := mypkg.Sum(2, 3)
t.Run("Expect correct sum", subtest.Number(result).Equal(5))
```

If you have a function that isn't reliable, you can repeat a check:

```go
vf := func() int {
	return Sum[int](2, 2, 1)
}

t.Run("Expect stabler results", subx.Test(vf,
	subx.AllOf(subx.Repeat(1000, subx.CompareEqual[int](5))...),
))
```

If you want to run a check outside of a test, you can do that as well:

```go
result := mypkg.Sum(2, 3)
cf := subtest.CompareEqual(5)
fmt.Println("CHECK sum error:", cf(subtest.Value(result)))
```

## Challenge: go beyond the playground

Now I have told you how I went beyond the playground to re-design a package with Go generics. Now, it's your turn!

I hereby challenge you to find something _useful_ to do with generics. Re-design a package. Write an ORM with generics. PR more checks for subx. What ever you find that you want to do, following the README instructions of [go2go][go2go-readme] and [subx][subx] should be enough to get you started.

Fair warning: As of the time of writing, writing anything with generics is a bit of a time-travel in terms of editor support. We are now quite spoiled by the features of `gopls`, `gofmt`, `goimport` and more. Not to mention, syntax highlighting. At the time of writing, none of this works for Go generics. You have to remove your own whitespace, insert your own imports, and align your own struct tags. No that you are warned; God speed!

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
