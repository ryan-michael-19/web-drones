package schemas

import (
	"colony-bots/api"
	"time"

	"gorm.io/gorm"
)

type Bots struct {
	gorm.Model
	ID         int
	Identifier string
	Name       string
	Status     api.BotStatus
	X          float64
	Y          float64
}

type BotActions struct {
	gorm.Model
	ID                int
	BotID             int
	Bot               Bots `gorm:"foreignKey:BotID;references:ID"`
	TimeActionStarted time.Time
	New_X             float64
	New_Y             float64
}

type Mines struct {
	gorm.Model
	ID int
	X  float64
	Y  float64
}
