package spectrogram

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math/cmplx"

	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	defaultOverlap    = 1
	defaultWindowSize = 1024
)

// A Spectrogram is a 2-dimensional collection of audio frequency information over time.
type Spectrogram struct {
	overlap    float64
	windowSize int
	visual     GradientType

	timeDomain []float64
	freqDomain [][]float64
}

// Options for the spectrogram.
type Options struct {
	WindowSize int
	Overlap    float64
	Visual     GradientType
}

// DefaultOptions for the spectrogram.
var DefaultOptions = &Options{
	WindowSize: defaultWindowSize,
	Overlap:    defaultOverlap,
	Visual:     Spek,
}

// NewFromSamples processes raw time domain audio samples and creates a spectrogram.
func NewFromSamples(samples []float64, opts *Options) *Spectrogram {
	if opts == nil {
		opts = DefaultOptions
	}
	s := &Spectrogram{
		overlap:    opts.Overlap,
		windowSize: opts.WindowSize,
		visual:     opts.Visual,
		timeDomain: samples,
	}
	s.processAudio(samples)
	return s
}

// New creates an empty, unprocessed spectrogram. This is typically used for real-time
// processing, where one can add samples to be processed by the spectrogram with the
// AppendBuffer method.
func New(opts *Options) *Spectrogram {
	if opts == nil {
		opts = DefaultOptions
	}
	return &Spectrogram{
		overlap:    opts.Overlap,
		windowSize: opts.WindowSize,
		visual:     opts.Visual,
	}
}

// ProcessBuffer processes time-domain samples, adds them to the spectrogram and returns
// the processed spectral content.
func (s *Spectrogram) ProcessBuffer(buf []float64) [][]float64 {
	s.timeDomain = append(s.timeDomain, buf...)
	return s.processAudio(buf)
}

func (s *Spectrogram) TimeDomain() []float64 {
	return s.timeDomain
}

func (s *Spectrogram) FrequencyDomain() [][]float64 {
	return s.freqDomain
}

func (s *Spectrogram) WindowSize() int {
	return s.windowSize
}

// processAudio converts the time-domain samples into a frequency over time 2d slice.
func (s *Spectrogram) processAudio(buf []float64) [][]float64 {
	if len(buf) < s.windowSize {
		panic("askl;dfja;djsf")
	}

	nBins := len(buf) / int((float64(s.windowSize) * s.overlap))

	// Break the samples up into chunks of n samples each, which we call bins or windows.
	// We will need a 2D array to hold all of these bins.
	bins := make([][]float64, nBins)
	for i := range bins {
		bins[i] = make([]float64, s.windowSize)
	}

	// Split the array of all time domain samples into n sample chunks.
	var count int
	for n := 0; n < nBins; n++ {
		bins[n] = buf[count : count+s.windowSize]
		count += int((float64(s.windowSize) * s.overlap))
	}

	// Run the samples through a windowing function to minimise spectral leakage.
	hw := hannWindow(s.windowSize)
	for i := 0; i < len(bins); i++ {
		for j := 0; j < len(bins[i]); j++ {
			bins[i][j] = bins[i][j] * hw[j]
		}
	}

	// FFT the windowed time-domain samples to get the frequency domain.
	for i := 0; i < len(bins); i++ {
		frequencySpectrum := fftReal(bins[i])
		for j := 0; j < (s.windowSize/2)+1; j++ {
			mag := cmplx.Abs(frequencySpectrum[j])
			bins[i][j] = mag
		}
	}

	s.freqDomain = append(s.freqDomain, bins...)
	return bins
}

// Save writes a visual representation of the spectrogram image to the supplied writer.
func (s *Spectrogram) Save(w io.Writer) error {
	img := s.drawSpectrogram()
	return png.Encode(w, img)
}

// TODO (farleyschaefer): Testing purposes, can probably delete for now.
func (s *Spectrogram) drawSpectrogram() *image.RGBA {
	gradientWidth := 30
	padding := 10
	width := len(s.freqDomain) + padding + gradientWidth + padding
	height := (s.windowSize / 2) + 1 + padding

	spectrogram := image.NewRGBA(image.Rect(0, 0, width, height))
	colorSpectrum := GenerateSpectrum(s.visual)

	// Each bin corresponds to a vertical line.
	for i := 0; i < len(s.freqDomain); i++ {
		for j := 0; j < len(s.freqDomain[i])/2; j++ {
			power := -1 * magToDecibel(s.freqDomain[i][j], s.windowSize)
			color := GetColor(power, colorSpectrum)
			spectrogram.Set(i, height-j, color)
		}
	}

	// Draw the gradient.
	for y := height - 1; y >= 0; y-- {
		c := colorSpectrum.getInterpolatedColorFor(float64(y) / float64(height))
		draw.Draw(spectrogram, image.Rect(len(s.freqDomain)+padding, y+padding, len(s.freqDomain)+padding+gradientWidth+padding, y+1+padding), &image.Uniform{c}, image.ZP, draw.Src)
	}

	return spectrogram
}

func GetColor(power float64, spectrum gradientTable) color.Color {
	var max float64 = 100
	if power > max {
		return spectrum.getInterpolatedColorFor(max)
	}
	return spectrum.getInterpolatedColorFor(power / max)
}

type gradientTable []struct {
	col colorful.Color
	pos float64
}

func (t gradientTable) getInterpolatedColorFor(v float64) colorful.Color {
	for i := 0; i < len(t)-1; i++ {
		c1 := t[i]
		c2 := t[i+1]
		if c1.pos <= v && v <= c2.pos {
			v = (v - c1.pos) / (c2.pos - c1.pos)
			return c1.col.BlendHcl(c2.col, v).Clamped()
		}
	}
	return t[len(t)-1].col
}

func mustParseHex(color string) colorful.Color {
	c, err := colorful.Hex(color)
	if err != nil {
		panic(err)
	}
	return c
}

type GradientType int

const (
	Default GradientType = iota
	Spek
	GrayScale
)

func GenerateSpectrum(t GradientType) gradientTable {
	switch t {
	case Spek:
		return gradientTable{
			{mustParseHex("#FC181C"), 0.0},
			{mustParseHex("#FD7623"), 0.1},
			{mustParseHex("#FFEA35"), 0.2},
			{mustParseHex("#90FD32"), 0.3},
			{mustParseHex("#51FD30"), 0.4},
			{mustParseHex("#2AFD53"), 0.5},
			{mustParseHex("#24D3FD"), 0.6},
			{mustParseHex("#1162FB"), 0.7},
			{mustParseHex("#1324FB"), 0.8},
			{mustParseHex("#401DD5"), 0.9},
			{mustParseHex("#0E0E0E"), 1.0},
		}
	case GrayScale:
		return gradientTable{
			{mustParseHex("#000000"), 0.0},
			{mustParseHex("#181818"), 0.1},
			{mustParseHex("#383838"), 0.2},
			{mustParseHex("#585858"), 0.3},
			{mustParseHex("#707070"), 0.4},
			{mustParseHex("#909090"), 0.5},
			{mustParseHex("#A9A9A9"), 0.6},
			{mustParseHex("#BEBEBE"), 0.7},
			{mustParseHex("#D3D3D3"), 0.8},
			{mustParseHex("#E8E8E8"), 0.9},
			{mustParseHex("#FFFFFF"), 1.0},
		}
	default:
		return gradientTable{
			{mustParseHex("#FFFFFF"), 0.0},
			{mustParseHex("#FFFEE8"), 0.1},
			{mustParseHex("#FFFDD6"), 0.2},
			{mustParseHex("#FFFF00"), 0.3},
			{mustParseHex("#FFA31A"), 0.4},
			{mustParseHex("#F41B00"), 0.5},
			{mustParseHex("#D2003F"), 0.6},
			{mustParseHex("#990076"), 0.7},
			{mustParseHex("#51007B"), 0.8},
			{mustParseHex("#00004D"), 0.9},
			{mustParseHex("#000003"), 1.0},
		}
	}
}
