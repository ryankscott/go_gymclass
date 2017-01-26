package lm

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/PuloV/ics-golang"
	"github.com/satori/go.uuid"
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
var loc, _ = time.LoadLocation("Pacific/Auckland")
var testClasses = []GymClass{
	{UUID: uuid.FromStringOrNil("2d50d47a-e355-11e6-ac91-5cf9388e20a4"), Gym: "city", Name: "BODYPUMP", Location: "Studio 1", StartDateTime: time.Date(2016, 12, 18, 8, 10, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 10, 0, 0, loc), InsertDateTime: time.Time{}},
	{UUID: uuid.FromStringOrNil("2d50d480-e355-11e6-ac91-5cf9388e20a4"), Gym: "city", Name: "RPM", Location: "RPM Studio", StartDateTime: time.Date(2016, 12, 18, 8, 20, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 05, 0, 0, loc), InsertDateTime: time.Time{}},
	{UUID: uuid.FromStringOrNil("2d50d483-e355-11e6-ac91-5cf9388e20a4"), Gym: "city", Name: "CXWORX", Location: "Studio 2", StartDateTime: time.Date(2016, 12, 18, 9, 0, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 9, 30, 0, 0, loc), InsertDateTime: time.Time{}},
	{UUID: uuid.FromStringOrNil("2d50d486-e355-11e6-ac91-5cf9388e20a4"), Gym: "city", Name: "BODYBALANCE", Location: "Studio 1", StartDateTime: time.Date(2016, 12, 18, 9, 10, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 10, 10, 0, 0, loc), InsertDateTime: time.Time{}},
	{UUID: uuid.FromStringOrNil("2d56ed4a-e355-11e6-ac91-5cf9388e20a4"), Gym: "city", Name: "RPM", Location: "RPM Studio", StartDateTime: time.Date(2016, 12, 18, 9, 20, 0, 0, loc), EndDateTime: time.Date(2016, 12, 18, 10, 20, 0, 0, loc), InsertDateTime: time.Time{}}}

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

func TestStoreClasses(t *testing.T) {
	testConfig, err := NewConfig()
	if err != nil {
		t.Errorf("Failed to create database %s", err)
	}
	err = StoreClasses(testClasses, testConfig)
	if err != nil {
		t.Errorf("Error when storing classes %s", err)
	}
}

type queryClassTest struct {
	query              GymQuery
	expectedClassCount int
}

func TestQueryClasses(t *testing.T) {
	testConfig, err := NewConfig()
	if err != nil {
		t.Errorf("Failed to create database %s", err)
	}
	loc, _ := time.LoadLocation("Pacific/Auckland")
	queryClassTests := []queryClassTest{
		{query: GymQuery{Gym: Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}, Class: "RPM", Before: time.Date(2099, 01, 01, 01, 01, 01, 01, loc), After: time.Date(2000, 0, 0, 0, 0, 0, 0, loc), Limit: "100"}, expectedClassCount: 2},
		{query: GymQuery{Gym: Gym{"britomart", "96382586-e31c-df11-9eaa-0050568522bb"}, Class: "RPM", Before: time.Date(2099, 01, 01, 01, 01, 01, 01, loc), After: time.Date(2000, 0, 0, 0, 0, 0, 0, loc), Limit: "100"}, expectedClassCount: 0},
		{query: GymQuery{Gym: Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}, Class: "CXWORX", Before: time.Date(2099, 01, 01, 01, 01, 01, 01, loc), After: time.Date(2000, 0, 0, 0, 0, 0, 0, loc), Limit: "100"}, expectedClassCount: 1},
		{query: GymQuery{Gym: Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}, Class: "RPM", Before: time.Date(2099, 01, 01, 01, 01, 01, 01, loc), After: time.Date(2020, 0, 0, 0, 0, 0, 0, loc), Limit: "100"}, expectedClassCount: 0},
		{query: GymQuery{Gym: Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}, Class: "RPM", Before: time.Date(2015, 01, 01, 01, 01, 01, 01, loc), After: time.Date(2000, 0, 0, 0, 0, 0, 0, loc), Limit: "100"}, expectedClassCount: 0},
	}
	for _, test := range queryClassTests {
		classes, err := QueryClasses(test.query, testConfig)
		if err != nil {
			t.Errorf("Failed to query classes %s", err)
		}
		if len(classes) != test.expectedClassCount {
			t.Errorf("Did not get expected number of classes, expected: %d but got: %d", test.expectedClassCount, len(classes))
		}

	}

}

type storeUserClassTest struct {
	user  string
	class GymClass
}

func TestStoreUserClass(t *testing.T) {
	testConfig, err := NewConfig()

	if err != nil {
		t.Errorf("Failed to create database %s", err)
	}
	allClasses, _ := QueryClasses(GymQuery{Gym: Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}, Class: "", Before: time.Date(2099, 01, 01, 01, 01, 01, 01, loc), After: time.Date(2000, 0, 0, 0, 0, 0, 0, loc), Limit: "100"}, testConfig)

	storeUserClassTests := []storeUserClassTest{
		{"123", allClasses[0]},
		{"123", allClasses[1]},
		{"123", allClasses[2]},
		{"456", allClasses[3]},
	}
	for _, test := range storeUserClassTests {
		err := StoreUserClass(test.user, test.class.UUID, testConfig)
		if err != nil {
			t.Errorf("Failed to store user class %s", err)
		}
	}
}

type queryUserClassTest struct {
	user               string
	expectedClassCount int
}

func TestQueryUserClasses(t *testing.T) {
	testConfig, err := NewConfig()

	if err != nil {
		t.Errorf("Failed to create database %s", err)
	}
	queryUserClassTests := []queryUserClassTest{
		{"123", 3},
		{"456", 1},
		{"789", 0},
	}
	for _, test := range queryUserClassTests {
		actualClasses, err := QueryUserClasses(test.user, testConfig)
		if err != nil {
			t.Errorf("Failed to get user classes %s", err)
		}
		fmt.Println(actualClasses)
		if len(actualClasses) != test.expectedClassCount {
			t.Errorf("Did not get expected number of user classes got %d but expected %d", len(actualClasses), test.expectedClassCount)
		}
	}

}

type deleteUserClassTest struct {
	user  string
	class GymClass
}

func TestDeleteUserClass(t *testing.T) {
	testConfig, err := NewConfig()

	if err != nil {
		t.Errorf("Failed to create database %s", err)
	}

	allClasses, _ := QueryClasses(GymQuery{Gym: Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}, Class: "", Before: time.Date(2099, 01, 01, 01, 01, 01, 01, loc), After: time.Date(2000, 0, 0, 0, 0, 0, 0, loc), Limit: "100"}, testConfig)
	fmt.Println(len(allClasses))

	deleteUserClassTests := []deleteUserClassTest{
		{"123", allClasses[0]},
		{"123", allClasses[1]},
		{"123", allClasses[2]},
		{"456", allClasses[3]},
	}
	for _, test := range deleteUserClassTests {
		err := DeleteUserClass(test.user, test.class.UUID, testConfig)
		if err != nil {
			t.Errorf("Failed to get user classes %s", err)
		}
	}

}
