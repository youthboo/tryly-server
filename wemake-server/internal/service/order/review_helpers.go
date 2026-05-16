package order

import (
	"strings"

	"github.com/yourusername/wemake/internal/domain"
)

const maxReviewImages = 5

func normalizeReviewImageURLs(values domain.StringArray) domain.StringArray {
	seen := make(map[string]struct{})
	out := make(domain.StringArray, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
