package cli

import (
	"flag"
	"fmt"
	"time"
)

type FlagSet struct {
	*flag.FlagSet
}

// -- time.Time Value
type timeValue time.Time

func newTimeValue(val time.Time, p *time.Time) *timeValue {
	*p = val
	return (*timeValue)(p)
}

func (t *timeValue) Set(s string) error {
	layouts := []string{
		time.RFC3339[:10], // covers 2025-01-01
		time.RFC3339,      // covers 2025-01-01T16:00:00
	}

	for _, layout := range layouts {
		if v, err := time.Parse(layout, s); err == nil {
			*t = timeValue(v)
			return nil
		}
	}

	return fmt.Errorf("cannot parse \"%s\" with layouts %v", s, layouts)
}

func (t *timeValue) Get() any { return time.Time(*t) }

func (t *timeValue) String() string { return (*time.Time)(t).String() }

func (f *FlagSet) TimeVar(p *time.Time, name string, value time.Time, usage string) {
	f.Var(newTimeValue(value, p), name, usage)
}

func TimeVar(p *time.Time, name string, value time.Time, usage string) {
	flag.CommandLine.Var(newTimeValue(value, p), name, usage)
}

func (f *FlagSet) Time(name string, value time.Time, usage string) *time.Time {
	p := new(time.Time)
	f.TimeVar(p, name, value, usage)
	return p
}
