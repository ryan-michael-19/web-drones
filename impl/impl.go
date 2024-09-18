package impl

import (
	"colony-bots/api"
	"colony-bots/schemas"
	"context"
	"log"
	"math"
	"strings"

	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

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
	fmt.Println(sqlString)
	return sqlString
}

func BuildBotAction(ctx context.Context, bot_uuid string, x int64, y int64) {
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
var ctx = context.Background()
var logger = BuildLogger()

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
				ledger[i].Time_Action_Started,
				ledger[i+1].Time_Action_Started,
				0.5,
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
				ledger[i].Time_Action_Started,
				currentDatetime,
				0.5,
			)
			if err != nil {
				logger.Fatal(err)
			}
			bot := api.Bot{
				Coordinates: currentBotCoords,
				Identifier:  ledger[i].Identifier,
				Name:        ledger[i].Name,
				Status:      api.MOVING,
			}
			bots = append(bots, bot)
		}
	}
	return bots
}

// (GET /bots)
func (Server) GetBots(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(ctx,
		"SELECT bots.Identifier, bots.Name,"+
			" bot_actions.new_x, bot_actions.new_y, bot_actions.time_action_started"+
			" FROM bots"+
			" LEFT JOIN bot_actions ON bots.ID = bot_actions.bot_id"+
			" ORDER BY bot_actions.Time_Action_Started ASC",
	)
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()
	ledger, err := pgx.CollectRows(rows, pgx.RowToStructByName[BotsWithActions])
	if err != nil {
		logger.Fatal(err)
	}
	resp := GetBotsFromLedger(ledger, time.Now())
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
	// TODO: Use transactions
	// tx, err := db.Begin(ctx)
	// if err != nil {
	// logger.Fatal(err)
	// }
	_, err := db.Exec(ctx,
		"DELETE FROM bot_actions WHERE 1=1 ;"+
			" DELETE FROM bots WHERE 1=1 ;"+
			" DELETE FROM mines WHERE 1=1",
	)
	if err != nil {
		logger.Fatal(err)
	}
	batch := &pgx.Batch{}
	batch.Queue(
		BuildInsert("bots", "Identifier", "Name"),
		"definitely-a-uuid", "Big Chungus",
	)
	batch.Queue(
		"INSERT INTO bot_actions (created_at,updated_at,deleted_at,bot_id,time_action_started,new_x,new_y)"+
			" VALUES (NOW(), NULL, NULL, (SELECT id FROM bots WHERE identifier = $1), NOW(), $2, $3)",
		"definitely-a-uuid", 55, 80,
	)
	mineCount := 10
	for range mineCount {
		batch.Queue(
			BuildInsert("mines", "X", "Y"),
			rand.Float64(), rand.Float64(),
		)
	}
	err = db.SendBatch(ctx, batch).Close()
	if err != nil {
		log.Fatal(err)
	}
	// tx.Rollback(ctx)
	// TODO: Update openapi to include bots and mines in response
	rows, err := db.Query(ctx,
		"SELECT bots.Identifier, bots.Name, bot_actions.Time_Action_Started, bot_actions.New_X, bot_actions.New_Y"+
			" FROM bots JOIN bot_actions ON bot_actions.bot_id = bots.id",
	)
	if err != nil {
		logger.Fatal(err)
	}
	resp, err := pgx.CollectRows(rows, pgx.RowToStructByName[BotsWithActions])
	if err != nil {
		logger.Fatal(err)
	}
	// tx.Commit(ctx)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// (POST /bots/{botId}/move)
func (Server) PostBotsBotIdMove(w http.ResponseWriter, r *http.Request, botId string) {
	// TODO: IMPLEMENT THIS
	var bot schemas.Bots
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
