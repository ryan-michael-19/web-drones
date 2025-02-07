package stateful

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"strconv"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ryan-michael-19/web-drones/api"
	"github.com/ryan-michael-19/web-drones/utils/stateless"
	"github.com/ryan-michael-19/web-drones/webdrones/public/model"
	. "github.com/ryan-michael-19/web-drones/webdrones/public/table"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Move these to a "globals" file.
var _ = BuildLogger()
var DB = OpenDB()

// TODO: Use a config file for these?
var BotVelocity = 0.5
var MineMax = 50.0
var MineMin = -50.0

func OpenDB() *sql.DB {
	db, err := sql.Open(
		"pgx", GetDBString(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// make sure the database is up and running
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func BuildLogger() *slog.Logger {
	replace := func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == "source" {
			src := a.Value.Any().(*slog.Source)
			return slog.String("source", src.File+":"+strconv.Itoa(src.Line))
		}
		return a
	}
	logBase := path.Join(".", "logs")
	err := os.MkdirAll(logBase, os.ModePerm)
	if err != nil {
		log.Fatalf(err.Error())
	}
	logName := path.Join(logBase, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02-15-04-05")))
	logFile, err := os.OpenFile(logName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf(err.Error())
	}
	w := io.MultiWriter(logFile, os.Stdout)
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replace,
	}))
	slog.SetDefault(logger)
	return logger
}

func CreateNewUser(username string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return &stateless.AuthError{OriginalError: err}
	}
	// add new user to db
	stmt := Users.INSERT(
		Users.CreatedAt, Users.UpdatedAt, Users.Username, Users.Password,
	).VALUES(
		NOW(), NOW(), String(username), String(string(hashedPassword)),
	)
	_, err = stmt.Exec(DB)
	if err != nil {
		if err.(*pgconn.PgError).Code == "23505" {
			return &stateless.AuthError{OriginalError: err, NewError: errors.New("username already exists")}
		} else {
			return &stateless.AuthError{OriginalError: err}
		}
	}
	return nil
}

func GetDBString() string {
	hostname, present := os.LookupEnv("DB_HOSTNAME")
	if !present || hostname == "" {
		slog.Warn("DB hostname not set. Using \"localhost\".")
		hostname = "localhost"
	}
	username, present := os.LookupEnv("POSTGRES_USER")
	if !present || username == "" {
		slog.Warn("DB username not set. Using \"user\".")
		username = "user"
	}
	var password string
	passwordFile, present := os.LookupEnv("POSTGRES_PASSWORD_FILE")
	if !present || passwordFile == "" {
		slog.Warn("DB password file location not set. Attempting to use environment variable.")
		password, present = os.LookupEnv("POSTGRES_PASSWORD")
		if !present || password == "" {
			slog.Warn("DB password not set. Using \"password\".")
			password = "password"
		}
	} else { // password file exists
		dbPasswordByteArray, err := os.ReadFile(passwordFile)
		if err != nil {
			log.Fatal(err)
		}
		password = string(dbPasswordByteArray)
	}
	dbName, present := os.LookupEnv("POSTGRES_DB")
	if !present || dbName == "" {
		slog.Warn("DB name not set. Using \"webdrones\".")
		dbName = "webdrones"
	}
	dbString := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", username, password, hostname, dbName)
	return dbString
}

type InitialGameState struct {
	Bots  []api.Bot
	Mines []api.Coordinates
}

func InitGame(username string) (*InitialGameState, error) {
	tx, err := DB.Begin()
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
		_, err := stmt.Exec(DB)
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
		_, err := stmt.Exec(DB)
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
		_, err := stmt.Exec(DB)
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
			NOW(), NOW(), uuid, newBot.botName, 0, stateless.GenerateUserIDSubquery(username),
		)
		_, err = stmt.Exec(DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		moveStatement := stateless.GenerateMoveActionQuery(
			uuid, username, newBot.coords.X, newBot.coords.Y,
		)
		_, err = moveStatement.Exec(DB)
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
		x, y := stateless.NewRandomCoordinates(MineMin, MineMax)
		stmt = stmt.VALUES(
			NOW(), NOW(), SELECT(Users.ID).FROM(Users).WHERE(Users.Username.EQ(String(username))), x, y,
		)
	}
	_, err = stmt.Exec(DB)
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
	return &InitialGameState{
		Bots:  bots,
		Mines: mines,
	}, nil

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
	var ledger []stateless.BotsWithActions
	err := stmt.Query(DB, &ledger)
	if err != nil {
		return []api.Bot{}, err
	}
	res, err := stateless.GetBotsFromLedger(ledger, time.Now(), BotVelocity)
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
	err := stmt.Query(DB, &dbResults)
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
	var ledger []stateless.BotsWithActions
	err := stmt.Query(DB, &ledger)
	if err != nil {
		return api.Bot{}, err
	}
	res, err := stateless.GetBotsFromLedger(ledger, time.Now(), BotVelocity)
	if err != nil {
		return api.Bot{}, err
	}
	return res[0], nil
}
