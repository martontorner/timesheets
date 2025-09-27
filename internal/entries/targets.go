package entries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/tornermarton/timesheets/internal/utils"
)

type TimeEntryTarget interface {
	PushTimeEntry(entry TimeEntry) error
}

type CapsysKronosTags map[string]map[string]any
type CapsysKronosDefaults struct {
	ActivityCategoryId int
	ActivityTypeId     int
	SiteId             int
	Comment            string
}
type CapsysKronos struct {
	Token   string
	Url     string
	Timeout time.Duration
	Ca      *string

	Tags CapsysKronosTags

	Defaults CapsysKronosDefaults
}

const (
	capsysKronosUrl     = "https://jira.capsys.hu"
	capsysKronosTimeout = 10 * time.Second

	capsysKronosDefaultsActivityCategoryId = 3
	capsysKronosDefaultsActivityTypeId     = 5
	capsysKronosDefaultsSiteId             = 31
	capsysKronosDefaultsComment            = ""
)

type capsysKronosTimeEntryWorklogInput struct {
	IssueKey            string  `json:"issueKey"`
	TimeSpent           float64 `json:"timeSpent"`
	StartOffsetDateTime string  `json:"startOffsetDateTime"`
	Comment             string  `json:"comment"`
	ActivityCategoryId  int     `json:"activityCategoryId"`
	ActivityTypeId      int     `json:"activityTypeId"`
	SiteId              int     `json:"siteId"`
}
type capsysKronosTimeEntryTravelInput struct {
	TravelToTimeSpentInMinutes   int  `json:"travelToTimeSpentInMinutes"`
	TravelFromTimeSpentInMinutes int  `json:"travelFromTimeSpentInMinutes"`
	FromSiteId                   *int `json:"fromSiteId"`
}
type capsysKronosTimeEntry struct {
	WorklogInput capsysKronosTimeEntryWorklogInput `json:"worklogInput"`
	TravelInput  capsysKronosTimeEntryTravelInput  `json:"travelInput"`
}

func (c *CapsysKronos) convertEntry(entry TimeEntry) (capsysKronosTimeEntry, error) {
	worklogInput := capsysKronosTimeEntryWorklogInput{
		IssueKey:            entry.Issue,
		TimeSpent:           math.Floor(entry.Till.Sub(entry.From).Minutes()),
		StartOffsetDateTime: entry.From.Truncate(time.Minute).Format(time.RFC3339),
		Comment:             utils.DefaultString(entry.Description, c.Defaults.Comment),
		ActivityCategoryId:  c.Defaults.ActivityCategoryId,
		ActivityTypeId:      c.Defaults.ActivityTypeId,
		SiteId:              c.Defaults.SiteId,
	}
	travelInput := capsysKronosTimeEntryTravelInput{
		TravelToTimeSpentInMinutes:   0,
		TravelFromTimeSpentInMinutes: 0,
		FromSiteId:                   nil,
	}

	var worklogInputData map[string]any
	b, _ := json.Marshal(worklogInput)
	_ = json.Unmarshal(b, &worklogInputData)
	for _, tag := range entry.Tags {
		if tagData, ok := c.Tags[tag]; ok {
			maps.Copy(worklogInputData, tagData)
		}
	}
	b, _ = json.Marshal(worklogInputData)
	_ = json.Unmarshal(b, &worklogInput)

	return capsysKronosTimeEntry{
		WorklogInput: worklogInput,
		TravelInput:  travelInput,
	}, nil
}

func (c *CapsysKronos) validateTimeEntryIssue(entry TimeEntry) error {
	client, err := utils.CreateHttpClient(c.Timeout, c.Ca)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.Url, "/rest/api/latest/issue/", entry.Issue)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+c.Token)

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("invalid CapsysKronos issue (%d)", response.StatusCode)
	}

	return nil

}

func (c *CapsysKronos) postEntry(entry capsysKronosTimeEntry) error {
	client, err := utils.CreateHttpClient(c.Timeout, c.Ca)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.Url, "/rest/kronos/1.0/log-entry")
	if err != nil {
		return err
	}

	body, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+c.Token)

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("could not create CapsysKronos entry (%d)", response.StatusCode)
	}

	return nil
}

func (c *CapsysKronos) PushTimeEntry(entry TimeEntry) error {
	if err := c.validateTimeEntryIssue(entry); err != nil {
		return err
	}

	entry_, err := c.convertEntry(entry)
	if err != nil {
		return err
	}

	return c.postEntry(entry_)
}

func createCapsysKronos(data map[string]any) (*CapsysKronos, error) {
	var token string
	if tokenParam, ok := data["token"].(string); ok && tokenParam != "" {
		token = tokenParam
	} else {
		return nil, fmt.Errorf("invalid or missing 'token' parameter for CapsysKronos source")
	}

	var url string
	if urlParam, ok := data["url"].(string); ok {
		url = urlParam
	} else {
		url = capsysKronosUrl
	}

	var timeout time.Duration
	if timeoutParam, ok := data["timeout"].(string); ok {
		timeoutDuration, err := time.ParseDuration(timeoutParam)
		if err != nil {
			return nil, fmt.Errorf("invalid 'timeout' parameter for CapsysKronos source: %w", err)
		}
		timeout = timeoutDuration
	} else {
		timeout = capsysKronosTimeout
	}

	var ca *string
	if caParam, ok := data["ca"].(string); ok {
		ca = &caParam
	} else {
		ca = nil
	}

	var tags CapsysKronosTags
	if tagsParam, ok := data["tags"].(map[string]map[string]any); ok {
		tags = CapsysKronosTags(tagsParam)
	} else {
		tags = CapsysKronosTags{}
	}

	var defaults map[string]any
	if defaultsParam, ok := data["defaults"].(map[string]any); ok {
		defaults = defaultsParam
	} else {
		defaults = map[string]any{}
	}

	var activityCategoryId int
	if activityCategoryIdParam, ok := defaults["activityCategoryId"].(int); ok {
		activityCategoryId = activityCategoryIdParam
	} else {
		activityCategoryId = capsysKronosDefaultsActivityCategoryId
	}

	var activityTypeId int
	if activityTypeIdParam, ok := defaults["activityTypeId"].(int); ok {
		activityTypeId = activityTypeIdParam
	} else {
		activityTypeId = capsysKronosDefaultsActivityTypeId
	}

	var siteId int
	if siteIdParam, ok := defaults["siteId"].(int); ok {
		siteId = siteIdParam
	} else {
		siteId = capsysKronosDefaultsSiteId
	}

	var comment string
	if commentParam, ok := defaults["comment"].(string); ok {
		comment = commentParam
	} else {
		comment = capsysKronosDefaultsComment
	}

	return &CapsysKronos{
		Token:   token,
		Url:     url,
		Timeout: timeout,
		Ca:      ca,

		Tags: tags,

		Defaults: CapsysKronosDefaults{
			ActivityCategoryId: activityCategoryId,
			ActivityTypeId:     activityTypeId,
			SiteId:             siteId,
			Comment:            comment,
		},
	}, nil
}

type TimeEntryTargetConfig struct {
	Kind string         `yaml:"kind"`
	Data map[string]any `yaml:"data"`
}

func NewTimeEntryTarget(config TimeEntryTargetConfig) (TimeEntryTarget, error) {
	switch config.Kind {
	case "CapsysKronos":
		return createCapsysKronos(config.Data)
	default:
		return nil, fmt.Errorf("unsupported time entry target: %s", config.Kind)
	}
}
