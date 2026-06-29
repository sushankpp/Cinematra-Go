package recommend

import (
	"log"
	"sync"
	"time"
)

// CorpusCache holds the precomputed corpus + trending scores so handlers
// never have to rebuild them per-request.
type CorpusCache struct {
	mu             sync.RWMutex
	corpus         *Corpus
	trendingScores map[string]float64
	maxTrending    float64
	lastBuilt      time.Time
}

var Cache = &CorpusCache{}

// StartBackgroundRefresh builds the cache immediately, then refreshes it
// every `interval` in a background goroutine. Call this once at server startup.
func StartBackgroundRefresh(interval time.Duration) {
	go func() {
		refresh() // build once in background — server starts listening immediately, doesn't wait

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			refresh()
		}
	}()
}

func refresh() {
	// start := time.Now()

	docs, err := FetchVideoDocs()
	if err != nil {
		log.Println("cache refresh: FetchVideoDocs failed:", err)
		return
	}
	corpus := BuildCorpus(docs)
	if corpus == nil {
		log.Println("cache refresh: BuildCorpus returned nil")
		return
	}

	scores, err := FetchTrendingScore()
	if err != nil {
		log.Println("cache refresh: FetchTrendingScore failed:", err)
		scores = map[string]float64{}
	}

	maxT := 0.0
	for _, s := range scores {
		if s > maxT {
			maxT = s
		}
	}

	Cache.mu.Lock()
	Cache.corpus = corpus
	Cache.trendingScores = scores
	Cache.maxTrending = maxT
	Cache.lastBuilt = time.Now()
	Cache.mu.Unlock()

	// log.Printf("cache refreshed in %v (videos=%d)\n", time.Since(start), len(docs))
}

// Snapshot returns the current corpus + trending data for handlers to use.
// Safe for concurrent use.
func (c *CorpusCache) Snapshot() (*Corpus, map[string]float64, float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.corpus, c.trendingScores, c.maxTrending
}
