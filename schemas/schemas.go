package schemas

import (
	"colony-bots/api"
	"time"
)

type Metadata struct {
	CreatedAt time.Time `db:"-"`
	UpdatedAt time.Time `db:"-"`
	DeletedAt time.Time `db:"-"`
}

type Bots struct {
	Metadata
	ID         int `db:"-"`
	Identifier string
	Name       string
	Status     api.BotStatus
	X          float64
	Y          float64
}

type BotActions struct {
	Metadata
	ID                int `db:"-"`
	BotID             int
	Bot               Bots `gorm:"foreignKey:BotID;references:ID"`
	TimeActionStarted time.Time
	NewX              float64
	NewY              float64
}

type Mines struct {
	Metadata
	ID int `db:"-"`
	X  float64
	Y  float64
}
