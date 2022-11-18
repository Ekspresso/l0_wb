package cache

import (
	"l0/internal/models"
	"sync"
)

type Cache struct {
	m  map[int]models.Order
	mu sync.RWMutex
}

func New(initSize int) *Cache {
	return &Cache{
		m: make(map[int]models.Order, initSize),
	}
}

func (c *Cache) Set(u models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[u.Id] = u
}

func (c *Cache) Get(id int) (_ models.Order, exist bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	u, ok := c.m[id]
	return u, ok
}
