package tests

import (
	"colony-bots/api"
	"colony-bots/impl"
	"colony-bots/schemas"
	"math"
	"reflect"
	"testing"
	"time"
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
	bob := schemas.Bots{
		ID:         1,
		Identifier: "test 1",
		Name:       "Bob",
	}
	// Test a happy path of two action rows where bot reaches destination
	testLedger := []impl.BotsWithActions{
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				New_X:               0,
				New_Y:               0,
			},
		},
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 14, 0, time.UTC), // just before reaching destination
				New_X:               5,
				New_Y:               5,
			},
		},
	}
	expectedBots := []api.Bot{
		{
			Coordinates: api.Coordinates{X: 4.949747468305833, Y: 4.949747468305832},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.MOVING,
		},
	}
	// testResult := impl.GetBotsFromLedger(
	// 	testLedger, time.Date(2024, 8, 26, 11, 10, 5, 0, time.UTC),
	// )
	// if !reflect.DeepEqual(testResult, expectedBots) { // TODO: Use epsilon for floats (i am lazy)
	// 	t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	// }

	// Test three action rows.
	testLedger = []impl.BotsWithActions{
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				New_X:               0,
				New_Y:               0,
			},
		},
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 5, 0, time.UTC),
				New_X:               5,
				New_Y:               5,
			},
		},
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 25, 0, time.UTC),
				New_X:               -5,
				New_Y:               -5,
			},
		},
	}
	expectedBots = []api.Bot{
		{
			Coordinates: api.Coordinates{X: -2.0710678118654746, Y: -2.0710678118654755},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.MOVING,
		},
	}
	// testResult = impl.GetBotsFromLedger(
	// 	testLedger, time.Date(2024, 8, 26, 11, 10, 10, 0, time.UTC),
	// )
	// if !reflect.DeepEqual(testResult, expectedBots) {
	// 	t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	// }

	// Test where a subsequent action row interrupts the movement of its predecessor
	testLedger = []impl.BotsWithActions{
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				New_X:               0,
				New_Y:               0,
			},
		},
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 5, 0, time.UTC),
				New_X:               5,
				New_Y:               5,
			},
		},
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 6, 0, time.UTC),
				New_X:               -5,
				New_Y:               -5,
			},
		},
	}
	expectedBots = []api.Bot{
		{
			Coordinates: api.Coordinates{X: 4.646446609406726, Y: 4.646446609406726},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.MOVING,
		},
	}
	// testResult = impl.GetBotsFromLedger(
	// 	testLedger, time.Date(2024, 8, 26, 11, 10, 5, 1, time.UTC),
	// )
	// if !reflect.DeepEqual(testResult, expectedBots) {
	// 	t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	// }

	// Test multiple bots
	rob := schemas.Bots{
		ID:         2,
		Identifier: "test 2",
		Name:       "Rob",
	}
	testLedger = []impl.BotsWithActions{
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 0, 0, time.UTC),
				New_X:               0,
				New_Y:               0,
			},
		},
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 5, 0, time.UTC),
				New_X:               5,
				New_Y:               5,
			},
		},
		{
			Bots: bob,
			BotActions: schemas.BotActions{
				Bot_Key:             1,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 10, 0, time.UTC),
				New_X:               -5,
				New_Y:               -5,
			},
		},
		{
			Bots: rob,
			BotActions: schemas.BotActions{
				Bot_Key:             2,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 6, 0, time.UTC),
				New_X:               3,
				New_Y:               3,
			},
		},
		{
			Bots: rob,
			BotActions: schemas.BotActions{
				Bot_Key:             2,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 7, 0, time.UTC),
				New_X:               -5,
				New_Y:               -5,
			},
		},
		{
			Bots: rob,
			BotActions: schemas.BotActions{
				Bot_Key:             2,
				Time_Action_Started: time.Date(2024, 8, 26, 11, 8, 8, 0, time.UTC),
				New_X:               100,
				New_Y:               100,
			},
		},
	}
	expectedBots = []api.Bot{
		{
			Coordinates: api.Coordinates{X: -5, Y: -5},
			Identifier:  "test 1",
			Name:        "Bob",
			Status:      api.MOVING,
		},
		{
			Coordinates: api.Coordinates{X: 100, Y: 100},
			Identifier:  "test 2",
			Name:        "Rob",
			Status:      api.MOVING,
		},
	}
	testResult := impl.GetBotsFromLedger(
		testLedger, time.Date(2024, 8, 26, 11, 59, 5, 0, time.UTC),
	)
	if !reflect.DeepEqual(testResult, expectedBots) {
		t.Fatalf("Expected %#v but got %#v", expectedBots, testResult)
	}

}
