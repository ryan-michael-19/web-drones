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
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct{}

func NewServer() Server {
	return Server{}
}

// func openDB() *gorm.DB {
func openDB() *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), "postgres://gorm:gorm@localhost:5432/gorm")
	if err != nil {
		log.Fatal(err)
	}
	return conn
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

func BuildLogger() *log.Logger {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return log.Default()
}

var db = openDB()
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
	// id := "Totally random id"
	// name := "Beep Boop"
	// status := api.IDLE
	// coords := api.Coordinates{X: 6, Y: 45}
	// resp := []api.Bot{
	// {
	// Coordinates: coords,
	// Identifier:  id,
	// Name:        name,
	// Status:      status,
	// },
	// }
	type BotsWithActions struct {
		schemas.Bots
		schemas.BotActions
	}
	// var bots []BotsWithActions
	// db.Table("bots").
	// Select("bots.ID, bots.Identifier, bots.Name, bots.Status, bots.X, bots.Y").
	// Joins("JOIN bot_actions ON bots.ID = bot_actions.bot_id").
	// Find(&bots)
	rows, err := db.Query(ctx,
		"SELECT bots.Identifier, bots.Name, bots.Status, bots.X, bots.Y,"+
			" bot_actions.New_X, bot_actions.New_Y, bot_actions.Time_Action_Started,"+
			" FROM bots"+
			" JOIN bot_actions ON bots.ID = bot_actions.bot_id",
	)
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()
	if err != nil {
		logger.Fatal(err)
	}
	var resp []api.Bot
	now := time.Now()
	// TODO: Change to CollectRows
	for rows.Next() {
		var bot BotsWithActions
		rows.Scan(
			&bot.Identifier, &bot.Name, &bot.Status, &bot.X, &bot.Y,
			&bot.NewX, &bot.NewY, &bot.TimeActionStarted,
		)
		loc, err := GetBotLocation(
			api.Coordinates{X: bot.X, Y: bot.Y},
			api.Coordinates{X: float64(bot.NewX), Y: float64(bot.NewY)},
			bot.TimeActionStarted, now,
			0.5,
		)
		if err != nil {
			fmt.Println(err)
			loc = api.Coordinates{X: math.NaN(), Y: math.NaN()}
		}
		resp = append(resp, api.Bot{
			Coordinates: loc,
			Identifier:  bot.Identifier,
			Name:        bot.Name,
			Status:      bot.Status,
		})
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
	_, err := db.Exec(ctx,
		"DELETE FROM bots WHERE 1=1 ;"+
			" DELETE FROM bot_actions WHERE 1=1 ;"+
			" DELETE FROM mines WHERE 1=1",
	)
	if err != nil {
		logger.Fatal(err)
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		logger.Fatal(err)
	}
	defer tx.Rollback(ctx)
	batch := &pgx.Batch{}
	// _, err = tx.Exec(ctx,
	batch.Queue(
		BuildInsert("bots", "Identifier", "Name", "Status", "X", "Y"),
		"definitely-a-uuid", "Big Chungus", api.IDLE, 5, 30,
	)
	mineCount := 10
	for range mineCount {
		batch.Queue(
			BuildInsert("mines", "X", "Y"),
			rand.Float64(), rand.Float64(),
		)
	}
	db.SendBatch(ctx, batch)
	// TODO: Update openapi to include bots and mines in response
	rows, err := db.Query(ctx,
		"SELECT Identifier, Name, Status, X, Y FROM bots",
	)
	if err != nil {
		logger.Fatal(err)
	}
	resp, err := pgx.CollectRows(rows, pgx.RowToStructByName[schemas.Bots])
	if err != nil {
		logger.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// (POST /bots/{botId}/move)
func (Server) PostBotsBotIdMove(w http.ResponseWriter, r *http.Request, botId string) {
	var bot schemas.Bots
	// db.First(&bot, "Identifier = ?", botId)
	// db.Table("bot_actions").In
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
