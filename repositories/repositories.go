package repositories

import (
	"platform-backend/casino"
)

type Repos struct {
	Casino   casino.Repository
}

func NewRepositories(Casino casino.Repository) *Repos {
	return &Repos{
		Casino: Casino,
	}
}
