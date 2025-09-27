package cli

import (
	"github.com/tornermarton/timesheets/internal/config"
)

type Context struct {
	Version string
	Config  *config.Config
}
