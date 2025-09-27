package entries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math"
	"net/http"
	"net/url"
	"strings"
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
	Url     url.URL
	Timeout time.Duration
	Ca      *string

	Tags CapsysKronosTags

	Defaults CapsysKronosDefaults
}

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
	// Kronos is unaware of timezones so it must be converted to its default
	tz, _ := time.LoadLocation("Europe/Budapest")

	worklogInput := capsysKronosTimeEntryWorklogInput{
		IssueKey:            entry.Issue,
		TimeSpent:           math.Floor(entry.Till.Sub(entry.From).Minutes()),
		StartOffsetDateTime: entry.From.Truncate(time.Minute).In(tz).Format(time.RFC3339),
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

	reference, err := url.Parse(fmt.Sprintf("/rest/api/latest/issue/%s", entry.Issue))
	if err != nil {
		return err
	}

	request, err := http.NewRequest("GET", c.Url.ResolveReference(reference).String(), nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

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

	reference, err := url.Parse("/rest/kronos/1.0/log-entry")
	if err != nil {
		return err
	}

	requestBody, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", c.Url.ResolveReference(reference).String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+c.Token)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("could not create CapsysKronos entry (%d): %s", response.StatusCode, strings.ReplaceAll(string(responseBody), "\n", ""))
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

func createCapsysKronos(spec map[string]any) (*CapsysKronos, error) {
	var token string
	if tokenParam, ok := spec["token"].(string); ok && tokenParam != "" {
		token = tokenParam
	} else {
		return nil, fmt.Errorf("invalid or missing 'token' spec for CapsysKronos target")
	}

	var url = url.URL{
		Scheme: "https",
		Host:   "jira.capsys.hu",
	}
	if urlParam, ok := spec["url"].(string); ok {
		url_, err := url.Parse(urlParam)
		if err != nil {
			return nil, fmt.Errorf("invalid 'url' spec for CapsysKronos target: %w", err)
		}
		url = *url_
	}

	var timeout = 10 * time.Second
	if timeoutParam, ok := spec["timeout"].(string); ok {
		timeoutDuration, err := time.ParseDuration(timeoutParam)
		if err != nil {
			return nil, fmt.Errorf("invalid 'timeout' spec for CapsysKronos target: %w", err)
		}
		timeout = timeoutDuration
	}

	var ca *string = nil
	if caParam, ok := spec["ca"].(string); ok {
		ca = &caParam
	}

	var tags = CapsysKronosTags{}
	if tagsParam, ok := spec["tags"].(map[string]map[string]any); ok {
		tags = CapsysKronosTags(tagsParam)
	}

	var defaults = map[string]any{}
	if defaultsParam, ok := spec["defaults"].(map[string]any); ok {
		defaults = defaultsParam
	}

	var activityCategoryId = 3
	if activityCategoryIdParam, ok := defaults["activityCategoryId"].(int); ok {
		activityCategoryId = activityCategoryIdParam
	}

	var activityTypeId = 5
	if activityTypeIdParam, ok := defaults["activityTypeId"].(int); ok {
		activityTypeId = activityTypeIdParam
	}

	var siteId = 31
	if siteIdParam, ok := defaults["siteId"].(int); ok {
		siteId = siteIdParam
	}

	var comment = ""
	if commentParam, ok := defaults["comment"].(string); ok {
		comment = commentParam
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
	Spec map[string]any `yaml:"spec"`
}

func NewTimeEntryTarget(config TimeEntryTargetConfig) (TimeEntryTarget, error) {
	switch config.Kind {
	case "CapsysKronos":
		return createCapsysKronos(config.Spec)
	default:
		return nil, fmt.Errorf("unsupported time entry target: %s", config.Kind)
	}
}
