package limitorderbook

import (
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func TestLoB_UpdateOrAdd_Small(t *testing.T) {
	assert := assert.New(t)

	bidTestCases := []struct {
		bids              [][2]float64
		expectedBidsOrder []float64
		expectedMax       float64
		expectedMin       float64
	}{
		{
			[][2]float64{{9.21, 12}, {8.23, 98}, {7.54, 1}, {6.12, 12}, {5.66, 32}, {4.12, 44}, {3.75, 22}, {2.14, 42}, {1.98, 10}},
			[]float64{1.98, 2.14, 3.75, 4.12, 5.66, 6.12, 7.54, 8.23, 9.21},
			9.21,
			1.98,
		},
		{
			[][2]float64{{63.92, 229}, {251.23, 165}, {261.32, 292}, {92.91, 23}, {154.75, 85}, {346.18, 285}, {330.07, 138}, {154.85, 340}, {200.51, 349}, {120.71, 196}, {39.42, 273}, {319.45, 392}},
			[]float64{39.42, 63.92, 92.91, 120.71, 154.75, 154.85, 200.51, 251.23, 261.32, 319.45, 330.07, 346.18},
			346.18,
			39.42,
		},
	}

	for _, test := range bidTestCases {
		l2lob := NewL2LimitOrderBook(100)
		for _, pq := range test.bids {
			price := pq[0]
			quantity := pq[1]
			l2lob.UpdateOrAdd(price, quantity, "b")
		}
		actual := []float64{}
		l2lob.Bids.Ascend(func(price btree.Item) bool {
			actual = append(actual, float64(price.(btree.Int))/100)
			return true
		})
		assert.Equalf(test.expectedBidsOrder, actual, "The prices must be in ascending order!")
		assert.Equalf(test.expectedMax, float64(l2lob.Bids.Max().(btree.Int))/100, "The max bid price must be: %.2f", test.expectedMax)
		assert.Equalf(test.expectedMin, float64(l2lob.Bids.Min().(btree.Int))/100, "The min bid price must be: %.2f", test.expectedMin)
	}

}
