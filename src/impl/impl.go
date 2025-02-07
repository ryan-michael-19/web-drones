package impl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"

	"github.com/ryan-michael-19/web-drones/api"
	"github.com/ryan-michael-19/web-drones/utils/stateful"
	. "github.com/ryan-michael-19/web-drones/utils/stateless"

	"time"

	"github.com/ryan-michael-19/web-drones/webdrones/public/model"
	. "github.com/ryan-michael-19/web-drones/webdrones/public/table"

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

// TODO: Use a key type here
const SESSION_VALUE = "cookie"
const USERNAME_VALUE = "username"

// (GET /)
func (Server) Get(ctx context.Context, request api.GetRequestObject) (api.GetResponseObject, error) {
	return api.Get200TextResponse("Welcome to Web Drones! Send a POST with a Basic Auth header to /newUser to play. See https://ryan-michael-19.github.io/web-drones/ for more details."), nil
}

// (GET /bots)
func (Server) GetBots(ctx context.Context, request api.GetBotsRequestObject) (api.GetBotsResponseObject, error) {
	res, err := stateful.GetBotsFromDB(ctx.Value(USERNAME_VALUE).(string))
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return api.GetBots200JSONResponse(res), nil
}

// (GET /bots/{botId})
func (Server) GetBotsBotId(ctx context.Context, request api.GetBotsBotIdRequestObject) (api.GetBotsBotIdResponseObject, error) {
	res, err := stateful.GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return api.GetBotsBotId200JSONResponse(res), nil
}

// (POST /bots/{botId}/extract)
func (Server) PostBotsBotIdExtract(ctx context.Context, request api.PostBotsBotIdExtractRequestObject) (api.PostBotsBotIdExtractResponseObject, error) {
	bot, err := stateful.GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
	if err != nil {
		// TODO: Convert to 500
		slog.Error(err.Error())
		return nil, err
	}
	username := ctx.Value(USERNAME_VALUE).(string)
	var currentMine *api.Coordinates = nil
	mines, err := stateful.GetMinesFromDB(username)
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
		tx, err := stateful.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		// TODO: Convert to jet RawStatement (can't find support for x=x+1 in jet updates)
		_, err = stateful.DB.Exec(
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
			x, y = NewRandomCoordinates(stateful.MineMin, stateful.MineMax)
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
		res, err := stmt.Exec(stateful.DB)
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
		updatedBot, err := stateful.GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		return api.PostBotsBotIdExtract200JSONResponse(updatedBot), nil
	}
}

// (POST /bots/{botId}/newBot)
func (Server) PostBotsBotIdNewBot(ctx context.Context, request api.PostBotsBotIdNewBotRequestObject) (api.PostBotsBotIdNewBotResponseObject, error) {
	bot, err := stateful.GetSingleBotFromDB(request.BotId, ctx.Value(USERNAME_VALUE).(string))
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	if bot.Inventory >= 3 {
		username := ctx.Value(USERNAME_VALUE).(string)
		tx, err := stateful.DB.Begin()
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		defer tx.Rollback()
		stateful.DB.Exec(
			// TODO: Convert to jet and remove subquery
			"UPDATE bots SET inventory_count = inventory_count - 3, updated_at = NOW() "+
				"WHERE identifier = $1 AND user_id = (SELECT id FROM users WHERE username = $2)",
			bot.Identifier, username,
		)
		uuid := uuid.NewString()
		stmt := Bots.INSERT(
			Bots.CreatedAt, Bots.UpdatedAt, Bots.Identifier, Bots.InventoryCount, Bots.Name, Bots.UserID).VALUES(
			NOW(), NOW(), uuid, 0, request.Body.NewBotName, GenerateUserIDSubquery(username),
		)
		_, err = stmt.Exec(stateful.DB)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		stmt = GenerateMoveActionQuery(
			uuid, ctx.Value(USERNAME_VALUE).(string), bot.Coordinates.X, bot.Coordinates.Y,
		)
		_, err = stmt.Exec(stateful.DB)
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
	result, err := stateful.InitGame(username)
	if err != nil {
		return nil, err
	}
	resp := api.PostInit200JSONResponse{
		Bots:  result.Bots,
		Mines: result.Mines,
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
	status, err := stmt.Exec(stateful.DB)
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
	resp, err := stateful.GetSingleBotFromDB(request.BotId, username)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return api.PostBotsBotIdMove200JSONResponse(resp), nil
}

// (GET /mines)
func (Server) GetMines(ctx context.Context, request api.GetMinesRequestObject) (api.GetMinesResponseObject, error) {
	username := ctx.Value(USERNAME_VALUE).(string)
	mines, err := stateful.GetMinesFromDB(username)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return api.GetMines200JSONResponse(mines), nil
}

// (POST /newUser)
func (Server) PostNewUser(ctx context.Context, request api.PostNewUserRequestObject) (api.PostNewUserResponseObject, error) {
	username := ctx.Value("username").(string)
	slog.Info("New user created", "username", ctx.Value("username").(string))
	result, err := stateful.InitGame(username)
	if err != nil {
		return nil, err
	}
	resp := api.PostInit200JSONResponse{
		Bots:  result.Bots,
		Mines: result.Mines,
	}
	slog.Info("Game has been reset", "username", username)
	return api.PostNewUser200JSONResponse(resp), nil
}
