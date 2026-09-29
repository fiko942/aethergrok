package skills

import (
	"testing"
)

func TestCalculateSkillMatchScore(t *testing.T) {
	usingSuperpowers := Skill{
		ID:          "grok:using-superpowers",
		Name:        "using-superpowers",
		Description: "Use when starting any conversation - establishes how to find and use skills, requiring skill invocation before ANY response including clarifying questions",
		Category:    CategoryTools,
		Tags:        []string{"superpowers", "workflow"},
	}

	brainstorming := Skill{
		ID:          "grok:brainstorming",
		Name:        "brainstorming",
		Description: "You MUST use this before any creative work. Superpowers brainstorming skill.",
		Category:    CategoryTools,
		Tags:        []string{"creativity", "brainstorm"},
	}

	designTaste := Skill{
		ID:          "grok:design-taste-frontend",
		Name:        "design-taste-frontend",
		Description: "Anti-slop frontend skill for landing pages, portfolios, and redesigns.",
		Category:    CategoryFrontend,
		Tags:        []string{"design", "frontend", "ui"},
	}

	// 1. Query "superpowers":
	// using-superpowers has segment "superpowers" -> should score much higher than brainstorming (only in description)
	scoreUsing := CalculateSkillMatchScore(&usingSuperpowers, "superpowers")
	scoreBrainstorm := CalculateSkillMatchScore(&brainstorming, "superpowers")
	scoreDesign := CalculateSkillMatchScore(&designTaste, "superpowers")

	if scoreUsing <= scoreBrainstorm {
		t.Fatalf("expected using-superpowers score (%d) > brainstorming score (%d)", scoreUsing, scoreBrainstorm)
	}
	if scoreDesign != 0 {
		t.Fatalf("expected design-taste score 0 for query 'superpowers', got %d", scoreDesign)
	}

	// 2. Query "using":
	// using-superpowers starts with "using"
	scoreUsingPrefix := CalculateSkillMatchScore(&usingSuperpowers, "using")
	if scoreUsingPrefix < 700 {
		t.Fatalf("expected prefix score >= 700, got %d", scoreUsingPrefix)
	}

	// 3. Query "design":
	// design-taste-frontend starts with "design"
	scoreDesignPrefix := CalculateSkillMatchScore(&designTaste, "design")
	if scoreDesignPrefix < 700 {
		t.Fatalf("expected design prefix score >= 700, got %d", scoreDesignPrefix)
	}
}

func TestSearchRanking(t *testing.T) {
	reg := &Registry{
		skills: []Skill{
			{
				Name:        "brainstorming",
				Description: "Superpowers creative brainstorming skill for starting projects.",
			},
			{
				Name:        "using-superpowers",
				Description: "Establishes how to find and use skills.",
			},
			{
				Name:        "writing-plans",
				Description: "Superpowers plan writing skill.",
			},
		},
		indexed: true,
	}

	results := reg.Search("superpowers", "")
	if len(results) == 0 {
		t.Fatalf("expected results for 'superpowers'")
	}

	// "using-superpowers" must be the #1 result because "superpowers" is in its name
	if results[0].Name != "using-superpowers" {
		t.Fatalf("expected #1 result to be 'using-superpowers', got '%s'", results[0].Name)
	}
}
