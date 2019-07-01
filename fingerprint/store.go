package fingerprint

import (
	"fmt"
	"time"

	"github.com/sno6/columbo/audio"
)

type Store interface {
	Process(directory string) error
	FindFromHashes(hashes []string) (*audio.Track, int)
}

type MemoryStore struct {
	tracks map[*audio.Track]map[string]struct{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tracks: make(map[*audio.Track]map[string]struct{})}
}

// func (m *MemoryStore) Process(directory string) error {
// 	allowedExts := []string{".wav"}

// 	files, err := ioutil.ReadDir(directory)
// 	if err != nil {
// 		return err
// 	}

// 	start := time.Now()

// 	for _, file := range files {
// 		name := filepath.Join(directory, file.Name())

// 		if !in(filepath.Ext(name), allowedExts) {
// 			continue
// 		}

// 		f, err := os.Open(name)
// 		if err != nil {
// 			return err
// 		}
// 		defer f.Close()

// 		w, err := audio.NewWAV(f)
// 		if err != nil {
// 			return err
// 		}

// 		samples, err := w.ReadSamples()
// 		if err != nil {
// 			return fmt.Errorf("error reading samples from %s: %v", file.Name(), err)
// 		}

// 		s := spectrogram.NewFromSamples(samples, spectrogram.DefaultOptions)

// 		peaks := FindPeaks2(s.FrequencyDomain(), s.WindowSize())
// 		hashes := HashPeaks(peaks)
// 		track := &audio.Track{Name: name}

// 		fmt.Printf("[+] Processed %s created %d hashes\n", name, len(hashes))

// 		m.tracks[track] = newHashMap(hashes)
// 	}

// 	fmt.Printf("[+] Finished processing your audio library took: %v\n", time.Since(start))
// 	return nil
// }

func in(s string, sl []string) bool {
	for _, st := range sl {
		if s == st {
			return true
		}
	}
	return false
}

func newHashMap(hashes []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, hash := range hashes {
		m[hash] = struct{}{}
	}
	return m
}

func (m *MemoryStore) FindFromHashes(hashes []string) (*audio.Track, int) {
	var (
		c int
		a *audio.Track
	)

	type stuff struct {
		t *audio.Track
		c int
	}
	var things []stuff

	start := time.Now()
	for track, hashMap := range m.tracks {
		var ic int
		for _, hash := range hashes {
			if _, ok := hashMap[hash]; ok {
				ic++
			}
		}
		if ic > c {
			c = ic
			a = track
		}
		things = append(things, stuff{t: track, c: ic})
	}
	fmt.Printf("[+] Finding took %v\n", time.Since(start))

	for _, t := range things {
		fmt.Printf("Song: %v has %d hits\n", t.t.Name, t.c)
	}

	return a, c
}
