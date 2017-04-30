package lm

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	ics "github.com/PuloV/ics-golang"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

var now = time.Now().UTC()
var testClasses = []GymClass{
	{
		UUID:           uuid.FromStringOrNil("2d50d47a-e355-11e6-ac91-5cf9388e20a4"),
		Gym:            "city",
		Name:           "BODYPUMP",
		Location:       "Studio 1",
		StartDateTime:  time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC),
		EndDateTime:    time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC),
		InsertDateTime: time.Time{},
	},
	{
		UUID:           uuid.FromStringOrNil("2d50d480-e355-11e6-ac91-5cf9388e20a5"),
		Gym:            "city",
		Name:           "RPM",
		Location:       "RPM Studio",
		StartDateTime:  time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC),
		EndDateTime:    time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC),
		InsertDateTime: time.Time{},
	},
	{
		UUID:           uuid.FromStringOrNil("2d50d483-e355-11e6-ac91-5cf9388e20a6"),
		Gym:            "city",
		Name:           "RPM",
		Location:       "RPM Studio",
		StartDateTime:  time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC),
		EndDateTime:    time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+2, 0, 0, 0, time.UTC),
		InsertDateTime: time.Time{},
	},
	{
		UUID:           uuid.FromStringOrNil("2d50d486-e355-11e6-ac91-5cf9388e20a7"),
		Gym:            "city",
		Name:           "BODYBALANCE",
		Location:       "Studio 1",
		StartDateTime:  time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+3, 0, 0, 0, time.UTC),
		EndDateTime:    time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+4, 0, 0, 0, time.UTC),
		InsertDateTime: time.Time{}},
	{
		UUID:           uuid.FromStringOrNil("2d56ed4a-e355-11e6-ac91-5cf9388e20a8"),
		Gym:            "city",
		Name:           "CXWORX",
		Location:       "Studio 2",
		StartDateTime:  time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+4, 0, 0, 0, time.UTC),
		EndDateTime:    time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+5, 0, 0, 0, time.UTC),
		InsertDateTime: time.Time{},
	},
	{
		UUID:           uuid.FromStringOrNil("4175B894-F02E-4DA2-BA7B-563307B2D8A9"),
		Gym:            "britomart",
		Name:           "RPM",
		Location:       "RPM Studio",
		StartDateTime:  time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC),
		EndDateTime:    time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+2, 0, 0, 0, time.UTC),
		InsertDateTime: time.Time{}},
}

func init() {
	_ = os.Setenv("DEBUG", "true")
}

func TestGetClasses(t *testing.T) {
	c, err := GetClasses(Gyms)
	fmt.Printf("Returned %d classes from Les Mills\n", len(c))
	assert.NoError(t, err, "Got an error when returning classes")
	assert.Condition(t, func() (success bool) { return len(c) > 10 }, "Received less than 10 classes")
}

type parseICSTest struct {
	icsPath  string
	gym      Gym
	expected []GymClass
}

// Compares two slices of GymClasses but ignores UUIDs
func compareGymClasses(a []GymClass, b []GymClass) (success bool) {
	if len(a) != len(b) {
		return false
	}
	for k, _ := range a {
		if a[k].Gym != b[k].Gym {
			fmt.Printf("Gyms not equal:\n %s %s\n", a[k], b[k])
			return false
		}
		if a[k].Name != b[k].Name {
			fmt.Printf("Names not equal:\n %s %s\n", a[k], b[k])
			return false
		}
		if a[k].Location != b[k].Location {
			fmt.Printf("Locations not equal:\n %s %s\n", a[k], b[k])
			return false
		}
		if !(a[k].StartDateTime.Equal(b[k].StartDateTime)) {
			fmt.Printf("StartDateTime not equal:\n %s %s\n", a[k], b[k])
			return false
		}
		if !(a[k].EndDateTime.Equal(b[k].EndDateTime)) {
			fmt.Printf("EndDateTime not equal:\n %s %s\n", a[k], b[k])
			return false
		}

	}
	return true

}

// Compares two UserStatistics ignoring map order
func compareUserStatistics(a UserStatistics, b UserStatistics) (success bool) {
	if a.ClassesPerWeek != b.ClassesPerWeek {
		fmt.Printf("Classes per week were not the same: %v, %v\n", a.ClassesPerWeek, b.ClassesPerWeek)
		return false
	}
	if !a.LastClassDate.Equal(b.LastClassDate) {
		fmt.Printf("LastClassDate were not the same: %v, %v\n", a.LastClassDate, b.LastClassDate)
		return false
	}
	if a.TotalClasses != b.TotalClasses {
		fmt.Printf("TotalClasses were not the same: %v, %v\n", a.TotalClasses, b.TotalClasses)
		return false
	}
	// These need to iterate through the slice and deep equal each one
	if !reflect.DeepEqual(a.ClassPreferences, b.ClassPreferences) {
		fmt.Printf("ClassPreferences were not the same: %v, %v\n", a.ClassPreferences, b.ClassPreferences)
		return false
	}
	if !reflect.DeepEqual(a.WorkOutFrequency, b.WorkOutFrequency) {
		fmt.Printf("WorkOutFrequency were not the same: %v, %v\n", a.WorkOutFrequency, b.WorkOutFrequency)
		return false
	}
	if !reflect.DeepEqual(a.GymPreferences, b.GymPreferences) {
		fmt.Printf("GymPreferences were not the same: %v, %v\n", a.GymPreferences, b.GymPreferences)
		return false
	}
	return true
}

func TestParseICS(t *testing.T) {
	parseICSTests := []parseICSTest{
		{
			icsPath: "city.ics",
			gym:     Gym{Name: "city", ID: "96382586-e31c-df11-9eaa-0050568522bb"},
			expected: []GymClass{
				{
					Gym:            "city",
					Name:           "BODYPUMP",
					Location:       "Studio 1",
					StartDateTime:  time.Date(2016, 12, 18, 8, 10, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 9, 10, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "city",
					Name:           "RPM",
					Location:       "RPM Studio",
					StartDateTime:  time.Date(2016, 12, 18, 8, 20, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 9, 05, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "city",
					Name:           "CXWORX",
					Location:       "Studio 2",
					StartDateTime:  time.Date(2016, 12, 18, 9, 0, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 9, 30, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "city",
					Name:           "BODYBALANCE",
					Location:       "Studio 1",
					StartDateTime:  time.Date(2016, 12, 18, 9, 10, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 10, 10, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "city",
					Name:           "RPM",
					Location:       "RPM Studio",
					StartDateTime:  time.Date(2016, 12, 18, 9, 20, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 10, 20, 0, 0, time.Local),
					InsertDateTime: time.Time{}}}},
		{
			icsPath: "newmarket.ics",
			gym:     Gym{Name: "newmarket", ID: ""},
			expected: []GymClass{
				{
					Gym:            "newmarket",
					Name:           "BODYPUMP",
					Location:       "Studio 2",
					StartDateTime:  time.Date(2016, 12, 18, 8, 0, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 9, 0, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "newmarket",
					Name:           "RPM",
					Location:       "CHAIN Studio",
					StartDateTime:  time.Date(2016, 12, 18, 8, 30, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 9, 15, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "newmarket",
					Name:           "BODYBALANCE",
					Location:       "Studio 1",
					StartDateTime:  time.Date(2016, 12, 18, 9, 0, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 10, 0, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "newmarket",
					Name:           "CXWORX",
					Location:       "Studio 2",
					StartDateTime:  time.Date(2016, 12, 18, 9, 30, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 10, 0, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "newmarket",
					Name:           "CXWORX",
					Location:       "Studio 2",
					StartDateTime:  time.Date(2016, 12, 25, 17, 45, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 25, 18, 15, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
			},
		},
		{
			icsPath: "takapuna.ics",
			gym:     Gym{Name: "takapuna", ID: ""},
			expected: []GymClass{
				{
					Gym:            "takapuna",
					Name:           "RPM",
					Location:       "RPM Studio",
					StartDateTime:  time.Date(2016, 12, 18, 7, 0, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 7, 30, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "takapuna",
					Name:           "RPM",
					Location:       "RPM Studio",
					StartDateTime:  time.Date(2016, 12, 18, 8, 0, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 8, 45, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "takapuna",
					Name:           "BODYBALANCE",
					Location:       "Studio 1",
					StartDateTime:  time.Date(2016, 12, 18, 8, 0, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 8, 55, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "takapuna",
					Name:           "BODYPUMP",
					Location:       "Studio 1",
					StartDateTime:  time.Date(2016, 12, 18, 9, 0, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 9, 55, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
				{
					Gym:            "takapuna",
					Name:           "RPM",
					Location:       "RPM Studio",
					StartDateTime:  time.Date(2016, 12, 18, 9, 15, 0, 0, time.Local),
					EndDateTime:    time.Date(2016, 12, 18, 9, 45, 0, 0, time.Local),
					InsertDateTime: time.Time{}},
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
			assert.Condition(t, func() (success bool) { return compareGymClasses(classes, test.expected) }, "Did not receive the expected classes")
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
	defer testConfig.DB.Close()
}

type queryClassTest struct {
	Name               string
	Query              GymQuery
	ExpectedClassCount int
}

func TestQueryClasses(t *testing.T) {
	testConfig, err := NewConfig()
	if err != nil {
		t.Errorf("Failed to create database %s", err)
		return
	}
	defer testConfig.DB.Close()

	queryClassTests := []queryClassTest{
		{
			Name: "Good Gym and Class - Expected classes",
			Query: GymQuery{
				Gym:    []Gym{Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}},
				Class:  []string{"RPM"},
				Before: time.Date(2099, 01, 01, 01, 01, 01, 01, time.UTC),
				After:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)},
			ExpectedClassCount: 2},
		{
			Name: "Bad Gym, Good Class - No classes",
			Query: GymQuery{
				Gym:    []Gym{Gym{"takapuna", "98382586-e31c-df11-9eaa-0050568522bb"}},
				Class:  []string{"RPM"},
				Before: time.Date(2099, 01, 01, 01, 01, 01, 01, time.UTC),
				After:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)},
			ExpectedClassCount: 0},
		{
			Name: "Good Gym, Good class - Expected classes",
			Query: GymQuery{
				Gym:    []Gym{Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}},
				Class:  []string{"CXWORX"},
				Before: time.Date(2099, 01, 01, 01, 01, 01, 01, time.UTC),
				After:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)},
			ExpectedClassCount: 1},
		{
			Name: "Good Gym, Good Class, In future - No classes",
			Query: GymQuery{
				Gym:    []Gym{Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}},
				Class:  []string{"RPM"},
				Before: time.Date(2099, 01, 01, 01, 01, 01, 01, time.UTC),
				After:  time.Date(2020, 0, 0, 0, 0, 0, 0, time.UTC)},
			ExpectedClassCount: 0},
		{
			Name: "Good Gym, Good Class, In past - No classes",
			Query: GymQuery{
				Gym:    []Gym{Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}},
				Class:  []string{"RPM"},
				Before: time.Date(2015, 01, 01, 01, 01, 01, 01, time.UTC),
				After:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)},
			ExpectedClassCount: 0},
		{
			Name: "Good Gym, Multiple Classes - Expected classes",
			Query: GymQuery{
				Gym:    []Gym{Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}},
				Class:  []string{"RPM", "BODYPUMP"},
				Before: time.Date(2099, 01, 01, 01, 01, 01, 01, time.UTC),
				After:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)},
			ExpectedClassCount: 3},
		{
			Name: "Multiple Gym, Single Class - Expected classes",
			Query: GymQuery{
				Gym:    []Gym{Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}, Gym{"britomart", "744366a6-c70b-e011-87c7-0050568522bb"}},
				Class:  []string{"RPM"},
				Before: time.Date(2099, 01, 01, 01, 01, 01, 01, time.UTC),
				After:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)},
			ExpectedClassCount: 3},
	}

	for _, test := range queryClassTests {
		classes, err := QueryClasses(test.Query, testConfig)
		if err != nil {
			t.Errorf("Failed to query classes %s", err)
		}
		assert.Equal(t, test.ExpectedClassCount, len(classes), "Failed %s test - Did not get expected classes when querying", test.Name)

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
	defer testConfig.DB.Close()

	allClasses, _ := QueryClasses(GymQuery{
		Gym:    []Gym{Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"}},
		Class:  nil,
		Before: time.Date(2099, 01, 01, 01, 01, 01, 01, time.UTC),
		After:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)},
		testConfig)
	storeUserClassTests := []storeUserClassTest{
		{"123", allClasses[0]},
		{"123", allClasses[1]},
		{"123", allClasses[2]},
		{"123", allClasses[3]},
		{"456", allClasses[4]},
	}
	for _, test := range storeUserClassTests {
		err := StoreUserClass(test.user, test.class.UUID, testConfig)
		assert.NoError(t, err, "Failed to store user class without error")
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
	defer testConfig.DB.Close()
	queryUserClassTests := []queryUserClassTest{
		{"123", 4},
		{"456", 1},
		{"789", 0},
	}
	for _, test := range queryUserClassTests {
		actualClasses, err := QueryUserClasses(test.user, testConfig)
		if err != nil {
			t.Errorf("Failed to get user classes %s", err)
		}
		assert.Equal(t, test.expectedClassCount, len(actualClasses), "Did not get expected number of classes for user")
	}

}

type queryUserPreferencesTest struct {
	user       string
	preference UserPreference
}

func TestQueryUserPreferences(t *testing.T) {
	testConfig, err := NewConfig()
	if err != nil {
		t.Errorf("Failed to create database %s", err)
	}
	defer testConfig.DB.Close()
	queryUserPreferencesTests := []queryUserPreferencesTest{
		{
			"123",
			UserPreference{
				PreferredGym:   "city",
				PreferredClass: "RPM",
				PreferredTime:  now.Hour(),
				PreferredDay:   int(now.Weekday())},
		},
	}

	for _, test := range queryUserPreferencesTests {
		preference, err := QueryUserPreferences(test.user, testConfig)
		if err != nil {
			t.Errorf("Failed to get favourite class for user %s", err)
		}

		assert.Equal(t, test.preference, preference, "Did not get expected user preferences")
	}
}

type queryPreferredClassesTest struct {
	pref        UserPreference
	noOfClasses int
}

func TestQueryPreferredClassesTest(t *testing.T) {
	testConfig, err := NewConfig()
	if err != nil {
		t.Errorf("Failed to create database %s", err)
	}
	defer testConfig.DB.Close()

	var queryPreferredClassesTests = []queryPreferredClassesTest{
		{
			UserPreference{
				User:           "123",
				PreferredGym:   "city",
				PreferredClass: "RPM",
				PreferredTime:  now.Hour() + 2,
				PreferredDay:   int(now.Weekday())},
			3},
	}

	for _, test := range queryPreferredClassesTests {
		classes, err := QueryPreferredClasses(test.pref, testConfig)
		if err != nil {
			t.Errorf("Failed to query preferred classes %s", err)
		}
		assert.Equal(t, test.noOfClasses, len(classes), "Received wrong number of classes when finding preferred classes")

	}

}

type queryUserStatisticsTest struct {
	user  string
	stats UserStatistics
}

func TestQueryUserStatistics(t *testing.T) {
	testConfig, err := NewConfig()
	if err != nil {
		t.Errorf("Failed to create database %s", err)
	}
	defer testConfig.DB.Close()
	city := GetGymByName("city")
	_, week := now.ISOWeek()
	queryUserStatistics := []queryUserStatisticsTest{
		{
			"123",
			UserStatistics{
				TotalClasses:   4,
				ClassesPerWeek: 32,
				LastClassDate:  time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+3, 0, 0, 0, time.UTC),
				GymPreferences: []GymPreference{
					{
						Gym:        city,
						Preference: 1.0,
					},
				},
				ClassPreferences: []ClassPreference{
					{"BODYPUMP", 0.25},
					{"RPM", 0.5},
					{"BODYBALANCE", 0.25},
				},
				WorkOutFrequency: []WorkOutFrequency{
					{week, 4},
				},
			},
		},
	}

	for _, test := range queryUserStatistics {
		stats, err := QueryUserStatistics(test.user, testConfig)
		if err != nil {
			t.Errorf("Failed to get stats for user %s", err)
		}

		assert.Condition(t, func() (success bool) { return compareUserStatistics(test.stats, stats) }, "User stats were not the same as expected:\n %v \n %v", test.stats, stats)
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

	allClasses, _ := QueryClasses(GymQuery{
		Gym: []Gym{
			Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"},
		},
		Class:  nil,
		Before: time.Date(2099, 01, 01, 01, 01, 01, 01, time.UTC),
		After:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)},
		testConfig)

	deleteUserClassTests := []deleteUserClassTest{
		{"123", allClasses[0]},
		{"123", allClasses[1]},
		{"123", allClasses[2]},
		{"123", allClasses[3]},
		{"456", allClasses[4]},
	}
	for _, test := range deleteUserClassTests {
		err := DeleteUserClass(test.user, test.class.UUID, testConfig)
		assert.NoError(t, err, "Error when deleting user classes")
	}
	_ = os.Remove("gym.db")
}
