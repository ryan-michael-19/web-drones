package schemas

import (
	"colony-bots/api"

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
	ID     int
	Bot    Bots
	Action api.BotStatus
	New_X  float32
	New_Y  float32
}

type Mines struct {
	gorm.Model
	ID int
	X  float64
	Y  float64
}
