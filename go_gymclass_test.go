package lm

import (
	"reflect"
	"testing"
	"time"

	"github.com/PuloV/ics-golang"
)

// func TestNewConfigDefault(t *testing.T) {
// 	expected := Config{"sqlite3", "admin", "password", "./gym.db", nil}
// 	db, _ := sql.Open(expected.DBDriver, expected.DBPath)
// 	expected.DB = db
// 	actual, err := NewConfig()
// 	if err != nil {
// 		t.Errorf("Got an error while creating config: %s", err)
// 	}
// 	spew.Dump(expected.DB)
// 	spew.Dump(actual.DB)
// 	if !reflect.DeepEqual(expected.DB, actual.DB) {
// 		t.Errorf("Failed to create config expected: %s but got: %s", expected.DB, actual.DB)
// 	}

// }

type parseICSTest struct {
	icsPath  string
	gym      Gym
	expected []GymClass
}

func TestParseICS(t *testing.T) {
	loc, _ := time.LoadLocation("Pacific/Auckland")
	parseICSTests := []parseICSTest{
		{
			icsPath: "city.ics",
			gym:     Gym{Name: "city", ID: "96382586-e31c-df11-9eaa-0050568522bb"},
			expected: []GymClass{
				{Gym: "city", Name: "BODYPUMP", Location: "Studio 1", StartDateTime: time.Date(2016, 12, 18, 8, 10, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 10, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "city", Name: "RPM", Location: "RPM Studio", StartDateTime: time.Date(2016, 12, 18, 8, 20, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 05, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "city", Name: "CXWORX", Location: "Studio 2", StartDateTime: time.Date(2016, 12, 18, 9, 0, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 30, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "city", Name: "BODYBALANCE", Location: "Studio 1", StartDateTime: time.Date(2016, 12, 18, 9, 10, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 10, 10, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "city", Name: "RPM", Location: "RPM Studio", StartDateTime: time.Date(2016, 12, 18, 9, 20, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 10, 20, 0, 0, loc), InsertDateTime: time.Time{}}}},
		{
			icsPath: "newmarket.ics",
			gym:     Gym{Name: "newmarket", ID: ""},
			expected: []GymClass{
				{Gym: "newmarket", Name: "BODYPUMP", Location: "Studio 2", StartDateTime: time.Date(2016, 12, 18, 8, 0, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 0, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "newmarket", Name: "RPM", Location: "CHAIN Studio", StartDateTime: time.Date(2016, 12, 18, 8, 30, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 15, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "newmarket", Name: "BODYBALANCE", Location: "Studio 1", StartDateTime: time.Date(2016, 12, 18, 9, 0, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 10, 0, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "newmarket", Name: "CXWORX", Location: "Studio 2", StartDateTime: time.Date(2016, 12, 18, 9, 30, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 10, 0, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "newmarket", Name: "CXWORX", Location: "Studio 2", StartDateTime: time.Date(2016, 12, 25, 17, 45, 0, 0, loc), EndDateTime: time.Date(2016, 12, 25, 18, 15, 0, 0, loc), InsertDateTime: time.Time{}},
			},
		},
		{
			icsPath: "takapuna.ics",
			gym:     Gym{Name: "takapuna", ID: ""},
			expected: []GymClass{
				{Gym: "takapuna", Name: "RPM", Location: "RPM Studio", StartDateTime: time.Date(2016, 12, 18, 7, 0, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 7, 30, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "takapuna", Name: "RPM", Location: "RPM Studio", StartDateTime: time.Date(2016, 12, 18, 8, 0, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 8, 45, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "takapuna", Name: "BODYBALANCE", Location: "Studio 1", StartDateTime: time.Date(2016, 12, 18, 8, 0, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 8, 55, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "takapuna", Name: "BODYPUMP", Location: "Studio 1", StartDateTime: time.Date(2016, 12, 18, 9, 0, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 55, 0, 0, loc), InsertDateTime: time.Time{}},
				{Gym: "takapuna", Name: "RPM", Location: "RPM Studio", StartDateTime: time.Date(2016, 12, 18, 9, 15, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 45, 0, 0, loc), InsertDateTime: time.Time{}},
			},
		},
	}
	for _, test := range parseICSTests {
		parser := ics.New()
		inputChan := parser.GetInputChan()
		inputChan <- test.icsPath
		parser.Wait()
		cal, err := parser.GetCalendars()
		if err != nil {
			t.Errorf("Got an error parsing calendar - %s", err)
		}
		for _, c := range cal {
			classes, err := parseICS(c, test.gym)
			if err != nil {
				t.Errorf("Error found when parsing ICS %s", err)
			}
			for k, v := range classes {
				if !reflect.DeepEqual(v, test.expected[k]) {
					t.Errorf("Failed to parse ICS expected: %s but got: %s", v, test.expected[k])
				}
			}
		}
	}
}
