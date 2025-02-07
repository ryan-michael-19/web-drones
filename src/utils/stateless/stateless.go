package stateless

import (
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"time"

	"github.com/ryan-michael-19/web-drones/api"
	"github.com/ryan-michael-19/web-drones/webdrones/public/model"
	. "github.com/ryan-michael-19/web-drones/webdrones/public/table"

	"github.com/go-jet/jet/v2/postgres"
	. "github.com/go-jet/jet/v2/postgres"
)

type BotsWithActions struct {
	model.BotMovementLedger
	model.Bots
}

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
		return api.Coordinates{X: 0, Y: 0}, &GetBotLocationError{
			message: "Current time cannot be before movement start time"}
	}
	movementVector := api.Coordinates{
		X: destinationCoordinates.X - initialCoordinates.X, Y: destinationCoordinates.Y - initialCoordinates.Y,
	}
	// TOOD: Remove sqrt
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

func GetBotsFromLedger(ledger []BotsWithActions, currentDatetime time.Time, botVelocity float64) ([]api.Bot, error) {
	var bots []api.Bot
	var currentBotCoords api.Coordinates
	currentBotCoords = api.Coordinates{X: ledger[0].NewX, Y: ledger[0].NewY}
	for i := range ledger {
		if i < len(ledger)-1 && ledger[i].Identifier == ledger[i+1].Identifier {
			// continue calculating velocity
			var err error
			currentBotCoords, err = GetBotLocation(
				currentBotCoords,
				api.Coordinates{X: ledger[i].NewX, Y: ledger[i].NewY},
				ledger[i].TimeActionStarted,
				ledger[i+1].TimeActionStarted,
				botVelocity,
			)
			if err != nil {
				return []api.Bot{}, err
			}
		} else {
			// We need the final position of the bot based on the last action
			// it has recieved
			var err error
			currentBotCoords, err = GetBotLocation(
				currentBotCoords,
				api.Coordinates{X: ledger[i].NewX, Y: ledger[i].NewY},
				ledger[i].TimeActionStarted,
				currentDatetime,
				botVelocity,
			)
			if err != nil {
				return []api.Bot{}, err
			}
			var botStatus api.BotStatus
			// Set bot to idle if it is at the coordinates of its last move action
			if reflect.DeepEqual(currentBotCoords, api.Coordinates{X: ledger[i].NewX, Y: ledger[i].NewY}) {
				botStatus = api.IDLE
			} else {
				botStatus = api.MOVING
			}

			bot := api.Bot{
				Coordinates: currentBotCoords,
				Identifier:  ledger[i].Identifier,
				Name:        ledger[i].Name,
				Status:      botStatus,
				Inventory:   int(ledger[i].InventoryCount),
			}
			bots = append(bots, bot)
			// Initialize coordinates for the next bot (if it exists)
			if i < len(ledger)-1 {
				currentBotCoords = api.Coordinates{X: ledger[i+1].NewX, Y: ledger[i+1].NewY}
			}
		}
	}
	return bots, nil
}

func GenerateUserIDSubquery(username string) postgres.SelectStatement {
	return SELECT(Users.ID).FROM(Users).WHERE(Users.Username.EQ(String(username)))
}

func GenerateMoveActionQuery(identifier string, username string, x float64, y float64) postgres.InsertStatement {
	return BotMovementLedger.INSERT(
		BotMovementLedger.CreatedAt, BotMovementLedger.UpdatedAt,
		BotMovementLedger.BotID, BotMovementLedger.UserID,
		BotMovementLedger.TimeActionStarted, BotMovementLedger.NewX, BotMovementLedger.NewY,
	).VALUES(
		NOW(), NOW(),
		SELECT(Bots.ID).FROM(Bots).WHERE(Bots.Identifier.EQ(String(identifier))),
		GenerateUserIDSubquery(username),
		NOW(), x, y,
	)
}

func NewRandomCoordinates(mineDistanceMin float64, mineDistanceMax float64) (float64, float64) {
	// TODO: Make sure mines don't respawn on top of each other
	return mineDistanceMin + rand.Float64()*(mineDistanceMax-mineDistanceMin),
		mineDistanceMin + rand.Float64()*(mineDistanceMax-mineDistanceMin)
}

// Set up an error struct that will log what's going on with the server
// without leaking server errors to the client
// TODO: Use error wrapping???
type AuthError struct {
	OriginalError error
	NewError      error
}

func (e *AuthError) BothErrors() string {
	var newErrorMessage string
	if e.NewError != nil {
		newErrorMessage = e.NewError.Error()
	} else {
		newErrorMessage = ""
	}
	var originalErrorMessage string
	if e.OriginalError != nil {
		originalErrorMessage = e.OriginalError.Error()
	} else {
		originalErrorMessage = ""
	}
	// TODO: Convert this to slog for observability
	return fmt.Sprintf("Authentication error: Original: %s New: %s", originalErrorMessage, newErrorMessage)
}

func (e *AuthError) Error() string {
	if e.NewError == nil {
		return "Unspecified authentication error"
	} else {
		return fmt.Sprintf("Authentication error: %s", e.NewError.Error())
	}
}
