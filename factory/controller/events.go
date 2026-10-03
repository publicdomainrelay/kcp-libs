package controller

import "time"

func (c *Controller) recordEvent() {
	c.lastEventNanos.Store(c.opts.Now().UnixNano())
}

func (c *Controller) CacheAge() time.Duration {
	last := c.lastEventNanos.Load()
	if last == 0 {
		return 0
	}
	return c.opts.Now().Sub(time.Unix(0, last))
}
