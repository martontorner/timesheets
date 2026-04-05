package entries

import (
	"time"

	"charm.land/lipgloss/v2"

	"github.com/tornermarton/timesheets/internal/utils"
)

var primary = lipgloss.NewStyle().Bold(true)
var secondary = lipgloss.NewStyle().Faint(true)

type TimeEntry struct {
	Issue       string
	From        time.Time
	Till        time.Time
	Description string
	Tags        []string
}

func (te TimeEntry) String(location *time.Location) string {
	return lipgloss.Sprintf(
		"%s-%s %s %s %s %s",
		te.From.In(location).Format(time.DateTime),
		te.Till.In(location).Format(time.TimeOnly),
		secondary.Render(utils.FitString(te.Till.Sub(te.From).Round(time.Minute).String(), 9)),
		primary.Render(utils.FitString(te.Issue, 12)),
		primary.Render(utils.FitString(te.Description, 20)),
		secondary.Render(utils.FitArray(te.Tags, 20)),
	)
}
