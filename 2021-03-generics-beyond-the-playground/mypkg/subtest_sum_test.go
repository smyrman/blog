package mypkg_test

import (
	"reflect"
	"testing"

	"github.com/searis/subtest"
	"github.com/smyrman/blog/2021-03-generics-beyond-the-playground/mypkg"
	"github.com/stretchr/testify/assert"
)

func TestSum(t *testing.T) {
	a := mypkg.Vector{1, 0, 3}
	b := mypkg.Vector{0, 1, -2}
	expect := mypkg.Vector{1, 1, 1}

	result, err := mypkg.Sum(a, b)

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expect) {
		t.Errorf("Unexpected result: got: %v, want: %v", result, expect)
	}
}

func TestSum_assert(t *testing.T) {
	a := mypkg.Vector{1, 0, 3}
	b := mypkg.Vector{0, 1, -2}
	expect := mypkg.Vector{1, 1, 1}

	result, err := mypkg.Sum(a, b)

	assert.NoError(t, err)
	assert.Equal(t, result, expect)
}

func TestSum_subtest(t *testing.T) {
	a := mypkg.Vector{1, 0, 3}
	b := mypkg.Vector{0, 1, -2}
	expect := mypkg.Vector{1, 1, 1}

	result, err := mypkg.Sum(a, b)

	t.Run("Expect no error", subtest.Value(err).NoError())
	t.Run("Expect correct sum", subtest.Value(result).DeepEqual(expect))
}
