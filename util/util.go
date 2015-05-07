package util

import (
	"appengine"
	"time"
)

func LogTime(c appengine.Context, start time.Time, text string) {
	c.Infof("%s%s\n", text, time.Since(start))
}
