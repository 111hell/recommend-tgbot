package recommend

import "strings"

func DetectCategories(repo Repository) []Category {
	text := strings.ToLower(strings.Join(append([]string{
		repo.FullName,
		repo.Name,
		repo.Description,
		repo.Language,
	}, repo.Topics...), " "))

	var categories []Category
	if repo.Language == "Go" || containsAny(text, " golang", " go ", "go-framework", "go-zero", "grpc", "cobra") {
		categories = append(categories, CategoryGo)
	}
	if containsAny(text, "llm", "agent", "agents", "rag", "mcp", "openai", "prompt", "workflow automation", "eval") {
		categories = append(categories, CategoryAIAgent)
	}
	if containsAny(text, "kubernetes", "observability", "devops", "ci/cd", "ci-cd", "terraform", "deployment", "prometheus", "grafana", "infra") {
		categories = append(categories, CategoryDevOpsInfra)
	}
	for _, hint := range repo.CategoryHints {
		if !categoryExists(categories, hint) {
			categories = append(categories, hint)
		}
	}
	if len(categories) == 0 {
		categories = append(categories, CategoryGeneral)
	}
	return categories
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func categoryExists(categories []Category, target Category) bool {
	for _, category := range categories {
		if category == target {
			return true
		}
	}
	return false
}
