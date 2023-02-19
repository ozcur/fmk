package internal

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"sort"
	"sync"

	"github.com/rivo/duplo"
	"github.com/rs/zerolog/log"
)

const (
	minScore = -70.0
)

type hashStore struct {
	dup      *duplo.Store
	idx      map[string]duplo.Hash
	idxMutex sync.Mutex
	minScore float64
}

func newHashStore() *hashStore {
	return &hashStore{
		dup:      duplo.New(),
		idx:      make(map[string]duplo.Hash),
		minScore: minScore,
	}
}

func (h *hashStore) indexPaths(paths []string) error {
	s := newSpinner()
	defer s.Stop()

	var wg sync.WaitGroup
	pathCh := make(chan string, runtime.NumCPU()*2)

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range pathCh {
				reader, err := os.Open(path)
				if err != nil {
					log.Warn().Str("path", path).
						Err(err).
						Msg("Failed to open image")
					reader.Close()
					continue
				}

				img, _, err := image.Decode(reader)
				if err != nil {
					log.Warn().Str("path", path).
						Err(err).
						Msg("Failed to decode image")
					reader.Close()
					continue
				}

				pHash, _ := duplo.CreateHash(img)
				h.dup.Add(path, pHash)

				h.idxMutex.Lock()
				h.idx[path] = pHash
				h.idxMutex.Unlock()

				reader.Close()
			}
		}()
	}

	for i, path := range paths {
		s.SetMessage(fmt.Sprintf(" %d/%d", i, len(paths)))
		pathCh <- path
	}

	close(pathCh)
	wg.Wait()

	return nil
}

func (h *hashStore) findDuplicates() [][]string {
	var ret [][]string
	discovered := make(map[string]struct{})

	s := newSpinner()
	defer s.Stop()

	i := 0
	for path, hash := range h.idx {
		i++
		s.SetMessage(fmt.Sprintf(" %d/%d", i, len(h.idx)))

		matches := h.dup.Query(hash)
		sort.Sort(matches)

		var dupeBatch []string
		for _, m := range matches {
			if m.ID.(string) == path {
				continue
			}

			if m.Score > h.minScore {
				break
			}

			_, seen := discovered[m.ID.(string)]
			if seen {
				break
			}

			dupeBatch = append(dupeBatch, m.ID.(string))
			discovered[m.ID.(string)] = struct{}{}

			log.Debug().
				Str("path", path).
				Float64("score", m.Score).
				Int("iteration", i).
				Msg("Found dupe")
		}

		if len(dupeBatch) == 0 {
			continue
		}

		dupeBatch = append(dupeBatch, path)
		discovered[path] = struct{}{}
		ret = append(ret, dupeBatch)
	}

	return ret
}
