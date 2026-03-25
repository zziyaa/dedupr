package dedup

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"sync"

	"github.com/zeebo/xxh3"
)

const partialHashSize = 4 * 1024         // 4 KB
const partialHashThreshold = 1024 * 1024 // 1 MB

// Group is a set of files with identical content (≥2 paths).
type Group struct {
	Size  int64    // bytes
	Hash  string   // hex xxHash3-128 of full content
	Paths []string // sorted alphabetically
}

type hashResult struct {
	path string
	hash string
	err  error
}

// Find takes file paths and returns groups of duplicates.
// Returns an error if any path cannot be stat'd or read, or if ctx is cancelled.
func Find(ctx context.Context, paths []string) ([]Group, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	// Three-phase algorithm:
	//  1. Group by size — stat every file and bucket by byte count; groups smaller
	//     than 2 are dropped immediately.
	//  2. Partial-hash filter (large files only) — for files at or above
	//     partialHashThreshold, hash only the first partialHashSize bytes and drop
	//     any file whose prefix hash is unique. This cheaply eliminates non-duplicates
	//     before the expensive full read.
	//  3. Full-hash grouping — SHA-256 the surviving candidates in full; paths that
	//     share a hash form a duplicate group.

	// Phase 1: group by size
	sizeGroups, err := groupBySize(ctx, paths)
	if err != nil {
		return nil, err
	}

	var result []Group
	for size, sizeGroup := range sizeGroups {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Phase 2: partial hash filter for large files
		candidates := sizeGroup
		if size >= partialHashThreshold {
			filtered, err := filterByPartialHash(ctx, sizeGroup)
			if err != nil {
				return nil, err
			}
			candidates = filtered
			if len(candidates) < 2 {
				continue
			}
		}

		// Phase 3: full hash
		hashGroups, err := groupByFullHash(ctx, candidates)
		if err != nil {
			return nil, err
		}
		for hash, group := range hashGroups {
			if len(group) < 2 {
				continue
			}
			sort.Strings(group)
			result = append(result, Group{
				Size:  size,
				Hash:  hash,
				Paths: group,
			})
		}
	}

	// Sort groups: size descending, then first path ascending
	sort.Slice(result, func(i, j int) bool {
		if result[i].Size != result[j].Size {
			return result[i].Size > result[j].Size
		}
		return result[i].Paths[0] < result[j].Paths[0]
	})

	return result, nil
}

// groupBySize stats all paths and returns buckets of paths with equal sizes
// that have ≥2 members. Returns an error on any stat failure or context cancellation.
func groupBySize(ctx context.Context, paths []string) (map[int64][]string, error) {
	buckets := make(map[int64][]string)
	for _, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", p, err)
		}
		buckets[info.Size()] = append(buckets[info.Size()], p)
	}
	result := make(map[int64][]string)
	for size, group := range buckets {
		if len(group) >= 2 {
			result[size] = group
		}
	}
	return result, nil
}

// filterByPartialHash hashes the first 4 KB of each file and returns only
// paths that share a partial hash with at least one other file.
func filterByPartialHash(ctx context.Context, paths []string) ([]string, error) {
	results := computeHashes(ctx, paths, func(p string) (string, error) {
		f, err := os.Open(p)
		if err != nil {
			return "", fmt.Errorf("open %s: %w", p, err)
		}
		defer f.Close()
		h, err := hashContent(f, true)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", p, err)
		}
		return h, nil
	})
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
	}

	buckets := make(map[string][]string)
	for _, r := range results {
		buckets[r.hash] = append(buckets[r.hash], r.path)
	}

	var survivors []string
	for _, group := range buckets {
		if len(group) >= 2 {
			survivors = append(survivors, group...)
		}
	}
	return survivors, nil
}

// groupByFullHash computes a full xxHash3-128 for each path and returns buckets.
func groupByFullHash(ctx context.Context, paths []string) (map[string][]string, error) {
	results := computeHashes(ctx, paths, func(p string) (string, error) {
		f, err := os.Open(p)
		if err != nil {
			return "", fmt.Errorf("open %s: %w", p, err)
		}
		defer f.Close()
		h, err := hashContent(f, false)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", p, err)
		}
		return h, nil
	})
	buckets := make(map[string][]string)
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		buckets[r.hash] = append(buckets[r.hash], r.path)
	}
	return buckets, nil
}

// computeHashes runs hashFn concurrently over paths using a worker pool.
// It stops launching new work if ctx is cancelled.
func computeHashes(ctx context.Context, paths []string, hashFn func(string) (string, error)) []hashResult {
	sem := make(chan struct{}, runtime.NumCPU())
	results := make([]hashResult, len(paths))
	var wg sync.WaitGroup
	for i, p := range paths {
		select {
		case <-ctx.Done():
			// Fill unstarted entries with the cancellation error and wait for in-flight goroutines.
			for j := i; j < len(paths); j++ {
				results[j] = hashResult{path: paths[j], err: ctx.Err()}
			}
			wg.Wait()
			return results
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			defer func() { <-sem }()
			h, err := hashFn(path)
			results[idx] = hashResult{path: path, hash: h, err: err}
		}(i, p)
	}
	wg.Wait()
	return results
}

// hashContent computes an xxHash3-128 digest of r.
// If partial is true, only the first partialHashSize bytes are read;
// a reader shorter than partialHashSize is consumed in full (not an error).
// If partial is false, r is read to EOF.
func hashContent(r io.Reader, partial bool) (string, error) {
	h := xxh3.New128()
	if partial {
		if _, err := io.CopyN(h, r, partialHashSize); err != nil && err != io.EOF {
			return "", err
		}
	} else {
		if _, err := io.Copy(h, r); err != nil {
			return "", err
		}
	}
	sum := h.Sum128()
	return fmt.Sprintf("%016x%016x", sum.Hi, sum.Lo), nil
}
