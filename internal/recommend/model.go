package recommend

import "time"

type Category string

const (
	CategoryGo          Category = "go"
	CategoryAIAgent     Category = "ai_agent"
	CategoryDevOpsInfra Category = "devops_infra"
	CategoryGeneral     Category = "general"
)

type Repository struct {
	FullName        string
	Name            string
	Description     string
	HTMLURL         string
	Language        string
	Topics          []string
	Stars           int
	Forks           int
	Watchers        int
	OpenIssues      int
	Archived        bool
	Fork            bool
	LicenseName     string
	Homepage        string
	PushedAt        time.Time
	UpdatedAt       time.Time
	CategoryHints   []Category
	SearchRankBoost float64
}

type Recommendation struct {
	Repository Repository
	Categories []Category
	Score      float64
	Reason     string
}
