package catalog

import (
	"fmt"
	"slices"
	"testing"

	"gotest.tools/v3/assert"
)

type custom struct {
	Idx   int
	Label string
}

func (c custom) GetName() string {
	return c.Label
}

func TestByName(t *testing.T) {
	t.Run("sorted slice should keep sorted", func(t *testing.T) {
		sorted := make([]custom, 0, 10)
		for i := range 10 {
			sorted = append(sorted, custom{Idx: i, Label: fmt.Sprintf("name-%d", i)})
		}

		items := make([]custom, len(sorted))
		copy(items, sorted)

		slices.SortFunc(items, CompareByName)
		assert.DeepEqual(t, items, sorted)
	})

	t.Run("should short slice by name", func(t *testing.T) {
		items := make([]custom, 10)
		for i := range 10 {
			items[i] = custom{Idx: i, Label: fmt.Sprintf("name-%d", 10-i-1)}
		}

		sorted := make([]custom, len(items))
		copy(sorted, items)
		slices.Reverse(sorted)

		slices.SortFunc(items, CompareByName)
		assert.DeepEqual(t, items, sorted)
	})

	t.Run("should find element index by name", func(t *testing.T) {
		items := make([]custom, 10)
		for i := range 10 {
			items[i] = custom{Idx: i, Label: fmt.Sprintf("name-%d", i)}
		}
		idx := slices.IndexFunc(items, EqualByName[custom]("name-5"))
		assert.Equal(t, idx, 5)

		idx, found := slices.BinarySearchFunc(items, "name-5", CompareToName)
		assert.Equal(t, found, true)
		assert.Equal(t, idx, 5)
	})

	t.Run("should not find element index by name", func(t *testing.T) {
		items := make([]custom, 10)
		for i := range 10 {
			items[i] = custom{Idx: i, Label: fmt.Sprintf("name-%d", i)}
		}
		idx := slices.IndexFunc(items, EqualByName[custom]("name-10"))
		assert.Equal(t, idx, -1)

		idx, found := slices.BinarySearchFunc(items, "name-10", CompareToName)
		assert.Equal(t, found, false)
		assert.Equal(t, idx, 2)
	})
}

func TestNames(t *testing.T) {
	t.Run("should return all names", func(t *testing.T) {
		items := make([]custom, 10)
		expected := make([]string, 10)
		for i := range 10 {
			name := fmt.Sprintf("name-%d", i)
			items[i] = custom{Idx: i, Label: name}
			expected[i] = name
		}
		names := Names(slices.Values(items))
		assert.DeepEqual(t, slices.Collect(names), expected)
	})
}
