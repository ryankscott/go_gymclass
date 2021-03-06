package lm

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/PuloV/ics-golang"
	log "github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/jsgoecke/go-wit"
)

// Gyms provides a mapping of all the gyms that are available
var Gyms = []Gym{
	Gym{"city", "96382586-e31c-df11-9eaa-0050568522bb"},
	Gym{"britomart", "744366a6-c70b-e011-87c7-0050568522bb"},
	Gym{"takapuna", "98382586-e31c-df11-9eaa-0050568522bb"},
	Gym{"newmarket", "b6aa431c-ce1a-e511-a02f-0050568522bb"},
}

// Classes provides a list of all the support classes
var Classes = []string{
	// TODO: IMPLEMENT ME
}

// Config is used to store DB configuration for storing data
type Config struct {
	DBPath string
	DB     *storm.DB
}

// Gym provides a mapping between a gym's name and their unique ID
type Gym struct {
	Name string
	ID   string
}

// GymClass describes a class at Les Mills
type GymClass struct {
	UUID           string    `json:"uuid" db:"uuid" storm:"id"`
	Gym            string    `json:"gym" db:"gym" storm:"index"`
	Name           string    `json:"name" db:"name" storm:"index"`
	Location       string    `json:"location" db:"location" storm:"index"`
	StartDateTime  time.Time `json:"startdatetime" db:"start_datetime" storm:"index"`
	EndDateTime    time.Time `json:"enddatetime" db:"end_datetime" storm:"index"`
	InsertDateTime time.Time `json:"insertdatetime" db:"insert_datetime" storm:"index"`
}

// User desribes a person using a gym
type User struct {
	ID          string    `json:"sub" db:"id" storm:"id"`
	Name        string    `json:"name" db:"full_name"`
	FirstName   string    `json:"given_name" db:"first_name"`
	LastName    string    `json:"family_name" db:"last_name"`
	NickName    string    `json:"nickname" db:"nickname"`
	Gender      string    `json:"gender" db:"gender"`
	Email       string    `json:"email" db:"email"`
	Verified    bool      `json:"verified" db:"verified"`
	Locale      string    `json:"locale" db:"locale"`
	LastUpdated time.Time `json:"updated_at" db:"last_updated"`
}

// UserGymClass describes a user and all their associated classes
type UserGymClass struct {
	UserID  string     `storm:"id"`
	Classes GymClasses `storm:"index"`
}

// GymPreference describes a preference to go to a particular Gym. The preference should be a value between 0 - 1
type GymPreference struct {
	Gym        Gym     `json:"gym"`
	Preference float64 `json:"preference"`
}

// ClassPreference describes a preference to go to a particular class. The preference should be a value between 0 - 1
type ClassPreference struct {
	Class      string  `json:"class"`
	Preference float64 `json:"preference"`
}

// WorkOutFrequency describes the number of times a user went to any gym class on a particular week
type WorkOutFrequency struct {
	Week  int `json:"week"`
	Count int `json:"count"`
}

//ClassFrequency describes the number of times a user went to a particular class on a particular week

// UserStatistics describes the different statistics about a user
type UserStatistics struct {
	TotalClasses     int                `json:"totalClasses"`
	ClassesPerWeek   float64            `json:"classesPerWeek"`
	LastClassDate    time.Time          `json:"lastClassDate"`
	GymPreferences   []GymPreference    `json:"gymPreferences"`
	ClassPreferences []ClassPreference  `json:"classPreferences"`
	WorkOutFrequency []WorkOutFrequency `json:"workOutFrequency"`
}

// UserPreference describes a users preferences when going to the gym
type UserPreference struct {
	User           string `json:"user" db:"user"`
	PreferredGym   string `json:"preferredGym" db:"preferred_gym"`
	PreferredClass string `json:"preferredClass" db:"preferred_class"`
	PreferredTime  int    `json:"preferredTime" db:"preferred_time"`
	PreferredDay   int    `json:"preferredDay" db:"preferred_day"`
}

// GymQuery describes a query for GymClasses
type GymQuery struct {
	Gym    []Gym
	Class  []string
	Before time.Time
	After  time.Time
}

// ByStartDateTime implements sort.Interface for GymClasses based on the StartDateTime
type ByStartDateTime GymClasses

// GymClasses describes a collection of GymClass
type GymClasses []GymClass

// NewConfig returns a new configuration with defaults
func NewConfig() (*Config, error) {
	c := &Config{}
	c.DBPath = "gym.db"
	dbb, err := storm.Open(c.DBPath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to open database")
		return c, err
	}
	c.DB = dbb
	return c, nil
}

// InQuery checks to see if the class is within the criteria of the GymQuery
// Returns true if it meets the critieria otherwise returns false
func (g GymClass) InQuery(q GymQuery) bool {
	return compareClassName(&q, &g) && compareClassGym(&q, &g) && compareClassAfterTime(&q, &g) && compareClassBeforeTime(&q, &g)
}

// Delete will remove a GymClass from the GymClasses slice by UUID
// It will returna boolean representing the success of the operation
func (g *GymClasses) Delete(classID string) bool {
	for i, v := range *g {
		if v.UUID == classID {
			*g = append((*g)[:i], (*g)[i+1:]...)
			return true
		}
	}
	return false
}

// Exists checks to see if a GymClass is contained within the GymClasses slice
func (g GymClasses) Exists(c GymClass) bool {
	for _, v := range g {
		if c == v {
			return true
		}
	}
	return false
}

// Total returns the total number of classes in the slice
func (g GymClasses) Total() int {
	return len(g)
}

// OldestClass returns the oldest class date in the slice
func (g GymClasses) OldestClass() GymClass {
	var oldest = time.Now()
	var lc GymClass
	for _, c := range g {
		if c.StartDateTime.Before(oldest) {
			oldest = c.StartDateTime
			lc = c
		}
	}
	return lc
}

// LatestClass returns the latest class date in the slice
func (g GymClasses) LatestClass() GymClass {
	var latest = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	var lc GymClass
	for _, c := range g {
		if c.StartDateTime.After(latest) {
			latest = c.StartDateTime
			lc = c
		}
	}
	return lc
}

// PerWeek returns the number of classes per week in the slice
func (g GymClasses) PerWeek() float64 {
	lc := g.LatestClass()
	oc := g.OldestClass()
	t := float64(g.Total())

	duration := lc.StartDateTime.Sub(oc.StartDateTime)
	d := (duration.Hours()) / 24.0
	if math.IsNaN(t/d) || math.IsInf(t/d, 0) {
		return 0.0
	}
	return t / d

}

// ClassPreferences breaks down the classes by their percentage of all classes
func (g GymClasses) ClassPreferences() []ClassPreference {
	cp := make(map[string]float64)
	t := float64(g.Total())
	for _, class := range g {
		cp[class.Name]++
	}
	var c []ClassPreference
	for k, v := range cp {
		var x ClassPreference
		x.Class = k
		x.Preference = v / t
		c = append(c, x)
	}

	return c
}

// GymPreferences breaks down the classes by their percentage of all classes
func (g GymClasses) GymPreferences() []GymPreference {
	cp := make(map[string]float64)
	t := float64(g.Total())
	for _, class := range g {
		cp[class.Gym]++
	}
	var c []GymPreference
	for k, v := range cp {
		var x GymPreference
		x.Gym = GetGymByName(k)
		x.Preference = v / t
		c = append(c, x)
	}
	return c
}

// WeeklyCount returns a slice of WorkOutFrequency that shows the number of workouts per week in the collection of GymClasses
func (g GymClasses) WeeklyCount() []WorkOutFrequency {
	w := make(map[int]int)
	for _, class := range g {
		_, wk := class.StartDateTime.ISOWeek()
		w[wk]++
	}
	var c []WorkOutFrequency
	for k, v := range w {
		var x WorkOutFrequency
		x.Week = k
		x.Count = v
		c = append(c, x)
	}
	return c
}

// MostFrequentedDay returns the weekday which contains the most number of classes
func (g GymClasses) MostFrequentedDay() int {
	// Map of weekday to count
	w := make(map[int]int)
	for _, class := range g {
		wk := int(class.StartDateTime.Weekday())
		w[wk]++
	}
	max := 0
	day := 0
	for k, v := range w {
		if v > max {
			max = v
			day = k
		}
	}
	return day
}

// MostFrequentedClass returns the class type which has the most number of visits
func (g GymClasses) MostFrequentedClass() string {
	// Map of class to count
	m := make(map[string]int)
	for _, class := range g {
		m[class.Name]++
	}
	max := 0
	class := ""
	for k, v := range m {
		if v > max {
			max = v
			class = k
		}
	}
	return class
}

// MostFrequentedGym returns the gym which has the most number of visits
func (g GymClasses) MostFrequentedGym() string {
	// Map of class to count
	m := make(map[string]int)
	for _, class := range g {
		m[class.Gym]++
	}
	max := 0
	gym := ""
	for k, v := range m {
		if v > max {
			max = v
			gym = k
		}
	}
	return gym
}

// MostFrequentedTime returns the hour which has the most number of visits
func (g GymClasses) MostFrequentedTime() int {
	// Map of class to count
	m := make(map[int]int)
	for _, class := range g {
		h := class.StartDateTime.Hour()
		m[h]++
	}
	max := 0
	hour := 0
	for k, v := range m {
		if v > max {
			max = v
			hour = k
		}
	}
	return hour
}

func (a ByStartDateTime) Len() int           { return len(a) }
func (a ByStartDateTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStartDateTime) Less(i, j int) bool { return a[i].StartDateTime.Before(a[j].StartDateTime) }

func translateName(className *string) {
	switch {
	case strings.Contains(strings.ToUpper(*className), "RPM"):
		*className = "RPM"
	case strings.Contains(strings.ToUpper(*className), "GRIT STRENGTH"):
		*className = "GRIT STRENGTH"
	case strings.Contains(strings.ToUpper(*className), "GRIT CARDIO"):
		*className = "GRIT CARDIO"
	case strings.Contains(strings.ToUpper(*className), "BODYPUMP"):
		*className = "BODYPUMP"
	case strings.Contains(strings.ToUpper(*className), "BODYBALANCE"):
		*className = "BODYBALANCE"
	case strings.Contains(strings.ToUpper(*className), "BODYATTACK"):
		*className = "BODYATTACK"
	case strings.Contains(strings.ToUpper(*className), "CXWORX"):
		*className = "CXWORX"
	case strings.Contains(strings.ToUpper(*className), "SH'BAM"):
		*className = "SH'BAM"
	case strings.Contains(strings.ToUpper(*className), "BODYCOMBAT"):
		*className = "BODYCOMBAT"
	case strings.Contains(strings.ToUpper(*className), "YOGA"):
		*className = "YOGA"
	case strings.Contains(strings.ToUpper(*className), "GRIT PLYO"):
		*className = "GRIT PLYO"
	case strings.Contains(strings.ToUpper(*className), "BODYJAM"):
		*className = "BODYJAM"
	case strings.Contains(strings.ToUpper(*className), "SPRINT"):
		*className = "SPRINT"
	case strings.Contains(strings.ToUpper(*className), "BODYVIVE"):
		*className = "BODYVIVE"
	case strings.Contains(strings.ToUpper(*className), "BODYSTEP"):
		*className = "BODYSTEP"
	case strings.Contains(strings.ToUpper(*className), "BORN TO MOVE"):
		*className = "BORN TO MOVE"
	}
}

func parseICS(cal *ics.Calendar, gym Gym) (GymClasses, error) {
	log.Infof("Parsing ICS file for %s", gym.Name)
	var foundClasses GymClasses
	var foundClass GymClass
	loc, err := time.LoadLocation("Pacific/Auckland")
	if err != nil {
		log.WithFields(log.Fields{"value": err}).Error("Failed to get timezone")
		return GymClasses{}, err
	}
	for _, event := range cal.GetEvents() {
		start := event.GetStart()
		end := event.GetEnd()
		startDateTime := time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute(), start.Second(), 0, loc)
		endDateTime := time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), end.Minute(), end.Second(), 0, loc)
		name := event.GetSummary()
		translateName(&name)
		id := fmt.Sprintf("%s%s%s%s", gym.Name, name, event.GetLocation(), startDateTime)
		u := sha256.Sum256([]byte(id))
		foundClass = GymClass{
			UUID:          fmt.Sprintf("%x", u),
			Gym:           gym.Name,
			Name:          name,
			Location:      event.GetLocation(),
			StartDateTime: startDateTime,
			EndDateTime:   endDateTime,
		}
		foundClasses = append(foundClasses, foundClass)
	}
	return foundClasses, nil
}

// GetClasses will return a list of classes as stored by LesMills for the next 7 days when passing one or more Gyms
func GetClasses(gyms []Gym) (GymClasses, error) {
	baseURL := "https://www.lesmills.co.nz/timetable-calander.ashx?club="
	var foundClasses GymClasses
	parser := ics.New()
	inputChan := parser.GetInputChan()

	for _, gym := range gyms {
		// Create the URL for the ICS based on the gym
		inputChan <- baseURL + gym.ID
		log.Infof("Getting classes for %s from %s", gym.Name, baseURL+gym.ID)
	}
	parser.Wait()
	cal, err := parser.GetCalendars()
	if err != nil {
		log.WithFields(log.Fields{"value": err}).Error("Failed to get calendars")
		return nil, err
	}
	for _, c := range cal {
		gym := GetGymByID(strings.Split(c.GetUrl(), baseURL)[1])
		classes, err := parseICS(c, gym)
		if err != nil {
			log.WithFields(log.Fields{"value": err}).Error("Failed to parse ICS")
			return nil, err
		}
		foundClasses = append(foundClasses, classes...)
	}
	sort.Sort(ByStartDateTime(foundClasses))
	return foundClasses, nil
}

// StoreClasses will store a list of classes into a database based on the configuration provided
func StoreClasses(classes GymClasses, dbConfig *Config) error {
	stdClasses := 0
	for _, class := range classes {
		err := dbConfig.DB.Save(&class)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to insert class into db")
			return err
		}
		stdClasses++
	}
	log.Infof("Stored %d classes", stdClasses)
	return nil
}

// QueryUserStatistics will return a list of statistics about a user based on their usage
func QueryUserStatistics(user string, dbConfig *Config) (UserStatistics, error) {
	var us UserStatistics
	c, err := QueryUserClasses(user, dbConfig)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "user": user}).Error("Failed to get classes for user statistics")
		return UserStatistics{}, err
	}
	us.ClassPreferences = c.ClassPreferences()
	us.ClassesPerWeek = c.PerWeek()
	us.TotalClasses = c.Total()
	us.LastClassDate = c.LatestClass().StartDateTime
	us.GymPreferences = c.GymPreferences()
	us.WorkOutFrequency = c.WeeklyCount()

	return us, nil
}

// StoreUser saves a user to the database
func StoreUser(user User, dbConfig *Config) error {
	err := dbConfig.DB.Save(&user)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "row": user}).Error("Failed to insert user into db")
		return err
	}
	log.Infof("Stored user with ID: %s", user.ID)
	return nil
}

// QueryUsers returns all users in the database
func QueryUsers(dbConfig *Config) ([]User, error) {
	var users []User
	err := dbConfig.DB.All(&users)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to get all users")
		return []User{}, err
	}
	return users, nil
}

// QueryUserClasses will return a list of classes that a particular user has saved
func QueryUserClasses(user string, dbConfig *Config) (GymClasses, error) {

	var u UserGymClass
	err := dbConfig.DB.One("UserID", user, &u)
	if err == storm.ErrNotFound {
		return GymClasses{}, nil
	} else if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to get user classes")
		return GymClasses{}, err
	}

	allClasses := u.Classes
	log.Infof("Returning %d gym classes", len(allClasses))
	sort.Sort(ByStartDateTime(allClasses))

	return allClasses, nil
}

// QueryUserPreferences will return a users gym going preferences
func QueryUserPreferences(user string, dbConfig *Config) (UserPreference, error) {
	var preference UserPreference
	c, err := QueryUserClasses(user, dbConfig)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to get user classes when building preferences")
		return UserPreference{}, err
	}
	preference.PreferredClass = c.MostFrequentedClass()
	preference.PreferredDay = c.MostFrequentedDay()
	preference.PreferredGym = c.MostFrequentedGym()
	preference.PreferredTime = c.MostFrequentedTime()

	return preference, nil
}

// QueryPreferredClasses returns a list of classes based on a users preference
func QueryPreferredClasses(preference UserPreference, dbConfig *Config) (GymClasses, error) {
	// Today
	year, month, day := time.Now().UTC().Date()
	/*
	   	 | Class | Gym | Time |
	   	 |   0   |  0  |  1   | - Any class, any gym at a preferred time
	         |   0   |  1  |  0   | - Any class, preferred gym at any time
	   	 |   0   |  1  |  1   | - Any class, preferred gym at preferred time
	   	 |   1   |  0  |  0   | - Preferrred class at any gym at any time
	   	 |   1   |  0  |  1   | - Preferred class at any gym at preferred time (3)
	   	 |   1   |  1  |  0   | - Preferred class at preferred gym at any time (2)
	   	 |   1   |  1  |  1   | - Preferred class at preferred gym at preferred time (1)
	*/

	// Preferred class at preferred gym at any time
	var preferredQuery1 = GymQuery{}
	preferredQuery1.Class = []string{preference.PreferredClass}
	preferredQuery1.After = time.Now().UTC()
	preferredQuery1.Before = time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
	preferredQuery1.Gym = []Gym{GetGymByName(preference.PreferredGym)}
	queryClasses1, err := QueryClasses(preferredQuery1, dbConfig)

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to query for preferred classes for a user")
	}

	// Preferred class at any gym at preferred time (after now)
	var preferredQuery2 = GymQuery{}
	var queryClasses2 = GymClasses{}
	if time.Now().UTC().Hour() < (preference.PreferredTime - 1) {
		preferredQuery2.Class = []string{preference.PreferredClass}
		preferredQuery1.After = time.Date(year, month, day, preference.PreferredTime-1, 0, 0, 0, time.UTC)
		preferredQuery1.Before = time.Date(year, month, day, preference.PreferredTime+1, 0, 0, 0, time.UTC)
		queryClasses2, err = QueryClasses(preferredQuery2, dbConfig)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to query for preferred classes for a user")
		}

	}

	allClasses := append(queryClasses1, queryClasses2...)

	var encountered = map[string]bool{}
	var deDuped = GymClasses{}
	for _, class := range allClasses {
		if encountered[class.UUID] == true {

		} else {
			encountered[class.UUID] = true
			deDuped = append(deDuped, class)
		}
	}
	sort.Sort(ByStartDateTime(deDuped))
	return deDuped, nil
}

// StoreUserClass will store a class against a user in the database
func StoreUserClass(user string, classID string, dbConfig *Config) error {
	// Get class from ID
	var c GymClass
	err := dbConfig.DB.One("UUID", classID, &c)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "class": classID}).Error("Failed to find class from bolt db")
		return err
	}

	var u UserGymClass
	err = dbConfig.DB.One("UserID", user, &u)
	// If the user doesn't exist then create
	if err == storm.ErrNotFound {
		u.UserID = user
		u.Classes = []GymClass{c}
		err = dbConfig.DB.Save(&u)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "user": user}).Error("Failed to save new user")
			return err
		}
		return nil
	} else if err != nil {
		log.WithFields(log.Fields{"error": err, "user": user}).Error("Failed to find classes for user")
		return err
	}
	// Update the classes for the user
	allC := u.Classes
	// If it already exists don't add it again
	if allC.Exists(c) {
		log.WithFields(log.Fields{"class": c.UUID, "user": user}).Info("Class already exists for user")
		return nil
	}

	err = dbConfig.DB.UpdateField(&UserGymClass{UserID: user}, "Classes", append(u.Classes, c))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "class": classID, "user": user}).Error("Failed to store user classes")
		return err
	}
	return nil
}

// DeleteUserClass will delete a class for a particular user in the database
func DeleteUserClass(user string, classID string, dbConfig *Config) error {
	// Get the User
	var u UserGymClass
	err := dbConfig.DB.One("UserID", user, &u)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "user": user}).Error("Failed to find user when deleting UserGymClass")
		return err
	}

	// Remove the unwanted class from the current classes
	allClasses := u.Classes
	allClasses.Delete(classID)
	// Update the UserGymClass
	err = dbConfig.DB.Update(&UserGymClass{UserID: user, Classes: allClasses})
	if err != nil {
		log.WithFields(log.Fields{"error": err, "user": user}).Error("Failed to update user when deleting UserGymClass")
		return err
	}
	return nil
}

// QueryClassesByName will take a query string and try parse out the correct query and return the results
func QueryClassesByName(query string, dbConfig *Config) (GymQuery, error) {

	log.Infof("Querying wit.ai for '%s'", query)
	accessToken := os.Getenv("WIT_ACCESS_TOKEN")
	if accessToken == "" {
		log.Error("Failed to get access token from environment vars")
		return GymQuery{}, errors.New("No access token found for Wit.ai, please set the environment variable WIT_ACCESS_TOKEN")
	}
	client := wit.NewClient(accessToken)
	request := &wit.MessageRequest{}
	request.Query = query
	result, err := client.Message(request)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to query wit.ai")
		return GymQuery{}, errors.New("Failed to query wit.ai")
	}

	var gymQuery GymQuery
	if len(result.Outcomes) >= 1 {

		outcome := result.Outcomes[0]
		class := outcome.Entities["gym_classname"]
		datetime := outcome.Entities["datetime"]
		location := outcome.Entities["gym_location"]

		// Parse Gyms
		if len(location) >= 1 {
			for _, v := range location {
				gymName := fmt.Sprintf("%v", *v.Value)
				gymQuery.Gym = append(gymQuery.Gym, GetGymByName(gymName))
			}
		} else {
			gymQuery.Gym = []Gym{}
		}

		// Parse class name
		if len(class) >= 1 {
			// For each class found
			for _, v := range class {
				// Save the class as a string
				cls := fmt.Sprintf("%v", *v.Value)
				gymQuery.Class = append(gymQuery.Class, cls)
			}
		} else {
			gymQuery.Class = []string{}
		}

		// Parse dates
		if len(datetime) >= 1 {
			// We're only going to take the first datetime
			// If it is a date range
			if *datetime[0].Type == "interval" {
				if datetime[0].From != nil {
					after, err := time.Parse("2006-01-02T15:04:05Z07:00", datetime[0].From.Value)
					if err != nil {
						gymQuery.After = time.Now()
					} else {
						gymQuery.After = after
					}
				} else {
					gymQuery.After = time.Now()
				}
				if datetime[0].To != nil {
					before, err := time.Parse("2006-01-02T15:04:05Z07:00", datetime[0].To.Value)
					if err != nil {
						gymQuery.Before = time.Now().AddDate(0, 0, 1)
					} else {
						gymQuery.Before = before
					}
				} else {
					gymQuery.Before = time.Now().AddDate(0, 0, 1)

				}
				log.Infof("Received a date interval parsing %v to %v as range %s to %s", datetime[0].From, datetime[0].To, gymQuery.After, gymQuery.Before)

				// Else if it's just a value
			} else if *datetime[0].Type == "value" {
				// If the type of time value is a day, we're going to assume it's between the time provided and 24 hours in the future
				if *datetime[0].Grain == "day" {
					dateVal := (*datetime[0].Value).(string)
					after, err := time.Parse("2006-01-02T15:04:05Z07:00", dateVal)
					if err != nil {
						gymQuery.After = time.Now()
					} else {
						gymQuery.After = after
					}
					gymQuery.Before = after.AddDate(0, 0, 1)
					// If it's a type of a week, we assume it's between the time provided and 7 days in the future
				} else if *datetime[0].Grain == "week" {
					dateVal := (*datetime[0].Value).(string)
					after, err := time.Parse("2006-01-02T15:04:05Z07:00", dateVal)
					if err != nil {
						gymQuery.After = time.Now()
					} else {
						gymQuery.After = after
					}
					gymQuery.Before = after.AddDate(0, 0, 7)
				} else {
					gymQuery.After = time.Now()
					gymQuery.Before = time.Now().AddDate(0, 0, 7)
				}
				log.Infof("Received a date with grain '%v' parsing %s as range %v to %v", *datetime[0].Grain, (*datetime[0].Value).(string), gymQuery.After, gymQuery.Before)
			}
		} else {
			gymQuery.After = time.Now().AddDate(0, 0, 0)
			gymQuery.Before = time.Now().AddDate(0, 0, 7)
			log.Infof("Couldn't find a datetime so parsing as range %v to %v", gymQuery.After, gymQuery.Before)
		}

		log.Infof("Returning the following query: %v", gymQuery)
		return gymQuery, nil

	}

	log.Info("Failed to get a response from wit.ai")
	return GymQuery{}, errors.New("Failed to find any classes")
}

// QueryClasses will query the classes from the stored database and return the results
func QueryClasses(query GymQuery, dbConfig *Config) (GymClasses, error) {
	allClasses := make(GymClasses, 0)
	var gc GymClasses
	err := dbConfig.DB.All(&gc)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to get all stored classes")
		return GymClasses{}, err
	}
	for _, c := range gc {
		if c.InQuery(query) {
			allClasses = append(allClasses, c)
		}
	}

	log.Infof("Returning %d gym classes", len(allClasses))
	sort.Sort(ByStartDateTime(allClasses))
	return allClasses, nil
}

// GetGymByName returns a Gym based on the name provided
func GetGymByName(name string) Gym {
	for _, gym := range Gyms {
		if name == gym.Name {
			return gym
		}
	}
	log.WithFields(log.Fields{"name": name}).Info("Unable to find gym")
	return Gym{}
}

// GetGymByID returns a Gym based on the ID provided
func GetGymByID(ID string) Gym {
	for _, gym := range Gyms {
		if ID == gym.ID {
			return gym
		}
	}
	log.WithFields(log.Fields{"ID": ID}).Info("Unable to find gym")
	return Gym{}
}

func compareClassName(query *GymQuery, class *GymClass) bool {
	if len(query.Class) == 0 {
		return true
	}
	for _, c := range query.Class {
		if strings.Contains(strings.ToLower(class.Name), strings.ToLower(c)) {
			return true
		}
	}
	return false
}

func compareClassGym(query *GymQuery, class *GymClass) bool {
	if len(query.Gym) == 0 {
		return true
	}
	for _, g := range query.Gym {
		if strings.ToLower(class.Gym) == strings.ToLower(g.Name) {
			return true
		}
	}
	return false
}

func compareClassAfterTime(query *GymQuery, class *GymClass) bool {

	if query.After.IsZero() {
		return true
	}
	return class.StartDateTime.After(query.After)
}

func compareClassBeforeTime(query *GymQuery, class *GymClass) bool {

	if query.Before.IsZero() {
		return true
	}
	return class.StartDateTime.Before(query.Before)
}

func init() {
	debug := os.Getenv("DEBUG")
	if debug == "true" {
		log.WithFields(log.Fields{"value": "DEBUG"}).Info("Setting log level to debug")
		log.SetLevel(log.DebugLevel)
	} else {
		log.Info("Setting log level to info")
		log.SetLevel(log.InfoLevel)
	}
	log.SetOutput(os.Stdout)
}
