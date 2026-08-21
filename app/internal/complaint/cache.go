package complaint

import (
	"context"
	"errors"
	"sync"
)

// cache holds the in-memory set of complaint replies so /complain doesn't
// need to hit the database on every invocation.
type cache struct {
	replies []Reply
	valid   bool
	sync.Mutex
}

// Get returns the element at index.
// If the cache is invalid it still returns the element but also an error
func (c *cache) Get(index int) (Reply, error) {
	var err error
	if !c.valid {
		err = errors.New("cache invalid")
	}

	return c.replies[index], err
}

func (c *cache) Len() int {
	return len(c.replies)
}

func (c *cache) Valid() bool {
	return c.valid
}

func (c *cache) Refresh(replies []Reply) {
	c.replies = replies
	c.valid = true
}

// refreshCache reloads replies from the database into c. Callers must hold
// c's lock.
func refreshCache(ctx context.Context, repo *Repository, c *cache) error {
	replies, err := repo.GetComplaintReplies(ctx)
	if err != nil {
		return err
	}

	c.Refresh(replies)
	return nil
}
