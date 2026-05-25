package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"recommend-bot/internal/recommend"
)

type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:   strings.TrimSpace(token),
		baseURL: "https://api.github.com",
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) SearchCandidates(ctx context.Context, minStars int, lookbackDays int) ([]recommend.Repository, error) {
	if lookbackDays <= 0 {
		lookbackDays = 30
	}
	pushedAfter := time.Now().UTC().AddDate(0, 0, -lookbackDays).Format("2006-01-02")

	queries := []querySpec{
		{Query: fmt.Sprintf("language:Go stars:>%d archived:false fork:false", minStars), Hint: recommend.CategoryGo, Boost: 30},
		{Query: fmt.Sprintf("topic:golang stars:>%d archived:false fork:false", minStars), Hint: recommend.CategoryGo, Boost: 25},
		{Query: fmt.Sprintf("language:Go pushed:>=%s stars:>%d archived:false fork:false", pushedAfter, minStars/2), Hint: recommend.CategoryGo, Boost: 35},
		{Query: fmt.Sprintf("topic:llm stars:>%d archived:false fork:false", minStars), Hint: recommend.CategoryAIAgent, Boost: 28},
		{Query: fmt.Sprintf("topic:agents stars:>%d archived:false fork:false", minStars/2), Hint: recommend.CategoryAIAgent, Boost: 30},
		{Query: fmt.Sprintf("topic:rag stars:>%d archived:false fork:false", minStars/2), Hint: recommend.CategoryAIAgent, Boost: 24},
		{Query: fmt.Sprintf("topic:mcp stars:>%d archived:false fork:false", minStars/2), Hint: recommend.CategoryAIAgent, Boost: 28},
		{Query: fmt.Sprintf("topic:kubernetes stars:>%d archived:false fork:false", minStars), Hint: recommend.CategoryDevOpsInfra, Boost: 25},
		{Query: fmt.Sprintf("topic:observability stars:>%d archived:false fork:false", minStars/2), Hint: recommend.CategoryDevOpsInfra, Boost: 24},
		{Query: fmt.Sprintf("topic:devops stars:>%d archived:false fork:false", minStars/2), Hint: recommend.CategoryDevOpsInfra, Boost: 22},
	}

	var all []recommend.Repository
	for _, spec := range queries {
		repos, err := c.search(ctx, spec)
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
	}
	return all, nil
}

type querySpec struct {
	Query string
	Hint  recommend.Category
	Boost float64
}

func (c *Client) search(ctx context.Context, spec querySpec) ([]recommend.Repository, error) {
	endpoint, err := url.Parse(c.baseURL + "/search/repositories")
	if err != nil {
		return nil, err
	}
	values := endpoint.Query()
	values.Set("q", spec.Query)
	values.Set("sort", "stars")
	values.Set("order", "desc")
	values.Set("per_page", "20")
	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github search request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github search failed for %q: %s", spec.Query, resp.Status)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode github search response: %w", err)
	}

	repos := make([]recommend.Repository, 0, len(result.Items))
	for _, item := range result.Items {
		repos = append(repos, item.toRepository(spec))
	}
	return repos, nil
}

type searchResponse struct {
	Items []repositoryItem `json:"items"`
}

type repositoryItem struct {
	FullName    string       `json:"full_name"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	HTMLURL     string       `json:"html_url"`
	Language    string       `json:"language"`
	Topics      []string     `json:"topics"`
	Stars       int          `json:"stargazers_count"`
	Forks       int          `json:"forks_count"`
	Watchers    int          `json:"watchers_count"`
	OpenIssues  int          `json:"open_issues_count"`
	Archived    bool         `json:"archived"`
	Fork        bool         `json:"fork"`
	License     *licenseItem `json:"license"`
	Homepage    string       `json:"homepage"`
	PushedAt    time.Time    `json:"pushed_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type licenseItem struct {
	Name string `json:"name"`
}

func (i repositoryItem) toRepository(spec querySpec) recommend.Repository {
	license := ""
	if i.License != nil {
		license = i.License.Name
	}
	return recommend.Repository{
		FullName:        i.FullName,
		Name:            i.Name,
		Description:     i.Description,
		HTMLURL:         i.HTMLURL,
		Language:        i.Language,
		Topics:          i.Topics,
		Stars:           i.Stars,
		Forks:           i.Forks,
		Watchers:        i.Watchers,
		OpenIssues:      i.OpenIssues,
		Archived:        i.Archived,
		Fork:            i.Fork,
		LicenseName:     license,
		Homepage:        i.Homepage,
		PushedAt:        i.PushedAt,
		UpdatedAt:       i.UpdatedAt,
		CategoryHints:   []recommend.Category{spec.Hint},
		SearchRankBoost: spec.Boost,
	}
}
