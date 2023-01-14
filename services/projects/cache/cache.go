package cache

type Cache struct {
	Statuses map[string]map[string]string	
}

func NewCache() Cache {
	return Cache{
		Statuses: make(map[string]map[string]string),
	}
}

func (c *Cache) Get(project string) (map[string]string, bool) {
	status, ok := c.Statuses[project]

	return status, ok
}

func (c *Cache) Set(project string, statuses map[string]string) {
	c.Statuses[project] = statuses
}

func (c *Cache) AddStatus(project string, serviceName string, status string) {
	c.Statuses[project][serviceName] = status	
}

func (c *Cache) RemoveStatus(project string, serviceName string) {
	delete(c.Statuses[project], serviceName)
}

func (c *Cache) Clear() {
	for project := range c.Statuses {
		delete(c.Statuses, project)
	}
}
