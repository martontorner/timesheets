package entries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/tornermarton/timesheets/internal/arrays"
	"github.com/tornermarton/timesheets/internal/utils"
)

type TimeEntrySource interface {
	PullTimeEntries(from time.Time, till time.Time) ([]TimeEntry, error)
}

type TogglTrackDefaults struct {
	Description string
}
type TogglTrack struct {
	Workspace int

	Token   string
	Url     url.URL
	Timeout time.Duration
	Ca      *string

	Defaults TogglTrackDefaults
}

type togglTrackEntry struct {
	Workspace   int      `json:"workspace_id"`
	Start       string   `json:"start"`
	Stop        *string  `json:"stop"`
	Duration    float64  `json:"duration"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func (t *TogglTrack) getEntries(from time.Time, till time.Time) ([]togglTrackEntry, error) {
	client, err := utils.CreateHttpClient(t.Timeout, t.Ca)
	if err != nil {
		return nil, err
	}

	fromStr := url.QueryEscape(from.Format(time.RFC3339))
	tillStr := url.QueryEscape(till.Format(time.RFC3339))
	reference, err := url.Parse(fmt.Sprintf("/api/v9/me/time_entries?start_date=%s&end_date=%s", fromStr, tillStr))
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("GET", t.Url.ResolveReference(reference).String(), nil)
	if err != nil {
		return nil, err
	}

	request.SetBasicAuth(t.Token, "api_token")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot get TogglTrack entries (%d)", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var togglTrackEntries []togglTrackEntry
	if err := json.Unmarshal(responseBody, &togglTrackEntries); err != nil {
		return nil, err
	}

	return togglTrackEntries, nil
}

func (t *TogglTrack) convertEntry(entry togglTrackEntry) (TimeEntry, error) {
	re := regexp.MustCompile(`\[([A-Za-z\d\-]+)]`)
	match := re.FindStringSubmatch(entry.Description)

	from, err := time.Parse(time.RFC3339, entry.Start)
	if err != nil {
		return TimeEntry{}, err
	}
	till := from.Add(time.Duration(entry.Duration) * time.Second)

	var issue string
	var description string
	if len(match) > 1 {
		issue = match[1]
		description = strings.TrimSpace(strings.Replace(entry.Description, "["+issue+"]", "", 1))
	} else {
		issue = entry.Description
		description = t.Defaults.Description
	}

	return TimeEntry{
		Issue:       issue,
		From:        from,
		Till:        till,
		Description: description,
		Tags:        entry.Tags,
	}, nil
}

func (t *TogglTrack) PullTimeEntries(from time.Time, till time.Time) ([]TimeEntry, error) {
	entries, err := t.getEntries(from, till)
	if err != nil {
		return nil, err
	}

	entries = arrays.Filter(entries, func(entry togglTrackEntry) bool { return entry.Workspace == t.Workspace && entry.Stop != nil })

	return arrays.MapE(entries, func(entry togglTrackEntry) (TimeEntry, error) { return t.convertEntry(entry) })
}

func createTogglTrack(spec map[string]any) (*TogglTrack, error) {
	var workspace int
	if workspaceParam, ok := spec["workspace"].(int); ok {
		workspace = workspaceParam
	} else {
		return nil, fmt.Errorf("invalid or missing 'workspace' spec for TogglTrack source")
	}

	var token string
	if tokenParam, ok := spec["token"].(string); ok && tokenParam != "" {
		token = tokenParam
	} else {
		return nil, fmt.Errorf("invalid or missing 'token' spec for TogglTrack source")
	}

	var url = url.URL{
		Scheme: "https",
		Host:   "api.track.toggl.com",
	}
	if urlParam, ok := spec["url"].(string); ok {
		url_, err := url.Parse(urlParam)
		if err != nil {
			return nil, fmt.Errorf("invalid 'url' spec for TogglTrack source: %w", err)
		}
		url = *url_
	}

	var timeout = 10 * time.Second
	if timeoutParam, ok := spec["timeout"].(string); ok {
		timeout_, err := time.ParseDuration(timeoutParam)
		if err != nil {
			return nil, fmt.Errorf("invalid 'timeout' spec for TogglTrack source: %w", err)
		}
		timeout = timeout_
	}

	var ca *string = nil
	if caParam, ok := spec["ca"].(string); ok {
		ca = &caParam
	}

	var defaults = map[string]any{}
	if defaultsParam, ok := spec["defaults"].(map[string]any); ok {
		defaults = defaultsParam
	}

	var description = ""
	if descriptionParam, ok := defaults["description"].(string); ok {
		description = descriptionParam
	}

	return &TogglTrack{
		Workspace: workspace,

		Token:   token,
		Url:     url,
		Timeout: timeout,
		Ca:      ca,

		Defaults: TogglTrackDefaults{
			Description: description,
		},
	}, nil
}

type TimeEntrySourceConfig struct {
	Kind string         `yaml:"kind"`
	Spec map[string]any `yaml:"spec"`
}

func NewTimeEntrySource(config TimeEntrySourceConfig) (TimeEntrySource, error) {
	switch config.Kind {
	case "TogglTrack":
		return createTogglTrack(config.Spec)
	default:
		return nil, fmt.Errorf("unsupported time entry source: %s", config.Kind)
	}
}
