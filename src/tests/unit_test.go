package tests

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/ryan-michael-19/web-drones/api"
	"github.com/ryan-michael-19/web-drones/impl"
	"github.com/ryan-michael-19/web-drones/webdrones/public/model"
)

func almostEquals(val1 float64, val2 float64) bool {
	return math.Abs(val1-val2) < 1e-9
}

func TestGetBotLocation(t *testing.T) {
	// I ran this function and printed the output, and spot checked that output.
	// Now I am using it as the correct test values. Low regression risk but
	// higher functional risk

	// test bot in movement
	test_coords_1, err_1 := impl.GetBotLocation(
		api.Coordinates{X: 10, Y: 20},
		api.Coordinates{X: 20, Y: 30},
		time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
		time.Date(2024, 8, 26, 11, 8, 10, 0, time.UTC),
		0.5,
	)
	if !almostEquals(test_coords_1.X, 13.535533905932738) ||
		!almostEquals(test_coords_1.Y, 23.535533905932738) {
		t.Fatalf("Expected %f, %f. Got %f, %f",
			13.535533905932738, 23.535533905932738, test_coords_1.X, test_coords_1.Y)
	}
	if err_1 != nil {
		t.Fatalf("Unexpected error: %v", err_1)
	}

	// test bot reaching location by boosting velocity
	test_coords_2, err_2 := impl.GetBotLocation(
		api.Coordinates{X: 10, Y: 20},
		api.Coordinates{X: 20, Y: 30},
		time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
		time.Date(2024, 8, 26, 11, 8, 10, 0, time.UTC),
		2.0,
	)
	if !almostEquals(test_coords_2.X, 20) ||
		!almostEquals(test_coords_2.Y, 30) {
		t.Fatalf("Expected %f, %f. Got %f, %f",
			20.0, 30.0, test_coords_2.X, test_coords_2.Y)
	}
	if err_2 != nil {
		t.Fatalf("Unexpected error: %v", err_2)
	}
	// test different quadrants
	test_coords_3, err_3 := impl.GetBotLocation(
		api.Coordinates{X: 10, Y: -20},
		api.Coordinates{X: -20, Y: -30},
		time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
		time.Date(2024, 8, 26, 11, 8, 10, 0, time.UTC),
		0.5,
	)
	if !almostEquals(test_coords_3.X, 5.256583509747431) ||
		!almostEquals(test_coords_3.Y, -21.58113883008419) {
		t.Fatalf("Expected %f, %f. Got %f, %f",
			5.256583509747431, -21.58113883008419, test_coords_3.X, test_coords_3.Y)
	}
	if err_3 != nil {
		t.Fatalf("Unexpected error: %v", err_3)
	}

	// test error when end time is before start time
	_, err_4 := impl.GetBotLocation(
		api.Coordinates{X: 10, Y: -20},
		api.Coordinates{X: -20, Y: -30},
		time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
		time.Date(2024, 8, 26, 11, 7, 0, 0, time.UTC),
		0.5,
	)
	if err_4 == nil {
		t.Fatalf("Expected error but did not get one")
	}
}

func TestGetBotsFromLedger(t *testing.T) {
	// TODO: Validate results with something better than spot checks and prayers
	bob := model.Bots{
		ID:         1,
		Identifier: "test 1",
		Name:       "Bob",
	}
	// Test ledger with single initialized bot
	testLedger := []impl.BotsWithActions{
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				NewX:              58,
				NewY:              56,
			},
		},
	}
	expectedBots := []api.Bot{
		{
			Coordinates: api.Coordinates{X: 58, Y: 56},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.IDLE,
		},
	}
	testResult, err := impl.GetBotsFromLedger(
		testLedger, time.Date(2024, 8, 26, 11, 8, 27, 0, time.UTC), 0.5, // just before reaching destination
	)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if !reflect.DeepEqual(testResult, expectedBots) { // TODO: Use epsilon for floats (i am lazy)
		t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	}

	// Test initialization of several bots
	rob := model.Bots{
		ID:         2,
		Identifier: "test 2",
		Name:       "Rob",
	}
	testLedger = []impl.BotsWithActions{
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				NewX:              58,
				NewY:              56,
			},
		},
		{
			Bots: rob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             2,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				NewX:              48,
				NewY:              36,
			},
		},
	}
	expectedBots = []api.Bot{
		{
			Coordinates: api.Coordinates{X: 58, Y: 56},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.IDLE,
		},
		{
			Coordinates: api.Coordinates{X: 48, Y: 36},
			Identifier:  "test 2",
			Name:        "Rob",
			Status:      api.IDLE,
		},
	}
	testResult, err = impl.GetBotsFromLedger(
		testLedger, time.Date(2024, 8, 26, 11, 8, 27, 0, time.UTC), 0.5, // just before reaching destination
	)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if !reflect.DeepEqual(testResult, expectedBots) { // TODO: Use epsilon for floats (i am lazy)
		t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	}

	// Test a happy path of two action rows where bot reaches destination
	testLedger = []impl.BotsWithActions{
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				NewX:              0,
				NewY:              0,
			},
		},
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 13, 0, time.UTC),
				NewX:              5,
				NewY:              5,
			},
		},
	}
	expectedBots = []api.Bot{
		{
			Coordinates: api.Coordinates{X: 4.949747468305833, Y: 4.949747468305832},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.MOVING,
		},
	}
	testResult, err = impl.GetBotsFromLedger(
		testLedger, time.Date(2024, 8, 26, 11, 8, 27, 0, time.UTC), 0.5, // just before reaching destination
	)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if !reflect.DeepEqual(testResult, expectedBots) { // TODO: Use epsilon for floats (i am lazy)
		t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	}

	// Test three action rows.
	testLedger = []impl.BotsWithActions{
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				NewX:              0,
				NewY:              0,
			},
		},
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 10, 0, 0, time.UTC),
				NewX:              5,
				NewY:              5,
			},
		},
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 12, 0, 0, time.UTC),
				NewX:              -5,
				NewY:              -5,
			},
		},
	}
	expectedBots = []api.Bot{
		{
			Coordinates: api.Coordinates{X: -5, Y: -5},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.IDLE,
		},
	}
	testResult, err = impl.GetBotsFromLedger(
		testLedger, time.Date(2024, 8, 26, 11, 14, 0, 0, time.UTC), 0.5,
	)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if !reflect.DeepEqual(testResult, expectedBots) {
		t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	}

	// Test where a subsequent action row interrupts the movement of its predecessor
	testLedger = []impl.BotsWithActions{
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				NewX:              0,
				NewY:              0,
			},
		},
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 5, 0, time.UTC),
				NewX:              5,
				NewY:              5,
			},
		},
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 10, 0, time.UTC),
				NewX:              -5,
				NewY:              -5,
			},
		},
	}
	expectedBots = []api.Bot{
		{
			Coordinates: api.Coordinates{X: -0.707106781186547, Y: -0.7071067811865477},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.MOVING,
		},
	}
	testResult, err = impl.GetBotsFromLedger(
		testLedger, time.Date(2024, 8, 26, 11, 8, 17, 0, time.UTC), 0.5,
	)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if !reflect.DeepEqual(testResult, expectedBots) {
		t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	}

	// Test multiple bots
	testLedger = []impl.BotsWithActions{
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				NewX:              0,
				NewY:              0,
			},
		},
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 5, 0, time.UTC),
				NewX:              5,
				NewY:              5,
			},
		},
		{
			Bots: bob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             1,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 10, 0, time.UTC),
				NewX:              -5,
				NewY:              -5,
			},
		},
		{
			Bots: rob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             2,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 6, 0, time.UTC),
				NewX:              3,
				NewY:              3,
			},
		},
		{
			Bots: rob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             2,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 7, 0, time.UTC),
				NewX:              -5,
				NewY:              -5,
			},
		},
		{
			Bots: rob,
			BotMovementLedger: model.BotMovementLedger{
				BotID:             2,
				TimeActionStarted: time.Date(2024, 8, 26, 11, 8, 8, 0, time.UTC),
				NewX:              100,
				NewY:              100,
			},
		},
	}
	expectedBots = []api.Bot{
		{
			Coordinates: api.Coordinates{X: -5, Y: -5},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.IDLE,
		},
		{
			Coordinates: api.Coordinates{X: 100, Y: 100},
			Identifier:  "test 2",
			Name:        "Rob",
			Status:      api.IDLE,
		},
	}
	testResult, err = impl.GetBotsFromLedger(
		testLedger, time.Date(2024, 8, 26, 11, 59, 5, 0, time.UTC), 0.05,
	)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if !reflect.DeepEqual(testResult, expectedBots) {
		t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	}

}
