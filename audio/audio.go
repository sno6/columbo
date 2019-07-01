package audio

// ToMono converts samples from multiple channels to a single (mono) channel by averaging.
func ToMono(s []float64) float64 {
	var average float64
	for _, v := range s {
		average += v
	}
	return average / float64(len(s))
}

// newRange is used to transform sample data into a new (possibly) lower range
// for the benefit of presenting this data in a diagram or something similar.
//
// For example, if we had an 8 bit sample with a value between 0->255
// (because 8 bit samples are unsigned) we can change that to a value between -1 and 1
// which is cleaner to represent in a graph.
//
// sample := newRange(samplePoint, 0, 255, -1, 1)
// sample will now hold a value between -1 and 1 relative to its old range of 0->255.
func newRange(value, oMin, oMax, nMin, nMax float64) float64 {
	return (((value - oMin) * (nMax - nMin)) / (oMax - oMin)) + nMin
}
