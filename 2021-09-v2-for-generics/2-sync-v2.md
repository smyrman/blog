# Package sync/v2 for generics? [DRAFT]

This article, we wil explore how the `sync` package (and `sync/atomic`) might look like, if a upgraded to support generics through _package versioning_. This is not a _formal_ proposal, but rather part of a [larger discussion][discussion] on Go's GitHub project. For a quick introduction to the generics proposal in Go, and to the flavor of versioning used in this series of articles, please read the [introduction article][intro].

[intro]: 1-v2-for-generics.md
[discussion]: https://github.com/golang/go/discussions/48287

## How to allow interoperability

Many packages today would likely refer directly to types, such as the `sync.Mutex` or `sync.ReadWriteMutex` types. Most of the time, it would do so internally, and not take these types as parameters, but there might be cases that do so. _Ideally_, it would be possible to use a `sync.Mutex` from `sync/v2` as a drop-in replacement for a `sync.Mutex`. Luckily Go allows _type aliases_, so if we _move_ the definition of `sync.Mutex` and other types to the `v2` package the `v1` package could use a type alias to refer to it.

If type aliases where to support type parameterization, then we could perhaps apply similar tactics to types that are of interest to converted to generics; e.g. `sync.Map`. This way we could reduce code duplication as well as binary bloat for programs that ends up including both package versions.

## sync/v2

To imagine how a `sync/v2` package, might look like, let's recap how the `sync` package looks today. Below is a list of the public interface of the package as of Go 1.17:

```go
type Cond
	func NewCond(l Locker) *Cond
	func (c *Cond) Broadcast()
	func (c *Cond) Signal()
	func (c *Cond) Wait()
type Locker
type Map
	func (m *Map) Delete(key interface{})
	func (m *Map) Load(key interface{}) (value interface{}, ok bool)
	func (m *Map) LoadAndDelete(key interface{}) (value interface{}, loaded bool)
	func (m *Map) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)
	func (m *Map) Range(f func(key, value interface{}) bool)
	func (m *Map) Store(key, value interface{})
type Mutex
	func (m *Mutex) Lock()
	func (m *Mutex) Unlock()
type Once
	func (o *Once) Do(f func())
type Pool
	func (p *Pool) Get() interface{}
	func (p *Pool) Put(x interface{})
type RWMutex
	func (rw *RWMutex) Lock()
	func (rw *RWMutex) RLock()
	func (rw *RWMutex) RLocker() Locker
	func (rw *RWMutex) RUnlock()
	func (rw *RWMutex) Unlock()
type WaitGroup
	func (wg *WaitGroup) Add(delta int)
	func (wg *WaitGroup) Done()
	func (wg *WaitGroup) Wait()
```

We can imagine that we move all of this code into a sub-folder `sync/v2`. Everything remain the same, except the `Map` and `Pool` types which are rewritten to support generics:

```go
type Map[K,V any]
	func (m *Map[K,V]) Delete(key K)
	func (m *Map[K,V]) Load(key K) (value V, ok bool)
	func (m *Map[K,V]) LoadAndDelete(key K) (value V, loaded bool)
	func (m *Map[K,V]) LoadOrStore(key K, value V) (actual V, loaded bool)
	func (m *Map[K,V]) Range(f func(key K, value V) bool)
	func (m *Map[K,V]) Store(key K, value V)
type Pool[V any]
	func (p *Pool[V]) Get() V
	func (p *Pool[V]) Put(x V)
```

As it turns out, the implementation of the `sync` "v1" can perhaps be written in as little as 10 lines of code:

```go
package sync

import "sync/v2"

type Cond = sync.Cond
type Locker = sync.Locker
type Map = sync.Map[interface{},interface{}]
type Mutex = sync.Mutex
type Once = sync.Once
type Pool = sync.Pool[interface{}]
type RWMutex = sync.RWMutex
type WaitGroup = sync.WaitGroup[]
```

Now, let's move on to the only sub-package.

## sync/v2/atomic

As mentioned in the [introduction article][intro], this series always introduce the versioning after the _first element_, and as a consequence, if we want to version the `sync` package, we also have to version the `sync/atomic` package. Wether this package _needs_ generics is a discussion in itself, but let's for now imagine that we do choose to add generics to almost all of the exposed functions. How could we do that?

To refresh, this is how the public interface for the package looks like today:

```go
func AddInt32(addr *int32, delta int32) (new int32)
func AddInt64(addr *int64, delta int64) (new int64)
func AddUint32(addr *uint32, delta uint32) (new uint32)
func AddUint64(addr *uint64, delta uint64) (new uint64)
func AddUintptr(addr *uintptr, delta uintptr) (new uintptr)
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)
func CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool)
func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool)
func CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool)
func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool)
func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool)
func LoadInt32(addr *int32) (val int32)
func LoadInt64(addr *int64) (val int64)
func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer)
func LoadUint32(addr *uint32) (val uint32)
func LoadUint64(addr *uint64) (val uint64)
func LoadUintptr(addr *uintptr) (val uintptr)
func StoreInt32(addr *int32, val int32)
func StoreInt64(addr *int64, val int64)
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)
func StoreUint32(addr *uint32, val uint32)
func StoreUint64(addr *uint64, val uint64)
func StoreUintptr(addr *uintptr, val uintptr)
func SwapInt32(addr *int32, new int32) (old int32)
func SwapInt64(addr *int64, new int64) (old int64)
func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) (old unsafe.Pointer)
func SwapUint32(addr *uint32, new uint32) (old uint32)
func SwapUint64(addr *uint64, new uint64) (old uint64)
func SwapUintptr(addr *uintptr, new uintptr) (old uintptr)
type Value
    func (v *Value) CompareAndSwap(old, new interface{}) (swapped bool)
    func (v *Value) Load() (val interface{})
    func (v *Value) Store(val interface{})
    func (v *Value) Swap(new interface{}) (old interface{})

```

Most obvious, perhaps, we could imagine rewriting the `Value` type to be generic:

```go
type Value[T any]
    func (v *Value) CompareAndSwap(old, new T) (swapped bool)
    func (v *Value) Load() (val T)
    func (v *Value) Store(val T)
    func (v *Value) Swap(new T) (old T)
```

Value is an interesting case, because by implementation it is controlling access to _pointers_ to the concrete type, and we could be discussing if we would really want to the Value type to look like this for a `v2` pacakage, but for the scope of this article, let's just assume that we do.

As far as I have been able to read in the discussion, this is as far as the proposals go in terms of transferring the `sync/atomic` package to use generics. However, it doesn't need to end there. In fact, we can _perhaps_ manage reduce the number of public functions from 29 to just 5 by introducing a few type set constraints:

```go

type Integer interface{
	~int32 | ~uint32 | ~int64 | ~uint64 | ~uintptr
}
type Atomic interface{
	~int32 | ~uint32 | ~int64 | ~uint64 | ~uintptr | unsafe.Pointer
}

func Add[T Integer](addr *T, delta T) (new T)
func CompareAndSwap[T Integer](addr *T, old, new T) (swapped bool)
func Load[T Atomic](addr *T) (val T)
func Store[T Atomic](addr *T, val T)
func Swap[T Atomic](addr *T, new T) (old T)
```

I say perhaps, because the devil here is in the detail. None of these functions are _actually_ implemented in the `sync/atomic` package, but rather in the internal portions of the Go runtime, and we do need them to be _fast_. Therefore we don't want to do no explicit type-switching in Go here in order to determine which runtime method to call; instead, we need this to be solved _compile-time_.

In the current sync atomic package, we _refer_ to the implementation using the Go assembly syntax in a file called `asm.s`. If this is to _work_, then the assembly syntax itself must be extended to support type parameterization values. E.g.:

```asm
TEXT ·Swap[int32](SB),NOSPLIT,$0
	JMP	runtime∕internal∕atomic·Xchg(SB)

TEXT ·Swap[uint32](SB),NOSPLIT,$0
	JMP	runtime∕internal∕atomic·Xchg(SB)

TEXT ·Swap[int64](SB),NOSPLIT,$0
	JMP	runtime∕internal∕atomic·Xchg64(SB)

TEXT ·Swap[uint64](SB),NOSPLIT,$0
	JMP	runtime∕internal∕atomic·Xchg64(SB)

TEXT ·Swap[uintptr](SB),NOSPLIT,$0
	JMP	runtime∕internal∕atomic·Xchguintptr(SB)
```

With this done, `sync/atomic` can either be changed to become a wrapper for `sync/v2/atomic`:

```go
package sync

import "sync/v2/atomic"

func AddInt32(addr *int32, delta int32) (new int32) {
	return atomic.Add(addr, delta)
}
func AddInt64(addr *int64, delta int64) (new int64) {
	return atomic.Add(addr, delta)
}

// And so on...

type Value = atomic.Value[interface{}]
```

Or retain it's current assembly mapping for the functions and define a type alias only for `Value`.
