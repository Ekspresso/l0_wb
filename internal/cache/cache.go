package cache

import (
	"l0/internal/models"
	"sync"
)

//Данные храняться в отображении(карте). RWMutex нужен для того,
//чтобы не возникало ошибок в ситуации, когда кто-то пытается считать данные из буфера параллельно с записью.
type Cache struct {
	m  map[int]models.Order
	mu sync.RWMutex
}

func New(initSize int) *Cache {
	return &Cache{
		m: make(map[int]models.Order, initSize),
	}
}

//Метод записи в кэш
func (c *Cache) Set(u models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[u.Id] = u
}

//Метод выборки данных из кэша
func (c *Cache) Get(id int) (_ models.Order, exist bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	u, ok := c.m[id]
	return u, ok
}
