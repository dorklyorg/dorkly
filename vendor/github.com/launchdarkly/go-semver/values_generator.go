package semver

import (
	"fmt"
	"testing"
)

type ValuesGenerator struct {
	valueDefs []valueConstraint
}

type valueConstraint struct {
	min int
	max int
}

func NewValuesGenerator() *ValuesGenerator {
	return &ValuesGenerator{}
}

func (g *ValuesGenerator) AddValue(min, max int) *ValuesGenerator {
	g.valueDefs = append(g.valueDefs, valueConstraint{min, max})
	return g
}

func (g ValuesGenerator) MakeAllPermutations() [][]int {
	var ret [][]int
	numValues := len(g.valueDefs)
	values := make([]int, numValues)
	for i := 0; i < numValues; i++ {
		values[i] = g.valueDefs[i].min
	}
	for {
		copied := make([]int, numValues)
		copy(copied, values)
		ret = append(ret, copied)

		// increment the values starting at the left, rolling over and incrementing the next to the right
		for pos := 0; pos < numValues; pos++ {
			if values[pos] < g.valueDefs[pos].max {
				values[pos]++
				break
			}
			if pos == numValues-1 {
				return ret // we've covered all permutations
			}
			values[pos] = g.valueDefs[pos].min
		}
	}
}

func (g ValuesGenerator) TestAll(t *testing.T, action func(*testing.T, []int)) {
	var permutations = g.MakeAllPermutations()
	for _, perm := range permutations {
		values := perm
		desc := "values"
		for _, v := range values {
			desc = desc + fmt.Sprintf(" %d", v)
		}
		t.Run(desc, func(t *testing.T) {
			action(t, values)
		})
	}
}

func (g ValuesGenerator) TestAll1(t *testing.T, action func(*testing.T, int)) {
	g.TestAll(t, func(t *testing.T, values []int) {
		action(t, values[0])
	})
}

func (g ValuesGenerator) TestAll2(t *testing.T, action func(*testing.T, int, int)) {
	g.TestAll(t, func(t *testing.T, values []int) {
		action(t, values[0], values[1])
	})
}

func (g ValuesGenerator) TestAll3(t *testing.T, action func(*testing.T, int, int, int)) {
	g.TestAll(t, func(t *testing.T, values []int) {
		action(t, values[0], values[1], values[2])
	})
}

func (g ValuesGenerator) TestAll4(t *testing.T, action func(*testing.T, int, int, int, int)) {
	g.TestAll(t, func(t *testing.T, values []int) {
		action(t, values[0], values[1], values[2], values[3])
	})
}
