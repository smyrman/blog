package mypkg_test

import (
	"testing"

	"github.com/searis/subtest"
	"github.com/smyrman/blog/2020-06-test-with-expect/mypkg"
)

func TestSum_(t *testing.T) {
	t.Run("With non-empty int slice", func(t *testing.T) {
		s := []int{1, 2, 3}
		i, err := mypkg.Sum(s)
		t.Run("Expect no error", subtest.Value(err).NoError())
		t.Run("Expect correct sum", subtest.Value(i).NumericEqual(6))
		t.Run("Expect input is unchanged", subtest.Value(s).DeepEqual([]int{1, 2, 3}))
	})
	t.Run("With empty int slice", func(t *testing.T) {
		s := []int{}
		i, err := mypkg.Sum(s)
		t.Run("Expect no error", subtest.Value(err).NoError())
		t.Run("Expect zero", subtest.Value(i).NumericEqual(0))
		t.Run("Expect input is unchanged", subtest.Value(s).DeepEqual([]int{}))
	})
}
