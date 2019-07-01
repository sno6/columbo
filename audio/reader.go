package audio

import (
	"encoding/binary"
	"io"
	"math"
)

type reader interface {
	read() ([]float64, error)
}

type (
	pcm8 struct {
		r  io.Reader
		ch int
	}

	pcm16 struct {
		r  io.Reader
		ch int
	}
)

func (r *pcm8) read() ([]float64, error) {
	s := make([]float64, r.ch)
	for i := range s {
		var d uint8
		if err := binary.Read(r.r, binary.LittleEndian, &d); err != nil {
			return nil, err
		}
		s[i] = newRange(float64(d), 0, math.MaxUint8, -1, 1)
	}
	return s, nil
}

func (r *pcm16) read() ([]float64, error) {
	s := make([]float64, r.ch)
	for i := range s {
		var d int16
		if err := binary.Read(r.r, binary.LittleEndian, &d); err != nil {
			return nil, err
		}
		s[i] = newRange(float64(d), math.MinInt16, math.MaxInt16, -1, 1)
	}
	return s, nil
}
