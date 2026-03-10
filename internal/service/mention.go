package service

import (
	"regexp"
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

// mentionRe captures @tokens that may include hyphens, underscores, and an
// optional second word (e.g. @seo-lead, @seo_lead, @seoLead, @seo lead).
var mentionRe = regexp.MustCompile(`@([\w][\w-]*(?:\s[\w][\w-]*)?)`)

// normalizeName strips hyphens, underscores, and spaces, then lowercases.
// "seo-lead" / "seo_lead" / "seoLead" / "seo lead" all become "seolead".
func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.NewReplacer("-", "", "_", "", " ", "").Replace(name)
	return name
}

// matchMentionedUsers finds all @mentions in content and returns matching
// users, skipping the author. It tries both exact (lowered) and normalized
// matching so @seo-lead, @seoLead, @seo_lead, @seo lead all resolve to
// user "seo-lead".
func matchMentionedUsers(content string, allUsers []model.User, authorID string) []model.User {
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	// Build two lookup maps: exact lowercase and normalized
	exactMap := make(map[string]model.User, len(allUsers))
	normMap := make(map[string]model.User, len(allUsers))
	for _, u := range allUsers {
		exactMap[strings.ToLower(u.Name)] = u
		normMap[normalizeName(u.Name)] = u
	}

	var mentioned []model.User
	seen := make(map[string]bool)

	for _, match := range matches {
		token := match[1]

		// Try exact lowercase match first
		u, ok := exactMap[strings.ToLower(token)]
		if !ok {
			// Try normalized match
			u, ok = normMap[normalizeName(token)]
		}
		if !ok || u.ID == authorID || seen[u.ID] {
			continue
		}
		seen[u.ID] = true
		mentioned = append(mentioned, u)
	}

	return mentioned
}
