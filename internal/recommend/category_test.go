package recommend

import "testing"

func TestDetectCategoriesPrefersGoMetadata(t *testing.T) {
	repo := Repository{
		FullName:    "gin-gonic/gin",
		Language:    "Go",
		Description: "Gin is a HTTP web framework written in Go.",
		Topics:      []string{"go", "golang", "framework"},
	}

	categories := DetectCategories(repo)

	if !hasCategory(categories, CategoryGo) {
		t.Fatalf("DetectCategories() = %v, want %s", categories, CategoryGo)
	}
}

func TestDetectCategoriesFindsAIAgentProjects(t *testing.T) {
	repo := Repository{
		FullName:    "owner/agent-kit",
		Language:    "Python",
		Description: "LLM agent framework with RAG workflow automation.",
		Topics:      []string{"llm", "agents", "rag"},
	}

	categories := DetectCategories(repo)

	if !hasCategory(categories, CategoryAIAgent) {
		t.Fatalf("DetectCategories() = %v, want %s", categories, CategoryAIAgent)
	}
}

func TestDetectCategoriesFindsDevOpsInfraProjects(t *testing.T) {
	repo := Repository{
		FullName:    "owner/deploy-tool",
		Language:    "Go",
		Description: "Kubernetes observability and CI/CD deployment toolkit.",
		Topics:      []string{"kubernetes", "observability", "devops"},
	}

	categories := DetectCategories(repo)

	if !hasCategory(categories, CategoryDevOpsInfra) {
		t.Fatalf("DetectCategories() = %v, want %s", categories, CategoryDevOpsInfra)
	}
}

func hasCategory(categories []Category, want Category) bool {
	for _, category := range categories {
		if category == want {
			return true
		}
	}
	return false
}
