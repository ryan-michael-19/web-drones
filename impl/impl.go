package impl

import (
	"colony-bots/api"
	"colony-bots/schemas"
	"context"
	"log"
	"math"
	"reflect"
	"strings"

	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Server struct{}

func NewServer() Server {
	return Server{}
}

func BuildInsert(tableName string, colNames ...string) string {
	// TODO: This feels very injectable...
	valueFormats := []string{} // ["$1", "$2", etc..]
	for i := range colNames {
		valueFormats = append(valueFormats, "$"+fmt.Sprintf("%d", i+1))
	}
	valString := "NOW(),NULL,NULL," + strings.Join(valueFormats, ",")
	colString := "created_at,updated_at,deleted_at," + strings.Join(colNames, ",")
	sqlString := "INSERT INTO " + tableName + "(" + colString + ") VALUES (" + valString + ")"
	return sqlString
}

func InRange(val1 float64, val2 float64) bool {
	return math.Abs(val1-val2) < 1e-2
}

type BotsWithActions struct {
	schemas.Bots
	schemas.BotActions
}

func BuildLogger() *log.Logger {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return log.Default()
}

var db = schemas.OpenDB()
var global_ctx = context.Background()
var logger = BuildLogger()
var botVelocity = 0.5

func NewRandomCoordinates(mineDistanceMin float64, mineDistanceMax float64) (float64, float64) {
	// TODO: Make sure mines don't respawn on top of each other
	return mineDistanceMin + rand.Float64()*(mineDistanceMax-mineDistanceMin),
		mineDistanceMin + rand.Float64()*(mineDistanceMax-mineDistanceMin)
}

func GetMinesFromDB() []api.Coordinates {
	rows, err := db.Query(global_ctx, "SELECT mines.x, mines.y FROM mines")
	if err != nil {
		log.Fatal(err)
	}
	mines, err := pgx.CollectRows(rows, pgx.RowToStructByName[api.Coordinates])
	if err != nil {
		log.Fatal(err)
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

func GetBotsFromLedger(ledger []BotsWithActions, currentDatetime time.Time) []api.Bot {
	var bots []api.Bot
	currentBotCoords := api.Coordinates{X: ledger[0].New_X, Y: ledger[0].New_Y}
	for i := range ledger {
		// check if the next record exists and refers to the same bot
		if i < len(ledger)-1 && ledger[i].Identifier == ledger[i+1].Identifier {
			// continue calculating velocity
			var err error
			currentBotCoords, err = GetBotLocation(
				currentBotCoords,
				api.Coordinates{X: ledger[i+1].New_X, Y: ledger[i+1].New_Y},
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
				api.Coordinates{X: ledger[i].New_X, Y: ledger[i].New_Y},
				ledger[i].TimeActionStarted,
				currentDatetime,
				botVelocity,
			)
			if err != nil {
				logger.Fatal(err)
			}
			var botStatus api.BotStatus
			// Set bot to idle if it is at the coordinates of its last move action
			if reflect.DeepEqual(currentBotCoords, api.Coordinates{X: ledger[i].New_X, Y: ledger[i].New_Y}) {
				botStatus = api.IDLE
			} else {
				botStatus = api.MOVING
			}

			bot := api.Bot{
				Coordinates: currentBotCoords,
				Identifier:  ledger[i].Identifier,
				Name:        ledger[i].Name,
				Status:      botStatus,
			}
			bots = append(bots, bot)
		}
	}
	return bots
}

var botsWithActionsQuery = "" + // empty string gets around linter weirdness
	"SELECT bots.Identifier, bots.Name," +
	" bot_movement_ledger.new_x, bot_movement_ledger.new_y, bot_movement_ledger.time_action_started" +
	" FROM bots" +
	" LEFT JOIN bot_movement_ledger ON bots.ID = bot_movement_ledger.bot_id" +
	" ORDER BY bot_movement_ledger.Time_Action_Started ASC"

var botsWithActionsAndFilterQuery = "" +
	"SELECT bots.Identifier, bots.Name," +
	" bot_movement_ledger.new_x, bot_movement_ledger.new_y, bot_movement_ledger.time_action_started" +
	" FROM bots" +
	" LEFT JOIN bot_movement_ledger ON bots.ID = bot_movement_ledger.bot_id" +
	" WHERE bots.Identifier = $1" +
	" ORDER BY bot_movement_ledger.Time_Action_Started ASC"

var insertMoveActionQuery = "" +
	"INSERT INTO bot_movement_ledger (created_at,updated_at,deleted_at,bot_id,time_action_started,new_x,new_y)" +
	" VALUES (NOW(), NULL, NULL, (SELECT id FROM bots WHERE identifier = $1), NOW(), $2, $3)"

// (GET /bots)
func (Server) GetBots(ctx context.Context, request api.GetBotsRequestObject) (api.GetBotsResponseObject, error) {
	rows, err := db.Query(ctx, botsWithActionsQuery)
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()
	ledger, err := pgx.CollectRows(rows, pgx.RowToStructByName[BotsWithActions])
	if err != nil {
		logger.Fatal(err)
	}
	return api.GetBots200JSONResponse(GetBotsFromLedger(ledger, time.Now())), nil
}

// (GET /bots/{botId})
func (Server) GetBotsBotId(ctx context.Context, request api.GetBotsBotIdRequestObject) (api.GetBotsBotIdResponseObject, error) {
	rows, err := db.Query(ctx, botsWithActionsAndFilterQuery, request.BotId)
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()
	ledger, err := pgx.CollectRows(rows, pgx.RowToStructByName[BotsWithActions])
	if err != nil {
		logger.Fatal(err)
	}
	bot := GetBotsFromLedger(ledger, time.Now())[0]
	return api.GetBotsBotId200JSONResponse(bot), nil
}

// (POST /bots/{botId}/mine)
func (Server) PostBotsBotIdMine(ctx context.Context, request api.PostBotsBotIdMineRequestObject) (api.PostBotsBotIdMineResponseObject, error) {
	rows, err := db.Query(ctx, botsWithActionsAndFilterQuery, request.BotId)
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()
	ledger, err := pgx.CollectRows(rows, pgx.RowToStructByName[BotsWithActions])
	if err != nil {
		logger.Fatal(err)
	}
	bot := GetBotsFromLedger(ledger, time.Now())[0]
	var currentMine *api.Coordinates = nil
	for _, mine := range GetMinesFromDB() {
		if InRange(bot.Coordinates.X, mine.X) && InRange(bot.Coordinates.Y, mine.Y) {
			currentMine = &mine
		}
	}
	if currentMine == nil {
		return api.PostBotsBotIdMine422TextResponse("Bot is not currently near a mine"), nil
	} else {
		// Add scrap metal to bot's inventory.
		// Then delete the mine and create a new one.
		batch := &pgx.Batch{}
		batch.Queue(
			"UPDATE bots SET inventory_count = inventory_count + 1 updated_at = NOW() WHERE identifier = $1",
			bot.Identifier,
		)
		x, y := NewRandomCoordinates(-100, 100)
		batch.Queue(
			BuildInsert("mines", "x", "y"), x, y,
		)
		err := db.SendBatch(global_ctx, batch).Close()
		if err != nil {
			logger.Fatal(err)
		}
		return api.PostBotsBotIdMine200JSONResponse(bot), nil
	}
}

// (POST /bots/{botId}/newBot)
func (Server) PostBotsBotIdNewBot(ctx context.Context, request api.PostBotsBotIdNewBotRequestObject) (api.PostBotsBotIdNewBotResponseObject, error) {
	rows, err := db.Query(ctx, botsWithActionsAndFilterQuery, request.BotId)
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()
	ledger, err := pgx.CollectRows(rows, pgx.RowToStructByName[BotsWithActions])
	if err != nil {
		logger.Fatal(err)
	}
	bot := GetBotsFromLedger(ledger, time.Now())[0]
	if bot.Inventory >= 3 {
		batch := &pgx.Batch{}
		batch.Queue(
			"UPDATE bots SET inventory_count = inventory_count - 3 updated_at = NOW() WHERE identifier = $1",
			bot.Identifier,
		)
		uuid := uuid.NewString()
		batch.Queue(
			BuildInsert(
				"bots", "identifier", "inventory_count", "name",
			),
			uuid, 0, request.Body.NewBotName,
		)
		batch.Queue(
			insertMoveActionQuery,
			// TODO: Make new bot coordinates some random interval away from the making bot
			uuid, bot.Coordinates.X, bot.Coordinates.Y,
		)
		err := db.SendBatch(global_ctx, batch).Close()
		if err != nil {
			logger.Fatal(err)
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
	// TODO: Use transactions
	// tx, err := db.Begin(ctx)
	// if err != nil {
	// logger.Fatal(err)
	// }
	_, err := db.Exec(ctx,
		"DELETE FROM bot_movement_ledger WHERE 1=1 ;"+
			" DELETE FROM bots WHERE 1=1 ;"+
			" DELETE FROM mines WHERE 1=1",
	)
	if err != nil {
		logger.Fatal(err)
	}
	uuid := uuid.NewString()
	batch := &pgx.Batch{}
	batch.Queue(
		BuildInsert("bots", "Identifier", "Name"),
		uuid, "Bob",
	)
	batch.Queue(insertMoveActionQuery, uuid, 0, 0)
	mineCount := 10
	mineDistanceMax := 100.0
	mineDistanceMin := -100.0
	for range mineCount {
		x, y := NewRandomCoordinates(mineDistanceMin, mineDistanceMax)
		batch.Queue(BuildInsert("mines", "X", "Y"), x, y)
	}
	err = db.SendBatch(ctx, batch).Close()
	if err != nil {
		log.Fatal(err)
	}
	// tx.Rollback(ctx)
	// TODO: Update openapi to include bots and mines in response
	rows, err := db.Query(ctx, botsWithActionsQuery)
	if err != nil {
		logger.Fatal(err)
	}
	ledger, err := pgx.CollectRows(rows, pgx.RowToStructByName[BotsWithActions])
	if err != nil {
		logger.Fatal(err)
	}
	bot := GetBotsFromLedger(ledger, time.Now())[0]
	mines := make([]api.Coordinates, 0)
	resp := api.PostInit200JSONResponse{
		Bot:   &bot,
		Mines: &mines,
	}
	return api.PostInit200JSONResponse(resp), nil
}

// (POST /bots/{botId}/move)
func (Server) PostBotsBotIdMove(ctx context.Context, request api.PostBotsBotIdMoveRequestObject) (api.PostBotsBotIdMoveResponseObject, error) {
	status, err := db.Exec(
		global_ctx, insertMoveActionQuery,
		request.BotId, request.Body.X, request.Body.Y)
	if err != nil {
		log.Fatal(err)
	}
	if status.RowsAffected() != 1 {
		log.Fatalf(
			"Expected 1 row to be affected but %d were affected",
			status.RowsAffected())
	}
	rows, err := db.Query(
		global_ctx, botsWithActionsAndFilterQuery, request.BotId,
	)
	if err != nil {
		log.Fatal(err)
	}
	ledger, err := pgx.CollectRows(rows, pgx.RowToStructByName[BotsWithActions])
	if err != nil {
		log.Fatal(err)
	}
	resp := GetBotsFromLedger(ledger, time.Now())[0]
	return api.PostBotsBotIdMove200JSONResponse(resp), nil
}

// (GET /mines)
func (Server) GetMines(ctx context.Context, request api.GetMinesRequestObject) (api.GetMinesResponseObject, error) {
	mines := GetMinesFromDB()
	return api.GetMines200JSONResponse(mines), nil
}
