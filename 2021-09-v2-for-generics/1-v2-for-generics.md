# Exploring v2 packages for generics [DRAFT]

In [Golang Weekly #379][weekly], the feature story was [a discussion][discussion] on how to introduce generics into the standard library and elsewhere. Among the popular options is a naming schema (`..Of[T]`) and different variants of adding a [default type][default]. There is also a suggestion of [package versioning][pkgver] in the standard library. In this series of four articles, I want to explore the latter option by trying to visualize how this _might_ look like for different packages in the standard library. Our main goal is _not_ to create final proposals for how the v2 packages should look like, but rather to explore the possibility and identify possible concerns.

The following posts are part of this series:

1. This article introduction to the prospects and risks of package versioning in the Go standard library (this post).
2. An illustration of how a potential `sync/v2` package [might look like](2-sync-v2.md), if done for generics.
3. An illustration of how a potential `math/v2`package [might look like](3-math-v2.md), if done for generics.

[weekly]: https://golangweekly.com/issues/379
[discussion]: https://github.com/golang/go/discussions/48287
[default]: https://github.com/golang/go/discussions/48287#discussioncomment-1303263
[pkgver]: https://github.com/golang/go/discussions/48287#discussioncomment-1303200

## Prior knowledge

To understand these articles, you should have experience of writing Go, and understand Go code. You should know that [go modules][gomod] recommend [a /vM suffix][v2] for v2 and beyond. Finally being interested or curious about the [type parameters proposal][proposal] (a.k.a. generics proposal) in Go is an absolute requirement. If you however don't know all the _details_ of generics, don't worry, I will bring you up to speed with the most important bits.

## Speed intro to generics in Go

**If you are well familiar with the type parameterization proposal in Go, including recent updates such as the `~` syntax, please skip this section.**

The addition of type parameters in Go, is not _to different_ from function parameters, except that where function parameters declare a _type_ (e.g. `float64`) and takes a _value_ (e.g `2.0`), type parameters declare a _constraint_ (e.g. `constraints.Number`) and takes a _type_ (e.g. `float64`). Passing in a type parameter to a type parameterized function or type, result in an _initialized_ instance.

Consider the following function:

```go
func Sum[T constraints.Number](values ...T) T {
	var r T
	for _, v := range values {
		r += v
	}
	return r
}
```

To get the float64 version of this function, you can call it like this:

```go
x := Sum[float64](1.0, 2.0)
```

Similar to how goes allows to infer the type of variables, i.e. you could wite `a := 2.0` instead of `var a float64 = 2`, it's specified that the compiler should be able to defer the value of the type parameter:

```go
var a float64 = 1.0
var b float64 = 2.0
x := Sum(a, b)
```

Next, up type parameterization is valid not only for functions, but also for _types_, in which case the type parameters can be used in all type methods:

```go
type ProtectedMap[K, V any] struct{
	l sync.RWMutex
	m map[K]V
}
func (pm *ProtectedMap[K, V]) Get(k K) V {
	l.RLock()
	defer l.RUnlock()
	if pm.m == 0 {
		return V{}
	}
	return pm.m[k]
}

func (pm *ProtectedMap[K, V]) Set(k K, v V) {
	// ...
}
```

We could end here, but as a final note, it's good to know a bit more about type constraints. Type constraints are declared as interfaces, although there are new "types" of interfaces you can construct, that are valid for use with type parameterization only; at least for now. Please note that when an interface is used as a type constraint, no matter which _type_ of interface it is, the function will see the _actual implementation_ used to initialize the type parametrization, and not wrap it in an interface type like it would do if you where using the interface "normally".

In the examples above, we have already encountered two types of constraints. The first constraint we encountered where the one called `constraints.Number`. This interface is part if a proposed `constraints` package, and is declared as a _type set_:

```go
package constraints

type Number interface{
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~complex64 | ~complex128
}
```

The `~` operator means that we allow all types where the _underlying type_ match. To match an _exact_ type, this operator can be omitted. When ever a type set is used as a _constraint_, then the function can only include operations that are permitted for _all_ listed types. For `Number` in particular, we are allowed to use any _arithmetic_ operation (`+`, `-`, `*`, etc.) as well as all the ordering operations (`>`, `>=`, `==`, etc.).

The second constraint we saw was `any`. This is an alias to an _empty_ interface (`interface{}`), and thus matches any type. This interface type you should already be familiar with from Go 1.0.

This leaves two types of constraints to be explained, which we will do only very briefly. First up, you can use an interface with a _method set_. This interface type we also recognize from Go 1.0. is the type of interface we are This means that the the passed in type must implement all methods of that interface, and you can call those methods within the scope of the type parameterization.

Finally, you can use the special built-in parameter type called `comparable`, which match all types that allow the equal and unequal operators (`==` and `!=`).

[proposal]: https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md
[gomod]: https://golang.org/ref/mod
[v2]: https://go.dev/blog/v2-go-modules

## How to version a package in the standard library

There have been a number of different suggestions for how to version a package in the standard library, The choice is an important one in terms of _communication_ and _intent_, and it's not without significant when it comes to the technical details. Here are a summary of three different suggestions:

1. As a global prefix (`v2/sync`, `v2/sync/atomic`), or even a suggestion of using `go2/...`.
2. Always at the end (e.g. `sync/v2`, `sync/atomic/v2`).
3. After the first path element (e.g. `sync/v2`, `sync/v2/atomic`).

From the discussion, option one seams least likely; the only option where it could be considered, would be if all packages where to be rewritten _at once_ (standard library v2), and that doesn't seam like any of the core developers are particular interested in at the moment. This leaves option 2 and 3, or perhaps some hybrid variant where the "module root" isn't always at a fixed level in the standard library three.
In this series, we are going to demonstrate option 3; not because it's necessarily the best option, but because it's useful to understand the _implication_ of this approach, which is that when ever you want to rewrite a package, you are forced to rewrite everything after the first path element, for better or worse. To quote Ion Taylor's [comment][mention] in the discussion thread:

> I don't think the interesting argument is whether the sub-package imports the parent package. I think the interesting argument is whether we would normally want to have a v2 version of all the sub-packages at the same time.

The deciding factor between would be that when ever you do want to rewrite something in the standard library, you rewrite _everything_ under As [mentioned][mention] by Ion Taylor, in the GitHub discussion thread, a potential

To be consistent within the scope of this series, we will apply the versioning right after the _first_ path element, i.e. `<first>/v2/<second>/`. As a technical consequence of this, when ever we evolve a package `<first>` into `<first>/v2`, we will also have to evolve each sub-package `<first>/<second>` into `<first>/v2/<second>`, with or without changes. We do not claim that this approach is going to be the best trade-off, but it has larger implications than versioning after the _last_ element, and we do want to understand these implications _better_. Which exact approach for package versioning in the standard library is better when it all comes down to it is out of scope for this series to determine.

[mention]: https://github.com/golang/go/discussions/48287#discussioncomment-1364805
