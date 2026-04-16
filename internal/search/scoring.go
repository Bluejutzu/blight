package search

import "strings"

// BestScore returns the best fuzzy/exact score across the provided targets and
// keywords for a given query. It is used by the launcher's cross-provider
// ranking layer.
func BestScore(query string, targets []string, keywords []string) int {
	queryNorm := strings.ToLower(strings.TrimSpace(query))
	if queryNorm == "" {
		return 0
	}

	best := 0
	for _, target := range targets {
		targetNorm := strings.ToLower(strings.TrimSpace(target))
		if targetNorm == "" {
			continue
		}
		if s := score(queryNorm, targetNorm); s > best {
			best = s
		}
	}

	for _, keyword := range keywords {
		keywordNorm := strings.ToLower(strings.TrimSpace(keyword))
		if keywordNorm == "" {
			continue
		}
		switch {
		case keywordNorm == queryNorm:
			if 9000 > best {
				best = 9000
			}
		case strings.HasPrefix(keywordNorm, queryNorm):
			if s := 6500 + len(queryNorm)*10; s > best {
				best = s
			}
		case strings.Contains(keywordNorm, queryNorm):
			if s := 2600 + len(queryNorm)*5; s > best {
				best = s
			}
		}
	}

	return best
}
