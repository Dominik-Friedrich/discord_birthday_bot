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

// Get returns the element at index.
// If the cache is invalid it still returns the element but also an error
func (c *Cache) Get(index int) (complaint.Reply, error) {
	var err error
	if !c.valid {
		err = errors.New("cache invalid")
	}

	return c.replies[index], err
}

func (c *Cache) Len() int {
	return len(c.replies)
}

func (c *Cache) Valid() bool {
	return c.valid
}

func (c *Cache) Refresh(replies []complaint.Reply) {
	c.replies = replies
	c.valid = true
}

func (c *Cache) Validate() {
	c.valid = true
}
