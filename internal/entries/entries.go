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

func (te TimeEntry) String() string {
	return fmt.Sprintf(
		"[%s - %s] [%s] %s %v",
		te.From.Format(time.DateTime),
		te.Till.Format(time.DateTime),
		utils.FitString(te.Issue, 12),
		utils.FitString(te.Description, 30),
		te.Tags,
	)
}
