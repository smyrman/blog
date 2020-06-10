package mypkg_test

import (
	"testing"

	"github.com/smyrman/blog/2020-06-test-with-expect/mypkg"

	"github.com/searis/subtest"
)

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
