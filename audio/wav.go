package audio

import (
	"encoding/binary"
	"errors"
	"io"
	"time"
)

type WAV struct {
	Header *WAVHeader
	r      io.Reader
}

// WAVHeader holds the three sub chunks of a wave file.
type WAVHeader struct {
	ChunkID         [4]byte // File type ID, typically "RIFF"                (4 bytes)
	ChunkSize       uint32  // Size of audio file.                           (4 bytes)
	Format          [4]byte // File format, typically "WAVE"                 (4 bytes)
	FirstChunkID    [4]byte // First chunk ID, typically "fmt"               (4 bytes)
	FirstChunkSize  uint32  // Size of the first sub chunk.                  (4 bytes)
	AudioFormat     uint16  // 1 = PCM, anything else suggests compression   (2 bytes)
	NumOfChannels   uint16  // 1 = Mono, 2 = Stereo                          (2 bytes)
	SampleRate      uint32  // Sample rate of the audio, typically 44100 Hz  (4 bytes)
	ByteRate        uint32  // Sample Byte rate.                             (4 bytes)
	BlockAlign      uint16  // Block align.                                  (2 bytes)
	BitsPerSample   uint16  // Bit depth of a sample.                        (2 bytes)
	SecondChunkID   [4]byte // Second chunk ID, typically "data"             (4 bytes)
	SecondChunkSize uint32  // Size of the following audio samples in bytes. (4 bytes)
}

// NewWAV returns a parsed wave file with header and samples.
func NewWAV(r io.Reader) (*WAV, error) {
	h, err := parseHeader(r)
	if err != nil {
		return nil, err
	}
	return &WAV{Header: h, r: r}, nil
}

// ReadSample reads a sample from its underlying reader.
// By default, PCM readers transform samples to a value range between -1 and 1.
//
// TODO (farleyschaefer): This can definitely be optimized in the future.
func (w *WAV) ReadSample() ([]float64, error) {
	var r reader
	ch := int(w.Header.NumOfChannels)
	switch w.Header.BitsPerSample {
	case 8:
		r = &pcm8{r: w.r, ch: ch}
	case 16:
		r = &pcm16{r: w.r, ch: ch}
	default:
		return nil, errors.New("currently only supporting 8 or 16 bit WAV audio")
	}
	return r.read()
}

// ReadAllSamples returns an array of all of the samples in the audio.
func (w *WAV) ReadAllSamples() ([]float64, error) {
	var samples [][]float64
	for {
		sample, err := w.ReadSample()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}
	monoSamples := make([]float64, len(samples))
	for i := range monoSamples {
		monoSamples[i] = ToMono(samples[i])
	}
	return monoSamples, nil
}

func parseHeader(r io.Reader) (*WAVHeader, error) {
	wh := &WAVHeader{}
	fields := []struct {
		data      interface{}
		byteOrder binary.ByteOrder
	}{
		{&wh.ChunkID, binary.BigEndian},
		{&wh.ChunkSize, binary.LittleEndian},
		{&wh.Format, binary.BigEndian},
		{&wh.FirstChunkID, binary.BigEndian},
		{&wh.FirstChunkSize, binary.LittleEndian},
		{&wh.AudioFormat, binary.LittleEndian},
		{&wh.NumOfChannels, binary.LittleEndian},
		{&wh.SampleRate, binary.LittleEndian},
		{&wh.ByteRate, binary.LittleEndian},
		{&wh.BlockAlign, binary.LittleEndian},
		{&wh.BitsPerSample, binary.LittleEndian},
		{&wh.SecondChunkID, binary.BigEndian},
		{&wh.SecondChunkSize, binary.LittleEndian},
	}
	for _, field := range fields {
		if err := binary.Read(r, field.byteOrder, field.data); err != nil {
			return nil, err
		}
	}
	return wh, wh.check()
}

func (wh *WAVHeader) check() error {
	if wh.BitsPerSample%8 != 0 {
		return errors.New("wave: BitsPerSample should be a multiple of 8")
	}
	if wh.BitsPerSample != 8 && wh.BitsPerSample != 16 {
		return errors.New("wave: BitsPerSample should be 8 or 16")
	}
	if wh.AudioFormat != 1 {
		return errors.New("wave: unable to read non PCM or compressed audio samples")
	}
	if string(wh.ChunkID[:]) != "RIFF" {
		return errors.New("wave: 'RIFF' file tag can not be found")
	}
	if string(wh.Format[:]) != "WAVE" {
		return errors.New("wave: 'WAVE' format tag can not be found")
	}
	return nil
}

// Duration returns the length of a wave in seconds.
func (w *WAV) Duration() time.Duration {
	samplesPerSecond := w.Header.SampleRate * uint32(w.Header.NumOfChannels)
	bytesPerSecond := samplesPerSecond * uint32((w.Header.BitsPerSample / 8))
	return time.Second * time.Duration(w.Header.SecondChunkSize/bytesPerSecond)
}

// SampleCount returns the number of samples in a WAV file.
func (w *WAV) SampleCount() int {
	return (int(w.Duration().Seconds()) * int(w.Header.SampleRate))
}
