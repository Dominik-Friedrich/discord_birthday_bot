package commands

import (
	"errors"
	"main/src/repository"
	"sync"
)

type Cache struct {
	replies []repository.Reply
	valid   bool
	sync.Mutex
}

// Get returns the element at index.
// If the cache is invalid it still returns the element but also an error
func (c *Cache) Get(index int) (repository.Reply, error) {
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

func (c *Cache) Refresh(replies []repository.Reply) {
	c.replies = replies
	c.valid = true
}

func (c *Cache) Validate() {
	c.valid = true
}
