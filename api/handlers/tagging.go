package handlers

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/example/cis-benchmark-intelligence/api/models"
)

var (
	versionFromNamePattern = regexp.MustCompile(`(?i)(?:^|[\s_-])v(?:ersion)?[\s_-]*(\d+(?:\.\d+)*)`)
	trailingVersionPattern = regexp.MustCompile(`(?i)(\d+(?:\.\d+)*)$`)
	versionStripPattern    = regexp.MustCompile(`(?i)(?:^|[\s_-])v(?:ersion)?[\s_-]*\d+(?:\.\d+)*`)
	multiSpacePattern      = regexp.MustCompile(`\s+`)
	nonAlphaNumericPattern = regexp.MustCompile(`[^a-z0-9 ]+`)
)

func normalizeNameText(raw string) string {
	normalized := strings.ReplaceAll(raw, "_", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.TrimSpace(multiSpacePattern.ReplaceAllString(normalized, " "))
	return normalized
}

func deriveVersionFromFilename(filename string) string {
	stem := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	normalized := normalizeNameText(stem)

	matches := versionFromNamePattern.FindAllStringSubmatch(normalized, -1)
	if len(matches) > 0 {
		last := matches[len(matches)-1]
		if len(last) > 1 {
			return strings.TrimSpace(last[1])
		}
	}

	tail := trailingVersionPattern.FindStringSubmatch(normalized)
	if len(tail) > 1 {
		return strings.TrimSpace(tail[1])
	}
	return ""
}

func deriveBenchmarkNameFromFilename(filename string) string {
	stem := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	normalized := normalizeNameText(stem)
	withoutVersion := strings.TrimSpace(versionStripPattern.ReplaceAllString(normalized, " "))
	withoutVersion = normalizeNameText(withoutVersion)
	return withoutVersion
}

func canonicalizeForSimilarity(text string) string {
	lowered := strings.ToLower(strings.TrimSpace(text))
	clean := nonAlphaNumericPattern.ReplaceAllString(lowered, " ")
	return strings.TrimSpace(multiSpacePattern.ReplaceAllString(clean, " "))
}

func levenshteinDistance(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		previous[j] = j
	}

	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			insertion := current[j-1] + 1
			deletion := previous[j] + 1
			substitution := previous[j-1] + cost
			current[j] = minInt(insertion, deletion, substitution)
		}
		copy(previous, current)
	}

	return previous[len(b)]
}

func minInt(values ...int) int {
	minimum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}

func similarityPercent(left, right string) float64 {
	a := canonicalizeForSimilarity(left)
	b := canonicalizeForSimilarity(right)
	if a == "" && b == "" {
		return 100
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 100
	}
	distance := levenshteinDistance(a, b)
	score := (1.0 - (float64(distance) / float64(maxLen))) * 100
	if score < 0 {
		return 0
	}
	return score
}

func (h *Handler) resolveFrameworkName(requestedFramework, filename string) (string, float64, bool) {
	candidate := strings.TrimSpace(requestedFramework)
	if candidate == "" || strings.EqualFold(candidate, "CIS Controls") {
		derived := deriveBenchmarkNameFromFilename(filename)
		if derived != "" {
			candidate = derived
		}
	}
	if candidate == "" {
		candidate = "CIS Controls"
	}

	existing := []models.Framework{}
	if err := h.DB.Find(&existing).Error; err != nil || len(existing) == 0 {
		return candidate, 0, false
	}

	bestName := ""
	bestScore := 0.0
	for _, framework := range existing {
		score := similarityPercent(candidate, framework.Name)
		if score > bestScore {
			bestScore = score
			bestName = framework.Name
		}
	}

	if bestName != "" && bestScore >= 95 {
		return bestName, bestScore, true
	}
	return candidate, bestScore, false
}
