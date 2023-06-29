package commands

import (
	"errors"
	"main/src/repository/complaint"
	"sync"
)

type Cache struct {
	replies []complaint.Reply
	valid   bool
	sync.Mutex
}

func (c *Cache) Get(index int) (complaint.Reply, error) {
	c.Lock()
	defer c.Unlock()

	if !c.valid {
		return complaint.Reply{}, errors.New("cache invalid")
	}

	return c.replies[index], nil
}

func (c *Cache) Len() int {
	c.Lock()
	defer c.Unlock()

	return len(c.replies)
}

func (c *Cache) Valid() bool {
	c.Lock()
	defer c.Unlock()

	return c.valid
}

func (c *Cache) Refresh(replies []complaint.Reply) {
	c.Lock()
	defer c.Unlock()

	c.replies = replies
	c.valid = true
}

func (c *Cache) Validate() {
	c.Lock()
	defer c.Unlock()

	c.valid = true
}
