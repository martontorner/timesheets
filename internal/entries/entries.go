package entries

import (
	"fmt"
	"time"

	"github.com/tornermarton/timesheets/internal/utils"
)

type TimeEntry struct {
	Issue       string
	From        time.Time
	Till        time.Time
	Description string
	Tags        []string
}

func (te TimeEntry) String(location *time.Location) string {
	return fmt.Sprintf(
		"[%s - %s] [%s] %s %s",
		te.From.In(location).Format(time.DateTime),
		te.Till.In(location).Format(time.DateTime),
		utils.FitString(te.Issue, 12),
		utils.FitString(te.Description, 20),
		utils.FitArray(te.Tags, 20),
	)
}
