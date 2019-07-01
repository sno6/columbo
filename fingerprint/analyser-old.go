package fingerprint

// // FindPeaks finds the peaks in a spectrogram.
// func (a *Analyser) FindPeaks() []Peak {
// 	var peaks []Peak
// 	maxFreq := a.spec.WindowSize()
// 	bins := a.spec.FrequencyDomain()

// 	if len(bins) < a.boxSize {
// 		panic("hmm")
// 	}

// 	for i := 0; i < len(bins)-15; i += a.boxSize {
// 		for j := 0; j < maxFreq-15; j += a.boxSize {
// 			peak := Peak{Mag: -10000}
// 			for x := 0; x < a.boxSize; x++ {
// 				for y := 0; y < a.boxSize; y++ {
// 					mag := bins[i+x][j+y]
// 					if mag > peak.Mag {
// 						peak.Mag = mag
// 						peak.Time = i + x
// 						peak.Freq = j + y
// 					}
// 				}
// 			}
// 			if peak.Mag != -10000 {
// 				peaks = append(peaks, peak)
// 			}
// 		}
// 	}

// 	// Find the average magnitude value.
// 	var mean float64
// 	for _, peak := range peaks {
// 		mean += peak.Mag
// 	}
// 	mean /= float64(len(peaks))

// 	// Find the standard deviation.
// 	var dev float64
// 	for _, peak := range peaks {
// 		dev += math.Pow(peak.Mag-mean, 2)
// 	}
// 	dev /= float64(len(peaks))
// 	dev = math.Sqrt(dev)

// 	return filter(peaks, mean)
// }

// func FindPeaks2(bins [][]float64, ws int) []Peak {
// 	var peaks []Peak
// 	maxFreq := ws

// 	for i := 0; i < len(bins)-15; i += 15 {
// 		for j := 0; j < maxFreq-15; j += 15 {
// 			peak := Peak{Mag: -10000}
// 			for x := 0; x < 15; x++ {
// 				for y := 0; y < 15; y++ {
// 					mag := bins[i+x][j+y]
// 					if mag > peak.Mag {
// 						peak.Mag = mag
// 						peak.Time = i + x
// 						peak.Freq = j + y
// 					}
// 				}
// 			}
// 			if peak.Mag != -10000 {
// 				peaks = append(peaks, peak)
// 			}
// 		}
// 	}

// 	// Find the average magnitude value.
// 	var mean float64
// 	for _, peak := range peaks {
// 		mean += peak.Mag
// 	}
// 	mean /= float64(len(peaks))

// 	// Find the standard deviation.
// 	var dev float64
// 	for _, peak := range peaks {
// 		dev += math.Pow(peak.Mag-mean, 2)
// 	}
// 	dev /= float64(len(peaks))
// 	dev = math.Sqrt(dev)

// 	return filter(peaks, mean)
// }

// func filter(peaks []Peak, threshold float64) []Peak {
// 	var filtered []Peak
// 	for _, peak := range peaks {
// 		if peak.Mag > threshold {
// 			filtered = append(filtered, peak)
// 		}
// 	}
// 	return filtered
// }

// // HashPeaks ...
// func HashPeaks(peaks []Peak) []string {
// 	var hashes []string
// 	maxDelta := 200

// 	for i := range peaks {
// 		for j := range peaks {
// 			if i == j || peaks[j].Time < peaks[i].Time {
// 				continue
// 			}
// 			delta := peaks[j].Time - peaks[i].Time
// 			if delta > maxDelta {
// 				continue
// 			}
// 			hashes = append(hashes, hash(peaks[i].Freq, peaks[j].Freq, delta))
// 		}
// 	}
// 	return hashes
// }

// func hash(freq1, freq2, delta int) string {
// 	h := md5.New()
// 	io.WriteString(h, fmt.Sprintf("%d:%d:%d", freq1, freq2, delta))
// 	return fmt.Sprintf("%x", h.Sum(nil))
// }

// Low pass..
// //Perform a low pass filter on this buf.
// xv := make([]float64, 3)
// yv := make([]float64, 3)

// ax := make([]float64, 3)
// by := make([]float64, 3)

// getCoefficients(44100, 5512, ax, by)

// for i := 0; i < len(buf); i++ {
// 	xv[2] = xv[1]
// 	xv[1] = xv[0]
// 	xv[0] = buf[i]
// 	yv[2] = yv[1]
// 	yv[1] = yv[0]

// 	yv[0] = (ax[0]*xv[0] + ax[1]*xv[1] + ax[2]*xv[2] - by[1]*yv[0] - by[2]*yv[1])

// 	buf[i] = yv[0]
// }

// //	Decimation...
// newBuf := make([]float64, len(buf)/4)
// for i := range newBuf {
// 	var average float64
// 	for _, v := range buf[i*4 : i*4+4] {
// 		average += v
// 	}
// 	newBuf[i] = average / 4
// }
