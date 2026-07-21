package main

import (
	"fmt"
	"sort"
)

const maxMatrixCombinations = 256

// MatrixConfig models a GitHub Actions strategy.matrix (simulator).
type MatrixConfig struct {
	Dimensions map[string][]string   `json:"dimensions"`
	Include    []map[string]string   `json:"include"`
	Exclude    []map[string]string   `json:"exclude"`
}

// JobCombination is one expanded matrix job.
type JobCombination struct {
	Values map[string]string
}

// ExpandMatrix performs Cartesian product, then include/exclude.
func ExpandMatrix(config MatrixConfig) ([]JobCombination, error) {
	if len(config.Dimensions) == 0 {
		return []JobCombination{{Values: map[string]string{}}}, nil
	}

	keys := make([]string, 0, len(config.Dimensions))
	for k := range config.Dimensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	combos := []map[string]string{{}}
	for _, key := range keys {
		vals := config.Dimensions[key]
		if len(vals) == 0 {
			continue
		}
		next := make([]map[string]string, 0, len(combos)*len(vals))
		for _, base := range combos {
			for _, v := range vals {
				m := make(map[string]string, len(base)+1)
				for bk, bv := range base {
					m[bk] = bv
				}
				m[key] = v
				next = append(next, m)
			}
		}
		combos = next
		if len(combos) > maxMatrixCombinations {
			return nil, fmt.Errorf("matrix too large: %d combinations exceed %d", len(combos), maxMatrixCombinations)
		}
	}

	for _, inc := range config.Include {
		cp := make(map[string]string, len(inc))
		for k, v := range inc {
			cp[k] = v
		}
		combos = append(combos, cp)
	}
	if len(combos) > maxMatrixCombinations {
		return nil, fmt.Errorf("matrix too large: %d combinations exceed %d", len(combos), maxMatrixCombinations)
	}

	filtered := make([]map[string]string, 0, len(combos))
	for _, c := range combos {
		if matchesAny(c, config.Exclude) {
			continue
		}
		filtered = append(filtered, c)
	}

	out := make([]JobCombination, len(filtered))
	for i, c := range filtered {
		out[i] = JobCombination{Values: c}
	}
	return out, nil
}

func matchesAny(job map[string]string, filters []map[string]string) bool {
	for _, f := range filters {
		if matchesFilter(job, f) {
			return true
		}
	}
	return false
}

func matchesFilter(job, filter map[string]string) bool {
	if len(filter) == 0 {
		return false
	}
	for k, v := range filter {
		if job[k] != v {
			return false
		}
	}
	return true
}
