package fileconfig

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
)

type sample struct {
	Address string `json:"address"`
	Count   int    `json:"count"`
}

func TestSaveAndLoad_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	in := sample{Address: "localhost:8080", Count: 42}

	require.NoError(t, Save(path, in))

	out, err := Load[sample](path)
	require.NoError(t, err)
	require.Equal(t, in, out)
}

func TestSave_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "cfg.json")
	in := sample{Address: "x", Count: 1}

	require.NoError(t, Save(path, in))

	_, err := os.Stat(path)
	require.NoError(t, err)

	out, err := Load[sample](path)
	require.NoError(t, err)
	require.Equal(t, in, out)
}

func TestAtomicity_ConcurrentWriteRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")

	type big struct {
		Payload string `json:"payload"`
		Count   int    `json:"count"`
	}

	payload1 := make([]byte, 256*1024)
	for i := range payload1 {
		payload1[i] = byte('A' + (i % 26))
	}
	payload2 := make([]byte, 256*1024)
	for i := range payload2 {
		payload2[i] = byte('a' + (i % 26))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var in big
				if count%2 == 0 {
					in = big{Payload: string(payload1), Count: count}
				} else {
					in = big{Payload: string(payload2), Count: count}
				}
				// Save must be atomic; readers should never see partial writes
				_ = Save(path, in)
				count++
			}
		}
	}()

	readers := 4
	var rwg sync.WaitGroup
	rwg.Add(readers)
	for i := 0; i < readers; i++ {
		go func() {
			defer rwg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					bs, err := os.ReadFile(path)
					if err != nil {
						// file may not exist yet, skip
						time.Sleep(5 * time.Millisecond)
						continue
					}
					var out big
					err = json.Unmarshal(bs, &out)
					// Must never fail due to partial writes
					require.NoError(t, err)
					require.GreaterOrEqual(t, out.Count, 0)
				}
			}
		}()
	}

	rwg.Wait()
	wg.Wait()
}
