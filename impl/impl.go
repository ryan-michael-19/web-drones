package impl

import (
	"colony-bots/api"
	"colony-bots/schemas"
	"context"
	"fmt"
	"log"
	"math"
	"reflect"

	"colony-bots/webdrones/public/model"
	. "colony-bots/webdrones/public/table"
	"math/rand"
	"time"

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

// type BotsWithActions struct {
// schemas.Bots
// schemas.BotActions
// }

type BotsWithActions struct {
	model.BotMovementLedger
	model.Bots
}

func BuildLogger() *log.Logger {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return log.Default()
}

var global_ctx = context.Background()
var db = schemas.OpenDB()
var logger = BuildLogger()
var botVelocity = 0.5
var mineMax = 50.0
var mineMin = -50.0

func GenerateUserIDSubquery(username string) postgres.SelectStatement {
	return SELECT(Users.ID).FROM(Users).WHERE(Users.Username.EQ(String(username)))
}

const SESSION_VALUE = "cookie"
const USERNAME_VALUE = "username"

func NewRandomCoordinates(mineDistanceMin float64, mineDistanceMax float64) (float64, float64) {
	// TODO: Make sure mines don't respawn on top of each other
	return mineDistanceMin + rand.Float64()*(mineDistanceMax-mineDistanceMin),
		mineDistanceMin + rand.Float64()*(mineDistanceMax-mineDistanceMin)
}

func GetSingleBotFromDB(botId string, username string) api.Bot {
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
	stmt.Query(db, &ledger)
	return GetBotsFromLedger(ledger, time.Now(), botVelocity)[0]
}

func GetBotsFromDB(username string) []api.Bot {
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
	err := stmt.Query(db, &ledger)
	fmt.Println(stmt.DebugSql())
	if err != nil {
		logger.Fatal(err)
	}
	return GetBotsFromLedger(ledger, time.Now(), botVelocity)
}

func GetMinesFromDB(username string) []api.Coordinates {
	// rows, err := db.Query(global_ctx, "SELECT mines.x, mines.y FROM mines")
	stmt := SELECT(Mines.X, Mines.Y).FROM(
		Mines.INNER_JOIN(Users, Users.ID.EQ(Mines.UserID)),
	).WHERE(
		Users.Username.EQ(String(username)),
	)
	fmt.Println(stmt.DebugSql())
	var dbResults []model.Mines
	err := stmt.Query(db, &dbResults)
	if err != nil {
		log.Fatal(err)
	}
	mines := make([]api.Coordinates, len(dbResults))
	for i, res := range dbResults {
		mines[i].X = res.X
		mines[i].Y = res.Y
	}
	return mines
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

func GetBotsFromLedger(ledger []BotsWithActions, currentDatetime time.Time, botVelocity float64) []api.Bot {
	var bots []api.Bot
	var currentBotCoords api.Coordinates
	// TODO: Write unit test to handle this case
	if len(ledger) == 1 {
		bots = append(bots, api.Bot{
			Coordinates: api.Coordinates{
				X: ledger[0].NewX, Y: ledger[0].NewY,
			},
			Identifier: ledger[0].Identifier,
			Name:       ledger[0].Name,
			Status:     api.IDLE,
			Inventory:  int(ledger[0].InventoryCount),
		})
		return bots
	}
	for i := range ledger {
		if i == 0 {
			currentBotCoords = api.Coordinates{X: ledger[i].NewX, Y: ledger[i].NewY}
			// check if the next record exists and refers to the same bot
		} else if i < len(ledger)-1 && ledger[i].Identifier == ledger[i+1].Identifier {
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
				logger.Fatal(err)
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
				logger.Fatal(err)
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
		}
	}
	return bots
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
	return api.GetBots200JSONResponse(GetBotsFromDB(ctx.Value(USERNAME_VALUE).(string))), nil
}

// (GET /bots/{botId})
func (Server) GetBotsBotId(ctx context.Context, request api.GetBotsBotIdRequestObject) (api.GetBotsBotIdResponseObject, error) {
	return api.GetBotsBotId200JSONResponse(GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))), nil
}

// (POST /bots/{botId}/mine)
func (Server) PostBotsBotIdMine(ctx context.Context, request api.PostBotsBotIdMineRequestObject) (api.PostBotsBotIdMineResponseObject, error) {
	bot := GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
	username := ctx.Value(USERNAME_VALUE).(string)
	var currentMine *api.Coordinates = nil
	for _, mine := range GetMinesFromDB(username) {
		if InRange(bot.Coordinates.X, mine.X) && InRange(bot.Coordinates.Y, mine.Y) {
			currentMine = &mine
		}
	}
	if currentMine == nil {
		return api.PostBotsBotIdMine422TextResponse("Bot is not currently near a mine"), nil
	} else {
		// Add scrap metal to bot's inventory.
		// Then delete the mine and create a new one.
		tx, err := db.Begin()
		defer tx.Rollback()
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Convert to jet RawStatement (can't find support for x=x+1 in jet updates)
		_, err = db.Exec(
			// TODO: Use join instead of subquery
			"UPDATE bots SET inventory_count = inventory_count + 1, updated_at = NOW() WHERE identifier = $1 AND user_id = (SELECT id FROM users WHERE username = $2)",
			bot.Identifier, username,
		)
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Throw error if more than one mine is removed
		// TODO: Convert to jet and remove subquery
		_, err = db.Exec(
			"DELETE FROM mines WHERE abs(x-$1) < $2 AND abs(y-$3) < $4 AND user_id = (SELECT id FROM users WHERE username = $5)",
			currentMine.X, 1e-9, currentMine.Y, 1e-9, username,
		)
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Handle case where new coordinates spawn on top of new mine
		x, y := NewRandomCoordinates(mineMin, mineMax)
		stmt := Mines.INSERT(Mines.X, Mines.Y, Mines.UserID).
			VALUES(x, y, GenerateUserIDSubquery(username))
		_, err = stmt.Exec(db)
		if err != nil {
			log.Fatal(err)
		}
		tx.Commit()
		return api.PostBotsBotIdMine200JSONResponse(GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))), nil
	}
}

// (POST /bots/{botId}/newBot)
func (Server) PostBotsBotIdNewBot(ctx context.Context, request api.PostBotsBotIdNewBotRequestObject) (api.PostBotsBotIdNewBotResponseObject, error) {
	bot := GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
	if bot.Inventory >= 3 {
		username := ctx.Value(USERNAME_VALUE).(string)
		tx, err := db.Begin()
		defer tx.Rollback()
		db.Exec(
			// TODO: Convert to jet and remove subquery
			"UPDATE bots SET inventory_count = inventory_count - 3, updated_at = NOW() WHERE identifier = $1 AND user_id = (SELECT id FROM users WHERE username = $2)",
			bot.Identifier, username,
		)
		uuid := uuid.NewString()
		stmt := Bots.INSERT(Bots.Identifier, Bots.InventoryCount, Bots.Name, Bots.UserID).VALUES(
			uuid, 0, request.Body.NewBotName, GenerateUserIDSubquery(username),
		)
		_, err = stmt.Exec(db)
		if err != nil {
			log.Fatal(err)
		}
		stmt = GenerateMoveActionQuery(
			uuid, ctx.Value(USERNAME_VALUE).(string), bot.Coordinates.X, bot.Coordinates.Y,
		)
		_, err = stmt.Exec(db)
		if err != nil {
			log.Fatal(err)
		}
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
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
	uname := ctx.Value(USERNAME_VALUE).(string)
	// TODO: Convert batch to transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	{
		stmt := BotMovementLedger.DELETE().USING(Users).WHERE(
			BotMovementLedger.UserID.EQ(Users.ID).AND(
				Users.Username.EQ(String(uname)),
			),
		)
		_, err := stmt.Exec(db)
		if err != nil {
			log.Fatal(err)
		}
	}
	{
		stmt := Bots.DELETE().USING(Users).WHERE(
			Bots.UserID.EQ(Users.ID).AND(
				Users.Username.EQ(String(uname)),
			),
		)
		_, err := stmt.Exec(db)
		if err != nil {
			log.Fatal(err)
		}
	}
	{
		stmt := Mines.DELETE().USING(Users).WHERE(
			Mines.UserID.EQ(Users.ID).AND(
				Users.Username.EQ(String(uname)),
			),
		)
		_, err := stmt.Exec(db)
		if err != nil {
			log.Fatal(err)
		}
	}
	uuid := uuid.NewString()
	stmt := Bots.INSERT(
		Bots.CreatedAt, Bots.UpdatedAt, Bots.Identifier, Bots.Name, Bots.InventoryCount, Bots.UserID,
	).VALUES(
		NOW(), NOW(), uuid, "Bob", 0, GenerateUserIDSubquery(uname),
	)
	_, err = stmt.Exec(db)
	if err != nil {
		log.Fatal(err)
	}
	moveStatement := GenerateMoveActionQuery(
		uuid, uname, 0, 0,
	)
	_, err = moveStatement.Exec(db)
	if err != nil {
		log.Fatal(err)
	}
	stmt = Mines.INSERT(
		Mines.CreatedAt, Mines.UpdatedAt, Mines.UserID, Mines.X, Mines.Y,
	)
	mineCount := 10
	for range mineCount {
		x, y := NewRandomCoordinates(mineMin, mineMax)
		stmt = stmt.VALUES(
			NOW(), NOW(), SELECT(Users.ID).FROM(Users).WHERE(Users.Username.EQ(String(uname))), x, y,
		)
	}
	_, err = stmt.Exec(db)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	bots := GetBotsFromDB(uname)
	mines := GetMinesFromDB(uname)
	resp := api.PostInit200JSONResponse{
		Bots:  &bots,
		Mines: &mines,
	}
	return api.PostInit200JSONResponse(resp), nil
}

// (POST /login)
func (Server) PostLogin(ctx context.Context, request api.PostLoginRequestObject) (api.PostLoginResponseObject, error) {
	return api.PostLogin200TextResponse("Login Successful"), nil
}

// (POST /bots/{botId}/move)
func (Server) PostBotsBotIdMove(ctx context.Context, request api.PostBotsBotIdMoveRequestObject) (api.PostBotsBotIdMoveResponseObject, error) {
	username := ctx.Value(USERNAME_VALUE).(string)
	stmt := GenerateMoveActionQuery(
		request.BotId, ctx.Value(USERNAME_VALUE).(string), request.Body.X, request.Body.Y,
	)
	status, err := stmt.Exec(db)
	if err != nil {
		log.Fatal(err)
	}
	rowCount, err := status.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCount != 1 {
		log.Fatalf(
			"Expected 1 row to be affected but %d were affected",
			rowCount)
	}
	resp := GetSingleBotFromDB(request.BotId, username)
	return api.PostBotsBotIdMove200JSONResponse(resp), nil
}

// (GET /mines)
func (Server) GetMines(ctx context.Context, request api.GetMinesRequestObject) (api.GetMinesResponseObject, error) {
	username := ctx.Value(USERNAME_VALUE).(string)
	mines := GetMinesFromDB(username)
	return api.GetMines200JSONResponse(mines), nil
}

// (POST /newUser)
func (Server) PostNewUser(ctx context.Context, request api.PostNewUserRequestObject) (api.PostNewUserResponseObject, error) {
	return api.PostNewUser200TextResponse("New User Created"), nil
}
