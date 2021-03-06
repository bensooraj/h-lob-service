package limitorderbook

import (
	"testing"

	"github.com/google/btree"
	"github.com/robaho/fixed"
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
		{
			[][2]float64{{370.13, 73.66}, {151.17, 33.14}, {387.9, 34.26}, {196.58, 37.96}, {148, 34.23}, {499.7, 53.9}, {210.41, 6.64}, {362.05, 67.16}, {249.82, 31.53}, {114.88, 74.17}, {143.28, 56.14}, {353.09, 61.02}, {303.53, 63.9}, {308.86, 60.5}, {158.65, 91.7}, {250.76, 93.13}, {344.36, 25.34}, {336.75, 28.02}, {244.15, 43.02}, {464.23, 82.7}, {205.81, 82.02}, {105.24, 67.17}, {351.17, 49.1}, {456.72, 81.5}, {281.51, 91.3}, {226.99, 12.46}, {146.25, 96.74}, {423.67, 33.3}, {358.11, 28.63}, {402.52, 19.45}, {175.67, 9.38}, {105.23, 79.41}, {302.33, 72.23}, {253.95, 86.74}, {213.57, 16.93}, {124.45, 55.54}, {171.81, 5.08}, {376.29, 86.5}, {134.71, 6.69}, {114.38, 58.19}, {254.56, 84.42}, {261.14, 33.25}, {423.14, 65.12}, {297.18, 84.35}, {234, 50.03}, {155.52, 32.87}, {135.42, 94.25}, {417.84, 76.1}, {291.43, 16.53}, {342.22, 83.26}, {103.57, 93.84}, {453, 45.66}, {364.46, 32.01}, {127.04, 62.41}, {241.34, 98.67}, {416.07, 71.07}, {109.52, 71.25}, {360.53, 55.77}, {464.79, 6.45}, {423.94, 40.17}, {279.78, 20.01}, {349.36, 93.77}, {197.57, 45.57}, {454.13, 61.92}, {371.15, 95.66}, {306.93, 25.76}, {421.89, 75.34}, {273.64, 44.55}, {198.02, 39.89}, {361.33, 92.75}, {113.66, 3.13}, {384.26, 56.18}, {166.63, 8.29}, {388.64, 20.78}, {209.67, 90.4}, {285.12, 58.66}, {165.73, 37.68}, {288.54, 18.71}, {405.31, 53.39}, {124.22, 44.32}, {405.6, 95.98}, {126.7, 55.13}, {288.17, 53.56}, {147.89, 84.85}, {114.74, 59.45}, {132.32, 22.39}, {243.76, 35.12}, {435.27, 8.92}, {392.99, 87.05}, {347.82, 84.56}, {477.38, 46.38}, {147.59, 27.89}, {205.78, 99.38}, {145.08, 60.29}, {138.6, 29.78}, {168.07, 65.07}, {429.28, 11.93}, {302.11, 9.25}, {168.55, 42.22}, {101.95, 7.44}},
			[]float64{101.95, 103.57, 105.23, 105.24, 109.52, 113.66, 114.38, 114.74, 114.88, 124.22, 124.45, 126.7, 127.04, 132.32, 134.71, 135.42, 138.6, 143.28, 145.08, 146.25, 147.59, 147.89, 148, 151.17, 155.52, 158.65, 165.73, 166.63, 168.07, 168.55, 171.81, 175.67, 196.58, 197.57, 198.02, 205.78, 205.81, 209.67, 210.41, 213.57, 226.99, 234, 241.34, 243.76, 244.15, 249.82, 250.76, 253.95, 254.56, 261.14, 273.64, 279.78, 281.51, 285.12, 288.17, 288.54, 291.43, 297.18, 302.11, 302.33, 303.53, 306.93, 308.86, 336.75, 342.22, 344.36, 347.82, 349.36, 351.17, 353.09, 358.11, 360.53, 361.33, 362.05, 364.46, 370.13, 371.15, 376.29, 384.26, 387.9, 388.64, 392.99, 402.52, 405.31, 405.6, 416.07, 417.84, 421.89, 423.14, 423.67, 423.94, 429.28, 435.27, 453, 454.13, 456.72, 464.23, 464.79, 477.38, 499.7},
			499.7,
			101.95,
		},
	}

	for _, test := range bidTestCases {
		l2lob := NewL2LimitOrderBook(0)
		for _, pq := range test.bids {
			price := pq[0]
			quantity := pq[1]

			fPrice := fixed.NewF(price)
			l2lob.UpdateOrAdd(LoBFixed(fPrice), quantity, "b")
		}
		actual := []float64{}
		l2lob.Bids.Ascend(func(price btree.Item) bool {
			fPrice := fixed.Fixed(price.(LoBFixed))
			actual = append(actual, fPrice.Float())
			return true
		})
		assert.Equalf(test.expectedBidsOrder, actual, "The prices must be in ascending order!")

		// Max
		max := fixed.Fixed(l2lob.Bids.Max().(LoBFixed))
		assert.Equalf(test.expectedMax, max.Float(), "The max bid price must be: %.2f", test.expectedMax)

		// Min
		min := fixed.Fixed(l2lob.Bids.Min().(LoBFixed))
		assert.Equalf(test.expectedMin, min.Float(), "The min bid price must be: %.2f", test.expectedMin)
	}

}
