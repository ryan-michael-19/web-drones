package impl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"reflect"

	"github.com/ryan-michael-19/web-drones/api"
	"github.com/ryan-michael-19/web-drones/utils"

	"math/rand"
	"time"

	"github.com/ryan-michael-19/web-drones/webdrones/public/model"
	. "github.com/ryan-michael-19/web-drones/webdrones/public/table"

	"github.com/go-jet/jet/v2/postgres"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type Server struct{}

func NewServer() Server {
	return Server{}
}

func InRange(val1 float64, val2 float64) bool {
	return math.Abs(val1-val2) < 1e-2
}

type BotsWithActions struct {
	model.BotMovementLedger
	model.Bots
}

var botVelocity = 0.5
var mineMax = 50.0
var mineMin = -50.0

func GenerateUserIDSubquery(username string) postgres.SelectStatement {
	return SELECT(Users.ID).FROM(Users).WHERE(Users.Username.EQ(String(username)))
}

// TODO: Use a key type here
const SESSION_VALUE = "cookie"
const USERNAME_VALUE = "username"

func NewRandomCoordinates(mineDistanceMin float64, mineDistanceMax float64) (float64, float64) {
	// TODO: Make sure mines don't respawn on top of each other
	return mineDistanceMin + rand.Float64()*(mineDistanceMax-mineDistanceMin),
		mineDistanceMin + rand.Float64()*(mineDistanceMax-mineDistanceMin)
}

func GetSingleBotFromDB(botId string, username string) (api.Bot, error) {
	stmt := SELECT(
		Bots.Identifier, Bots.Name, Bots.InventoryCount,
		BotMovementLedger.NewX, BotMovementLedger.NewY, BotMovementLedger.TimeActionStarted,
	).FROM(
		Bots.INNER_JOIN(
			BotMovementLedger, BotMovementLedger.BotID.EQ(Bots.ID),
		).INNER_JOIN(Users, Users.ID.EQ(Bots.UserID)),
	).WHERE(
		Users.Username.EQ(String(username)).AND(Bots.Identifier.EQ(String(botId))),
	).ORDER_BY(
		Bots.Identifier.ASC(),
		BotMovementLedger.TimeActionStarted.ASC(),
	)
	var ledger []BotsWithActions
	err := stmt.Query(utils.DB, &ledger)
	if err != nil {
		return api.Bot{}, err
	}
	res, err := GetBotsFromLedger(ledger, time.Now(), botVelocity)
	if err != nil {
		return api.Bot{}, err
	}
	return res[0], nil
}

func GetBotsFromDB(username string) ([]api.Bot, error) {
	stmt := SELECT(
		Bots.Identifier, Bots.Name, Bots.InventoryCount,
		BotMovementLedger.NewX, BotMovementLedger.NewY, BotMovementLedger.TimeActionStarted,
	).FROM(
		Bots.INNER_JOIN(
			BotMovementLedger, BotMovementLedger.BotID.EQ(Bots.ID),
		).INNER_JOIN(Users, Users.ID.EQ(Bots.UserID)),
	).WHERE(
		Users.Username.EQ(String(username)),
	).ORDER_BY(
		Bots.Identifier.ASC(),
		BotMovementLedger.TimeActionStarted.ASC(),
	)
	var ledger []BotsWithActions
	err := stmt.Query(utils.DB, &ledger)
	if err != nil {
		return []api.Bot{}, err
	}
	res, err := GetBotsFromLedger(ledger, time.Now(), botVelocity)
	if err != nil {
		return []api.Bot{}, err
	}
	return res, nil
}

func GetMinesFromDB(username string) ([]api.Coordinates, error) {
	stmt := SELECT(Mines.X, Mines.Y).FROM(
		Mines.INNER_JOIN(Users, Users.ID.EQ(Mines.UserID)),
	).WHERE(
		Users.Username.EQ(String(username)),
	)
	var dbResults []model.Mines
	err := stmt.Query(utils.DB, &dbResults)
	if err != nil {
		return []api.Coordinates{}, err
	}
	mines := make([]api.Coordinates, len(dbResults))
	for i, res := range dbResults {
		mines[i].X = res.X
		mines[i].Y = res.Y
	}
	return mines, nil
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

// (GET /bots)
func (Server) GetBots(ctx context.Context, request api.GetBotsRequestObject) (api.GetBotsResponseObject, error) {
	res, err := GetBotsFromDB(ctx.Value(USERNAME_VALUE).(string))
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return api.GetBots200JSONResponse(res), nil
}

// (GET /bots/{botId})
func (Server) GetBotsBotId(ctx context.Context, request api.GetBotsBotIdRequestObject) (api.GetBotsBotIdResponseObject, error) {
	res, err := GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return api.GetBotsBotId200JSONResponse(res), nil
}

// (POST /bots/{botId}/extract)
func (Server) PostBotsBotIdExtract(ctx context.Context, request api.PostBotsBotIdExtractRequestObject) (api.PostBotsBotIdExtractResponseObject, error) {
	bot, err := GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
	if err != nil {
		// TODO: Convert to 500
		slog.Error(err.Error())
		return nil, err
	}
	username := ctx.Value(USERNAME_VALUE).(string)
	var currentMine *api.Coordinates = nil
	mines, err := GetMinesFromDB(username)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	for _, mine := range mines {
		if InRange(bot.Coordinates.X, mine.X) && InRange(bot.Coordinates.Y, mine.Y) {
			currentMine = &mine
		}
	}
	if currentMine == nil {
		// TODO: Return error?
		return api.PostBotsBotIdExtract422TextResponse("Bot is not currently near a mine"), nil
	} else {
		// Add scrap metal to bot's inventory.
		// Then delete the mine and create a new one.
		tx, err := utils.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		// TODO: Convert to jet RawStatement (can't find support for x=x+1 in jet updates)
		_, err = utils.DB.Exec(
			// TODO: Use join instead of subquery
			"UPDATE bots SET inventory_count = inventory_count + 1, updated_at = NOW() "+
				" WHERE identifier = $1 AND user_id = (SELECT id FROM users WHERE username = $2)",
			bot.Identifier, username,
		)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		var x float64
		var y float64
		newMineSet := false
		for range 100 {
			x, y = NewRandomCoordinates(mineMin, mineMax)
			mineOverlaps := false
			for _, mine := range mines {
				if InRange(x, mine.X) || InRange(y, mine.Y) {
					mineOverlaps = true
					break
				}
			}
			if mineOverlaps {
				mineOverlaps = false
				continue
			} else {
				newMineSet = true
				break
			}
		}
		if !newMineSet {
			return nil, errors.New("could not get new mine coordinates")
		}
		// TODO: Use join instead of subquery
		now := time.Now()
		newMine := model.Mines{X: x, Y: y, UpdatedAt: &now}
		stmt := Mines.UPDATE(Mines.X, Mines.Y, Mines.UpdatedAt).MODEL(newMine).
			WHERE(
				Mines.X.BETWEEN(Float(currentMine.X-1e-2), Float(currentMine.X+1e-2)).
					AND(Mines.Y.BETWEEN(Float(currentMine.Y-1e-2), Float(currentMine.Y+1e-2))).
					AND(
						Mines.UserID.EQ(
							IntExp(
								SELECT(Users.ID).FROM(Users).WHERE(Users.Username.EQ(String(username))),
							),
						)))
		res, err := stmt.Exec(utils.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		rowCount, err := res.RowsAffected()
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		if rowCount != 1 {
			err = errors.New("more than one mine updated")
			slog.Error(err.Error())
			return nil, err
		}
		tx.Commit()
		updatedBot, err := GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		return api.PostBotsBotIdExtract200JSONResponse(updatedBot), nil
	}
}

// (POST /bots/{botId}/newBot)
func (Server) PostBotsBotIdNewBot(ctx context.Context, request api.PostBotsBotIdNewBotRequestObject) (api.PostBotsBotIdNewBotResponseObject, error) {
	bot, err := GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	if bot.Inventory >= 3 {
		username := ctx.Value(USERNAME_VALUE).(string)
		tx, err := utils.DB.Begin()
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		defer tx.Rollback()
		utils.DB.Exec(
			// TODO: Convert to jet and remove subquery
			"UPDATE bots SET inventory_count = inventory_count - 3, updated_at = NOW() "+
				"WHERE identifier = $1 AND user_id = (SELECT id FROM users WHERE username = $2)",
			bot.Identifier, username,
		)
		uuid := uuid.NewString()
		stmt := Bots.INSERT(Bots.Identifier, Bots.InventoryCount, Bots.Name, Bots.UserID).VALUES(
			uuid, 0, request.Body.NewBotName, GenerateUserIDSubquery(username),
		)
		_, err = stmt.Exec(utils.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		stmt = GenerateMoveActionQuery(
			uuid, ctx.Value(USERNAME_VALUE).(string), bot.Coordinates.X, bot.Coordinates.Y,
		)
		_, err = stmt.Exec(utils.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		// TODO: Get bot from database??
		return api.PostBotsBotIdNewBot200JSONResponse(
			api.Bot{
				Coordinates: bot.Coordinates,
				Identifier:  uuid,
				Inventory:   0,
				Name:        request.Body.NewBotName,
				Status:      api.IDLE,
			},
		), nil
	} else {
		return api.PostBotsBotIdNewBot422TextResponse("Bot doesn't have enough scrap metal."), nil
	}

}

// (POST /init)
func (Server) PostInit(ctx context.Context, request api.PostInitRequestObject) (api.PostInitResponseObject, error) {
	username := ctx.Value(USERNAME_VALUE).(string)
	tx, err := utils.DB.Begin()
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer tx.Rollback()
	// TODO: Convert this all into one query
	{
		stmt := BotMovementLedger.DELETE().USING(Users).WHERE(
			BotMovementLedger.UserID.EQ(Users.ID).AND(
				Users.Username.EQ(String(username)),
			),
		)
		_, err := stmt.Exec(utils.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
	}
	{
		stmt := Bots.DELETE().USING(Users).WHERE(
			Bots.UserID.EQ(Users.ID).AND(
				Users.Username.EQ(String(username)),
			),
		)
		_, err := stmt.Exec(utils.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
	}
	{
		stmt := Mines.DELETE().USING(Users).WHERE(
			Mines.UserID.EQ(Users.ID).AND(
				Users.Username.EQ(String(username)),
			),
		)
		_, err := stmt.Exec(utils.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
	}
	newBots := []struct {
		botName string
		coords  api.Coordinates
	}{
		{
			botName: "Bob",
			coords: api.Coordinates{
				X: 0, Y: 0,
			},
		},
		{
			botName: "Sam",
			coords: api.Coordinates{
				X: 5, Y: 5,
			},
		},
		{
			botName: "Gretchen",
			coords: api.Coordinates{
				X: -5, Y: -5,
			},
		},
	}
	for _, newBot := range newBots {
		uuid := uuid.NewString()
		stmt := Bots.INSERT(
			Bots.CreatedAt, Bots.UpdatedAt, Bots.Identifier, Bots.Name, Bots.InventoryCount, Bots.UserID,
		).VALUES(
			NOW(), NOW(), uuid, newBot.botName, 0, GenerateUserIDSubquery(username),
		)
		_, err = stmt.Exec(utils.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		moveStatement := GenerateMoveActionQuery(
			uuid, username, newBot.coords.X, newBot.coords.Y,
		)
		_, err = moveStatement.Exec(utils.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
	}
	stmt := Mines.INSERT(
		Mines.CreatedAt, Mines.UpdatedAt, Mines.UserID, Mines.X, Mines.Y,
	)
	mineCount := 10
	for range mineCount {
		x, y := NewRandomCoordinates(mineMin, mineMax)
		stmt = stmt.VALUES(
			NOW(), NOW(), SELECT(Users.ID).FROM(Users).WHERE(Users.Username.EQ(String(username))), x, y,
		)
	}
	_, err = stmt.Exec(utils.DB)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	bots, err := GetBotsFromDB(username)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	mines, err := GetMinesFromDB(username)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	resp := api.PostInit200JSONResponse{
		Bots:  bots,
		Mines: mines,
	}
	slog.Info("Game has been reset", "username", username)
	return api.PostInit200JSONResponse(resp), nil
}

// (POST /login)
func (Server) PostLogin(ctx context.Context, request api.PostLoginRequestObject) (api.PostLoginResponseObject, error) {
	slog.Info("New Login", "username", ctx.Value("username").(string))
	return api.PostLogin200TextResponse("Login Successful"), nil
}

// (POST /bots/{botId}/move)
func (Server) PostBotsBotIdMove(ctx context.Context, request api.PostBotsBotIdMoveRequestObject) (api.PostBotsBotIdMoveResponseObject, error) {
	username := ctx.Value(USERNAME_VALUE).(string)
	stmt := GenerateMoveActionQuery(
		request.BotId, ctx.Value(USERNAME_VALUE).(string), request.Body.X, request.Body.Y,
	)
	status, err := stmt.Exec(utils.DB)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	rowCount, err := status.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	if rowCount != 1 {
		errString := fmt.Sprintf("Expected 1 row to be affected but %d were affected", rowCount)
		slog.Error(errString)
		return nil, errors.New(errString)
	}
	resp, err := GetSingleBotFromDB(request.BotId, username)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return api.PostBotsBotIdMove200JSONResponse(resp), nil
}

// (GET /mines)
func (Server) GetMines(ctx context.Context, request api.GetMinesRequestObject) (api.GetMinesResponseObject, error) {
	username := ctx.Value(USERNAME_VALUE).(string)
	mines, err := GetMinesFromDB(username)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return api.GetMines200JSONResponse(mines), nil
}

// (POST /newUser)
func (Server) PostNewUser(ctx context.Context, request api.PostNewUserRequestObject) (api.PostNewUserResponseObject, error) {
	slog.Info("New user created", "username", ctx.Value("username").(string))
	return api.PostNewUser200TextResponse("New User Created"), nil
}
