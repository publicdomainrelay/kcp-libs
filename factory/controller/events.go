package controller

import "time"

func (c *Controller) RecordEvent() {
	c.lastEventNanos.Store(time.Now().UnixNano())
}

func (c *Controller) CacheAge() time.Duration {
	last := c.lastEventNanos.Load()
	if last == 0 {
		return 0
	}
	return time.Since(time.Unix(0, last))
}
