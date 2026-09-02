package scaffold

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
)

// Skill is one design-time skill the scaffold writes into a customer repo.
//
// They are searchable for the same reason the wiki is. An agent asked to build
// a deck reaches for `asgard-cli find slides`, and until this existed the answer
// was that nothing matched anywhere - while the skill that owns the whole
// subject sat in the repository it was standing in. The material was there and
// the tool's own way in did not reach it.
type Skill struct {
	Name        string
	Description string
	Path        string
}

const skillRoot = "templates/.agents/skills"

var frontmatterField = regexp.MustCompile(`(?m)^(name|description):\s*(.*)$`)

// Skills lists the design-time skills, sorted by name.
func Skills() ([]Skill, error) {
	entries, err := fs.ReadDir(templates, skillRoot)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", skillRoot, err)
	}

	var out []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		body, err := skillBody(e.Name())
		if err != nil {
			return nil, err
		}
		s := Skill{Name: e.Name(), Path: ".agents/skills/" + e.Name() + "/SKILL.md"}
		for _, m := range frontmatterField.FindAllStringSubmatch(body, -1) {
			if m[1] == "description" {
				s.Description = strings.TrimSpace(m[2])
			}
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// skillBody reads one skill's SKILL.md, rendered or not. The templated ones are
// searched as written, placeholders and all: a search hits prose, and the only
// placeholder in them is the workspace name.
func skillBody(name string) (string, error) {
	for _, suffix := range []string{"/SKILL.md.tmpl", "/SKILL.md"} {
		data, err := templates.ReadFile(skillRoot + "/" + name + suffix)
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("%s has no SKILL.md", name)
}

// SkillMatch is one skill covering the search.
type SkillMatch struct {
	Skill
	Lines []string
	Score int
	Terms []string
}

// SearchSkills finds skills covering the given terms, on the same two-pass rule
// as the other bodies: all of them, then any of them.
func SearchSkills(query string) ([]SkillMatch, error) {
	terms := kb.Terms(query)
	if len(terms) == 0 {
		return nil, fmt.Errorf("search needs at least one term")
	}

	all, err := Skills()
	if err != nil {
		return nil, err
	}

	var matches []SkillMatch
	for _, s := range all {
		body, err := skillBody(s.Name)
		if err != nil {
			return nil, err
		}
		lower := strings.ToLower(body)

		m := SkillMatch{Skill: s}
		for _, term := range terms {
			if kb.Covers(lower, term) {
				m.Terms = append(m.Terms, term)
			}
		}
		if len(m.Terms) == 0 {
			continue
		}

		for _, line := range strings.Split(skipFrontmatter(body), "\n") {
			lowerLine := strings.ToLower(line)
			for _, term := range m.Terms {
				if !kb.Covers(lowerLine, term) {
					continue
				}
				m.Score++
				// The description is printed above from the frontmatter, and a
				// line still carrying `<<...>>` is a template placeholder that
				// reads as broken text in a result.
				trimmed := strings.TrimSpace(line)
				if len(m.Lines) < 3 && len(trimmed) > 20 && !strings.Contains(trimmed, "<<") {
					m.Lines = append(m.Lines, trimmed)
				}
				break
			}
		}
		matches = append(matches, m)
	}

	sort.Slice(matches, func(i, j int) bool {
		if len(matches[i].Terms) != len(matches[j].Terms) {
			return len(matches[i].Terms) > len(matches[j].Terms)
		}
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > 0 && len(matches[0].Terms) == len(terms) {
		for i, m := range matches {
			if len(m.Terms) < len(terms) {
				return matches[:i], nil
			}
		}
	}
	return matches, nil
}

// skipFrontmatter drops the leading --- block, so a search quotes the skill's
// prose rather than the description it is already being shown beside.
func skipFrontmatter(body string) string {
	if !strings.HasPrefix(body, "---\n") {
		return body
	}
	if i := strings.Index(body[4:], "\n---\n"); i >= 0 {
		return body[4+i+5:]
	}
	return body
}
