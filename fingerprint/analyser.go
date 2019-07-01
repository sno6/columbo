package fingerprint

import (
	"github.com/sno6/columbo/spectrogram"
)

// TODO(sno6): The analyser does too many things. It reads from a stream, writes to a stream AND processes
// and anaylses' the audio. This could be split up into sub modules.

// An Analyser analyses a stream of raw audio in real-time and estimates the track it belongs to.
type Analyser struct {
	// Raw audio input.
	stream <-chan []float64

	// Predictions on the data are passed back on the output channel.
	output chan *Packet

	// Pre-processed audio.
	store Store

	// Internal spectral representation of the raw audio.
	spec *spectrogram.Spectrogram

	// "Constellation map" of the most important peaks in the processed audio.
	peaks []Peak

	// Last bin number that had been processed
	last int

	// Peak statistics.
	count, sum, mean, dev float64
}

const analysisInterval = 16 // Something divisible by 1024 to help peak box.

// A Peak is a strong magnitude at a time and frequency point in a spectrogram.
type Peak struct {
	Time int     `json:"time"`
	Freq int     `json:"freq"`
	Mag  float64 `json:"mag"`
}

// A Packet encapsulates all the information that the analyser will send to output.
type Packet struct {
	Type   PacketType  `json:"type"`
	Bins   [][]float64 `json:"bins,omitempty"`
	Peaks  []Peak      `json:"peaks,omitempty"`
	Result string      `json:"result,omitempty"`
}

type PacketType int

const (
	spectral PacketType = iota
	result
	peak
)

// NewAnalyser initialises a new audio analyser which it returns as well as an
// output channel which holds information that the analyser deems interesting.
func NewAnalyser(stream <-chan []float64, store Store) (*Analyser, chan *Packet) {
	output := make(chan *Packet)
	return &Analyser{
		stream: stream,
		output: output,
		store:  store,
		spec:   spectrogram.New(spectrogram.DefaultOptions),
	}, output
}

// Start starts listening for audio over the stream, processes the spectral content,
// finds interesting data in that content, and sends that information back to the caller
// over the output channel.
// TODO (farleyschaefer): The stream needs a mechanism to shutdown
func (a *Analyser) Start() {
	for buf := range a.stream {
		bins := a.spec.ProcessBuffer(buf)
		a.output <- &Packet{Type: spectral, Bins: bins}

		// Analyse the audio every time we have "analysisInterval" many bins.
		if len(a.spec.FrequencyDomain())%analysisInterval == 0 {
			a.analyse()
		}
	}
}

// Analyse the spectrogram and peak map to find matching audio.
func (a *Analyser) analyse() {
	a.findPeaks()
	a.last += analysisInterval
}

func (a *Analyser) updateStats(peaks []Peak) {
	a.count += float64(len(peaks))
	for _, peak := range peaks {
		a.sum += peak.Mag
	}
	a.mean = a.sum / float64(a.count)
}

func (a *Analyser) findPeaks() {
	bins := a.spec.FrequencyDomain()[a.last : a.last+analysisInterval]

	height := len(bins[0])

	// Strongest peaks from each segment.
	peaks := make([]Peak, analysisInterval)

	// Firstly segment the 2d array into smaller square sections.
	// Then find the strongest peak in each segment.
	jump := height / analysisInterval
	for i := 0; i < height; i += jump {
		var seg []Peak
		for j := 0; j < len(bins); j++ {
			for freq, mag := range bins[j][i : i+jump] {
				seg = append(seg, Peak{
					Time: (len(a.spec.FrequencyDomain()) - analysisInterval) + j,
					Freq: freq + i,
					Mag:  mag})
			}
		}
		peak := findStrongestPeak(seg)
		peaks = append(peaks, peak)
	}

	a.updateStats(peaks)

	// Secondly, filter the surviving peaks based on a density criterion.
	// Typically we want 3 peaks to survive per interval. Do this by updating
	// and checking a running standard deviation of the typical peak values.
	var candidates []Peak
	for _, peak := range peaks {
		if peak.Mag > a.mean {
			candidates = append(candidates, peak)
		}
	}
	a.peaks = append(a.peaks, candidates...)

	// Send the peaks back to the front-end to somehow render.
	a.output <- &Packet{Type: peak, Peaks: candidates}
}

// Simple max function to find the strongest peak
func findStrongestPeak(peaks []Peak) Peak {
	var mMag float64
	var mPeak Peak
	for _, peak := range peaks {
		if peak.Mag > mMag {
			mPeak = peak
			mMag = peak.Mag
		}
	}
	return mPeak
}
