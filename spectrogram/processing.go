package spectrogram

import (
	"math"

	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
)

func fftReal(samples []float64) []complex128 {
	return fft.FFTReal(samples)
}

func hannWindow(l int) []float64 {
	return window.Hann(l)
}

// log10 the magnitude to get it in db because db is a logarithmic scale dude: http://capitalavionics.com/techtime/more-on-power-logarithms-decibels-february-1999/
// The decibel value we get back is in dbfs which is relative to full scale. Purely for output.
func magToDecibel(mag float64, windowSize int) float64 {
	return 20 * math.Log10(2*mag/float64(windowSize/2))
}

// TODO: Handle downsampling.
