package constants

import (
	"time"
)

var NOW = time.Now()
var TODAY = time.Date(NOW.Year(), NOW.Month(), NOW.Day(), 0, 0, 0, 0, NOW.Location())
var TOMORROW = TODAY.Add(24 * time.Hour)
