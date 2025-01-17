package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/ryan-michael-19/web-drones/api"
	"github.com/ryan-michael-19/web-drones/impl"
	"github.com/ryan-michael-19/web-drones/utils/stateful"
)

var S = impl.NewServer()
var USERNAME = "test_username"

// Test initialization (rest of tests depend on initialization)
func TestInitialization(t *testing.T) {
	// Set up the user for the rest of the testing
	err := stateful.CreateNewUser("test_username", "test_password")
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println("CREATED USER")
	ctx := context.WithValue(context.Background(), impl.USERNAME_VALUE, USERNAME)
	value, err := S.PostInit(ctx, api.PostInitRequestObject{})
	if err != nil {
		t.Fatalf(err.Error())
	}
	bots := value.(api.PostInit200JSONResponse).Bots
	if len(bots) != 3 {
		t.Fatalf(
			"Expected 3 bots. Actual bots:  %v", bots,
		)
	}
	mines := value.(api.PostInit200JSONResponse).Mines
	if len(mines) != 10 {
		t.Fatalf(
			"Expected 10 mines. Actual mines:  %v", bots,
		)
	}
}

func TestGetBots(t *testing.T) {
	ctx := context.WithValue(context.Background(), impl.USERNAME_VALUE, USERNAME)
	value, err := S.GetBots(ctx, api.GetBotsRequestObject{})
	if err != nil {
		t.Fatalf(err.Error())
	}
	bots := value.(api.GetBots200JSONResponse)
	if len(bots) != 3 {
		t.Fatalf(
			"Expected 3 bots. Actual bots: %v", bots,
		)
	}
}
