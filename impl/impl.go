package impl

import (
	"colony-bots/api"
	"colony-bots/schemas"
	"math"

	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct{}

func NewServer() Server {
	return Server{}
}

func openDB() *gorm.DB {
	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=EST"
	db, db_err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if db_err != nil {
		fmt.Println("COULD NOT OPEN DB CONNECTION", db_err)
		os.Exit(1)
	}
	return db
}

var db = openDB()

type GetBotLocationError struct {
	message string
}

func (e *GetBotLocationError) Error() string {
	return e.message
}

func GetBotLocation(
	initialCoordinates api.Coordinates, destinationCoordinates api.Coordinates,
	movementStartTime time.Time, currentTime time.Time, botVelocity float64,
) (api.Coordinates, error) {
	if currentTime.Before(movementStartTime) {
		return api.Coordinates{X: 0, Y: 0}, &GetBotLocationError{message: "Current time cannot be before movement start time"}
	}
	// handle when bot is at destination location
	// Determine how long it would take to get destination
	movementVector := api.Coordinates{
		X: destinationCoordinates.X - initialCoordinates.X, Y: destinationCoordinates.Y - initialCoordinates.Y,
	}
	// TODO: Remove sqrt
	movementVectorLen := math.Sqrt(math.Pow(movementVector.X, 2) + math.Pow(movementVector.Y, 2))
	timeToReachDestination := movementVectorLen / botVelocity
	timeDelta := currentTime.Sub(movementStartTime).Seconds()
	if timeDelta > timeToReachDestination {
		return destinationCoordinates, nil
	}
	// Get the direction of the vector the bot is heading towards
	currentMovementMagnitude := timeDelta * botVelocity
	currentMovementDirection := math.Atan2(movementVector.Y, movementVector.X)

	currentLocation := api.Coordinates{
		X: (currentMovementMagnitude * math.Cos(currentMovementDirection)) + initialCoordinates.X,
		Y: (currentMovementMagnitude * math.Sin(currentMovementDirection)) + initialCoordinates.Y,
	}
	return currentLocation, nil
}

// (GET /bots)
func (Server) GetBots(w http.ResponseWriter, r *http.Request) {
	id := "Totally random id"
	name := "Beep Boop"
	status := api.IDLE
	coords := api.Coordinates{X: 6, Y: 45}
	resp := []api.Bot{
		{
			Coordinates: coords,
			Identifier:  id,
			Name:        name,
			Status:      status,
		},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// (GET /bots/{botId})
func (Server) GetBotsBotId(w http.ResponseWriter, r *http.Request, botId string) {
	resp := api.Bot{
		Coordinates: api.Coordinates{X: 55, Y: 78},
		Identifier:  "Another random Id",
		Name:        "Boop Beep",
		Status:      api.IDLE,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// (POST /init)
func (Server) PostInit(w http.ResponseWriter, r *http.Request) {
	db.Where("1=1").Delete(&schemas.Bots{})
	db.Where("1=1").Delete(&schemas.Mines{})

	db.Create(&schemas.Bots{
		Identifier: "definitely-a-uuid",
		Name:       "Big Chungus",
		Status:     api.IDLE,
		X:          5,
		Y:          30,
	})
	mineCount := 10
	mines := make([]schemas.Mines, mineCount)
	for i := range mineCount {
		mines[i] = schemas.Mines{
			X: rand.Float64(),
			Y: rand.Float64(),
		}
	}
	db.Create(mines)
	// TODO: Update openapi to include bots and mines in response
	var botFromDB schemas.Bots
	db.First(&botFromDB)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(botFromDB)
}

// (POST /bots/{botId}/move)
func (Server) PostBotsBotIdMove(w http.ResponseWriter, r *http.Request, botId string) {
	var bot schemas.Bots
	db.First(&bot, "Identifier = ?", botId)
	var newCoords api.Coordinates
	err := json.NewDecoder(r.Body).Decode(&newCoords)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp := api.Bot{
		Identifier:  bot.Identifier,
		Name:        bot.Name,
		Status:      api.MOVING,
		Coordinates: newCoords,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// (GET /mines)
func (Server) GetMines(w http.ResponseWriter, r *http.Request) {
	resp := []api.Coordinates{
		{X: 1, Y: 1},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
