package lm

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/PuloV/ics-golang"
	log "github.com/Sirupsen/logrus"
	"github.com/jsgoecke/go-wit"
	_ "github.com/mattn/go-sqlite3"
)

// TODO:

// Config is used to store DB configuration for storing data
type Config struct {
	DBDriver   string
	DBUsername string
	DBPassword string
	DBPath     string
	DB         *sql.DB
}

// NewConfig returns a new configuration with defaults
func NewConfig() (*Config, error) {
	c := &Config{}
	c.DBDriver = "sqlite3"
	c.DBUsername = "admin"
	c.DBPassword = "password"
	c.DBPath = "./gym.db"

	db, err := sql.Open(c.DBDriver, c.DBPath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to open database")
		return c, err
	}
	c.DB = db

	return c, nil
}

// Gym provides a mapping between a gym's name and their unique ID
type Gym struct {
	Name string
	ID   string
}

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

// GymClass describes a class at Les Mills
type GymClass struct {
	UUID           uuid.UUID `json:"uuid" db:"uuid"`
	Gym            string    `json:"gym" db:"gym"`
	Name           string    `json:"name" db:"name"`
	Location       string    `json:"location" db:"location"`
	StartDateTime  time.Time `json:"startdatetime" db:"start_datetime"`
	EndDateTime    time.Time `json:"enddatetime" db:"end_datetime"`
	InsertDateTime time.Time `json:"insertdatetime" db:"insert_datetime"`
}

// GymPreference describes a preference to go to a particular Gym. The preference should be a value between 0 - 1
type GymPreference struct {
	Gym        Gym     `json:"gym"`
	Preference float32 `json:"preference"`
}

// ClassPreference describes a preference to go to a particular class. The preference should be a value between 0 - 1
type ClassPreference struct {
	Class      string  `json:"class"`
	Preference float32 `json:"preference"`
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
	ClassesPerWeek   float32            `json:"classesPerWeek"`
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

// UserGymClass describes a saved GymClass by a user
type UserGymClass struct {
	User       string    `json:"user" db:"user"`
	GymClassID uuid.UUID `json:"gymClassUUID" db:"gymclass_uiid"`
}

// GymQuery describes a query for GymClasses
type GymQuery struct {
	Gym    Gym
	Class  string
	Before time.Time
	After  time.Time
	Limit  int
}

// ByStartDateTime implements sort.Interface for []GymClass based on the StartDateTime
type ByStartDateTime []GymClass

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

func parseICS(cal *ics.Calendar, gym Gym) ([]GymClass, error) {
	log.Infof("Parsing ICS file for %s", gym.Name)
	var foundClasses []GymClass
	var foundClass GymClass
	loc, err := time.LoadLocation("Pacific/Auckland")
	if err != nil {
		log.WithFields(log.Fields{"value": err}).Error("Failed to get timezone")
		return []GymClass{}, err
	}
	for _, event := range cal.GetEvents() {
		start := event.GetStart()
		end := event.GetEnd()
		startDateTime := time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute(), start.Second(), 0, loc)
		endDateTime := time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), end.Minute(), end.Second(), 0, loc)
		name := event.GetSummary()
		translateName(&name)
		foundClass = GymClass{
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
func GetClasses(gyms []Gym) ([]GymClass, error) {

	baseURL := "https://www.lesmills.co.nz/timetable-calander.ashx?club="
	var foundClasses []GymClass

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
func StoreClasses(classes []GymClass, dbConfig *Config) error {

	// Create table
	createTable := `
        CREATE TABLE IF NOT EXISTS class (
		   uuid VARCHAR(45) PRIMARY KEY,
           gym VARCHAR(9) NOT NULL,
           name VARCHAR(45) NOT NULL,
           location VARCHAR(27) NOT NULL,
           start_datetime DATETIME NOT NULL,
           end_datetime DATETIME NOT NULL,
           insert_datetime DATETIME NOT NULL);
`

	_, err := dbConfig.DB.Exec(createTable)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create table")
		return err
	}

	// Create index on table
	indexTable := `
	CREATE UNIQUE INDEX IF NOT EXISTS unique_class ON class(gym, location, start_datetime);`

	_, err = dbConfig.DB.Exec(indexTable)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to index table")
		return err
	}

	// Save all classes
	for _, class := range classes {

		// Prepare insert query
		stmt, err := dbConfig.DB.Prepare("INSERT OR IGNORE INTO class (uuid, gym, name, location, start_datetime, end_datetime, insert_datetime) values(?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to create row")
			return err
		}
		uuid := uuid.NewV1().String()

		_, err = stmt.Exec(uuid, class.Gym, class.Name, class.Location, class.StartDateTime.UTC(), class.EndDateTime.UTC(), time.Now().UTC())
		log.Infof("Executing insert query to store gym with args:\n UUID: %s\n gym: %s\n name: %s\n location: %s\n start_datetime: %s\n end_datetime: %s\n insert_datetime %s\n", uuid, class.Gym, class.Name, class.Location, class.StartDateTime.UTC(), class.EndDateTime.UTC(), time.Now().UTC())

		if err != nil {
			log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to insert row")
			return err
		}

	}
	return nil
}

// QueryUserStatistics will return a list of statistics about a user based on their usage
func QueryUserStatistics(user string, dbConfig *Config) (UserStatistics, error) {
	var stats UserStatistics
	queries := map[string]string{
		"TotalClasses":     "SELECT count(*) from user_class WHERE user = ?",
		"LastClassDate":    "SELECT c.start_datetime FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ? ORDER BY c.start_datetime DESC LIMIT 1",
		"ClassesPerWeek":   "SELECT 1.0*(strftime('%W', 'now') - strftime('%W',MIN(c.start_datetime)))/count(*) from class c inner join user_class uc where uc.class_id = c.uuid and uc.user = ?;",
		"GymPreferences":   "SELECT c.gym, 1.0*count(*)/(SELECT count(*) from class c INNER JOIN user_class uc ON uc.class_id = c.uuid where uc.user = ?) FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ? GROUP BY c.gym",
		"ClassPreferences": "SELECT c.name, 1.0*count(*)/(SELECT count(*) from class c INNER JOIN user_class uc ON uc.class_id = c.uuid where uc.user = ?) FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ? GROUP BY c.name",
		"WorkOutFrequency": "SELECT strftime('%W', datetime(start_datetime,'localtime')), count(*) as cnt FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ? GROUP BY strftime('%W', datetime(start_datetime,'localtime'))",
	}
	for key, query := range queries {
		stmt, err := dbConfig.DB.Prepare(query)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to prepare query")
			return UserStatistics{}, err
		}

		switch key {
		case "TotalClasses":
			log.Infof("Executing query for %s with args user: %s", key, user)
			row := stmt.QueryRow(user)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to query for total classes")
				return UserStatistics{}, err
			}
			err := row.Scan(&stats.TotalClasses)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to parse total classes")
			}
		case "LastClassDate":
			log.Infof("Executing query for %s with args user: %s", key, user)
			row := stmt.QueryRow(user)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to query for last class date")
				return UserStatistics{}, err
			}
			err := row.Scan(&stats.LastClassDate)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to parse last class date")
				return UserStatistics{}, err
			}
		case "ClassesPerWeek":
			log.Infof("Executing query for %s with args user: %s", key, user)
			row := stmt.QueryRow(user)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to query for classes per week")
				return UserStatistics{}, err
			}
			err := row.Scan(&stats.ClassesPerWeek)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to parse classes per week")
				return UserStatistics{}, err
			}

		case "GymPreferences":
			log.Infof("Executing query for %s with args user: %s", key, user)
			rows, err := stmt.Query(user, user)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to query for gym preferences")
				return UserStatistics{}, err
			}
			var allPref []GymPreference
			for rows.Next() {
				var pref GymPreference
				var gymName string
				err := rows.Scan(&gymName, &pref.Preference)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Error("Failed to query for gym preferences")
					return UserStatistics{}, err
				}
				pref.Gym = GetGymByName(gymName)
				allPref = append(allPref, pref)
			}
			stats.GymPreferences = allPref

		case "ClassPreferences":
			log.Infof("Executing query for %s with args user: %s", key, user)
			rows, err := stmt.Query(user, user)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to query for class preferences")
				return UserStatistics{}, err
			}
			var allPref []ClassPreference
			for rows.Next() {
				var pref ClassPreference
				err := rows.Scan(&pref.Class, &pref.Preference)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Error("Failed to query for class preferences")
					return UserStatistics{}, err
				}
				allPref = append(allPref, pref)
			}
			stats.ClassPreferences = allPref

		case "WorkOutFrequency":
			log.Infof("Executing query for %s with args user: %s", key, user)
			rows, err := stmt.Query(user)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to query for work out frequency")
				return UserStatistics{}, err
			}
			var allFreq []WorkOutFrequency
			for rows.Next() {
				var freq WorkOutFrequency
				err := rows.Scan(&freq.Week, &freq.Count)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Error("Failed to query for work out frequency")
					return UserStatistics{}, err
				}
				allFreq = append(allFreq, freq)
			}
			stats.WorkOutFrequency = allFreq
		}
	}
	return stats, nil
}

// QueryUserClasses will return a list of classes that a particular user has saved
func QueryUserClasses(user string, dbConfig *Config) ([]GymClass, error) {
	var err error
	var stmt *sql.Stmt
	stmt, err = dbConfig.DB.Prepare("SELECT c.* FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ?")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to prepare select statement")
		return []GymClass{}, err
	}
	log.Infof("Executing query with args user: %s", user)
	rows, err := stmt.Query(user)
	results := make([]GymClass, 0)
	for rows.Next() {
		var result GymClass
		err := rows.Scan(&result.UUID, &result.Gym, &result.Name, &result.Location, &result.StartDateTime, &result.EndDateTime, &result.InsertDateTime)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to marshal result")
		}
		results = append(results, result)
	}
	log.Infof("Returning %d gym classes", len(results))
	sort.Sort(ByStartDateTime(results))
	return results, nil
}

// QueryUserPreferences will return a users gym going preferences
func QueryUserPreferences(user string, dbConfig *Config) (UserPreference, error) {

	var preference UserPreference
	queries := map[string]string{
		"Day":       "SELECT strftime('%w', start_datetime), count(*) as cnt FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ? GROUP BY strftime('%w', start_datetime) ORDER BY cnt DESC LIMIT 1",
		"Class":     "SELECT c.name, count(*) as cnt FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ? GROUP BY c.name ORDER BY cnt DESC LIMIT 1",
		"Gym":       "SELECT c.gym, count(*) as cnt FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ? GROUP BY c.gym ORDER BY cnt DESC LIMIT 1",
		"StartTime": "SELECT strftime('%H', c.start_datetime), count(*) as cnt FROM class c INNER JOIN user_class uc ON uc.class_id = c.uuid WHERE uc.user = ? GROUP BY strftime('%H', c.start_datetime) ORDER BY cnt DESC LIMIT 1",
	}

	for key, query := range queries {
		stmt, err := dbConfig.DB.Prepare(query)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to prepare query")
			return UserPreference{}, err
		}

		log.Infof("Executing query for %s with args user: %s", key, user)
		row := stmt.QueryRow(user)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to query")
			return UserPreference{}, err
		}

		switch key {
		case "Day":
			err := row.Scan(&preference.PreferredDay, nil)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to parse preferred day")
			}
		case "Class":
			err := row.Scan(&preference.PreferredClass, nil)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to parse preferred class")
			}
		case "Gym":
			err := row.Scan(&preference.PreferredGym, nil)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to parse preferred gym")
			}
		case "StartTime":
			err := row.Scan(&preference.PreferredTime, nil)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to parse preferred time")
			}
		}
	}
	return preference, nil
}

// QueryPreferredClasses returns a list of classes based on a users preference
func QueryPreferredClasses(preference UserPreference, dbConfig *Config) ([]GymClass, error) {
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

	// Preferred class at preferred gym at preferred time
	var preferredQuery1 = GymQuery{}
	preferredQuery1.Class = preference.PreferredClass
	preferredQuery1.After = time.Date(year, month, day, preference.PreferredTime-1, 0, 0, 0, time.UTC)
	preferredQuery1.Before = time.Date(year, month, day, preference.PreferredTime+1, 0, 0, 0, time.UTC)
	preferredQuery1.Gym = GetGymByName(preference.PreferredGym)
	queryClasses1, err := QueryClasses(preferredQuery1, dbConfig)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to query for preferred classes for a user")
	}

	// Preferred class at preferred gym at any time
	var preferredQuery2 = GymQuery{}
	preferredQuery2.Class = preference.PreferredClass
	preferredQuery2.After = time.Now().UTC()
	preferredQuery2.Before = time.Date(year, month, day+2, 0, 0, 0, 0, time.UTC)
	preferredQuery2.Gym = GetGymByName(preference.PreferredGym)
	queryClasses2, err := QueryClasses(preferredQuery2, dbConfig)

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to query for preferred classes for a user")
	}

	// Preferred class at any gym at preferred time (after now)
	var preferredQuery3 = GymQuery{}
	var queryClasses3 = []GymClass{}
	if time.Now().UTC().Hour() < (preference.PreferredTime - 1) {
		preferredQuery3.Class = preference.PreferredClass
		preferredQuery1.After = time.Date(year, month, day, preference.PreferredTime-1, 0, 0, 0, time.UTC)
		preferredQuery1.Before = time.Date(year, month, day, preference.PreferredTime+1, 0, 0, 0, time.UTC)
		queryClasses3, err = QueryClasses(preferredQuery3, dbConfig)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to query for preferred classes for a user")
		}

	}

	allClasses := append(queryClasses1, queryClasses2...)
	allClasses = append(allClasses, queryClasses3...)

	var encountered = map[uuid.UUID]bool{}
	var deDuped = []GymClass{}
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
func StoreUserClass(user string, classID uuid.UUID, dbConfig *Config) error {

	// Create table
	createTable := `
	        CREATE TABLE IF NOT EXISTS user_class (
			   user VARCHAR(45) NOT NULL,
	           class_id VARCHAR(45) NOT NULL);
	`

	_, err := dbConfig.DB.Exec(createTable)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create table")
		return err
	}

	// Create index on table
	indexTable := `
		CREATE UNIQUE INDEX IF NOT EXISTS unique_user_class ON user_class(user, class_id);`

	_, err = dbConfig.DB.Exec(indexTable)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to index table")
		return err
	}

	// Check if that GymClass does actually exit
	stmt, err := dbConfig.DB.Prepare("SELECT COUNT(*) FROM class WHERE uuid = ?")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create search query")
		return err
	}
	row := stmt.QueryRow(classID)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to search for class")
		return err
	}

	var classes int
	err = row.Scan(&classes)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to marshal result")
	}
	if classes != 1 {
		log.WithFields(log.Fields{"class": classID, "returnedClasses": classes}).Error("Did not find exactly one class matching that ID")
		return errors.New("Not exactly one class found with that ID")
	}

	// Prepare insert query
	stmt, err = dbConfig.DB.Prepare("INSERT INTO user_class (user, class_id) values(?, ?)")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "class": classID}).Error("Failed to create row")
		return err
	}

	_, err = stmt.Exec(user, classID)
	log.Infof("Executing insert query to store user class with args:\n user: %s\n classID: %s\n", user, classID)

	if err != nil {
		log.WithFields(log.Fields{"error": err, "class": classID}).Error("Failed to insert row")
		return err
	}

	return nil
}

// DeleteUserClass will delete a class for a particular user in the database
func DeleteUserClass(user string, class uuid.UUID, dbConfig *Config) error {
	// Prepare delete query
	stmt, err := dbConfig.DB.Prepare("DELETE FROM user_class where user = ? and class_id = ?")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to delete row")
		return err
	}
	_, err = stmt.Exec(user, class)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to delete row")
		return err
	}

	return nil
}

// QueryClassesByName will take a query string and try parse out the correct query and return the results
func QueryClassesByName(query string, dbConfig *Config) ([]GymClass, error) {

	log.Infof("Querying wit.ai for '%s'", query)
	accessToken := os.Getenv("WIT_ACCESS_TOKEN")
	if accessToken == "" {
		log.Error("Failed to get access token from environment vars")
		return []GymClass{}, errors.New("No access token found for Wit.ai, please set the environment variable WIT_ACCESS_TOKEN")
	}
	client := wit.NewClient(accessToken)
	request := &wit.MessageRequest{}
	request.Query = query
	result, err := client.Message(request)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to query wit.ai")
		return []GymClass{}, errors.New("Failed to query wit.ai")
	}

	classes := make([]GymClass, 0)
	if len(result.Outcomes) >= 1 {
		outcome := result.Outcomes[0]
		class := outcome.Entities["agenda_entry"]
		datetime := outcome.Entities["datetime"]
		location := outcome.Entities["location"]

		gymQuery := GymQuery{}

		if len(location) >= 1 {
			gymName := fmt.Sprintf("%v", *location[0].Value)
			gymQuery.Gym = GetGymByName(gymName)
		} else {
			gymQuery.Gym = Gym{}
		}
		if len(class) >= 1 {
			cls := fmt.Sprintf("%v", *class[0].Value)
			cla := strings.Split(cls, " ")
			if len(cla) > 0 {
				gymQuery.Class = cla[0]
			} else {
				gymQuery.Class = cls
			}
		} else {
			gymQuery.Class = ""
		}
		if len(datetime) >= 1 {
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
					gymQuery.After = time.Now().AddDate(0, 0, 0)
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
				if *datetime[0].Grain == "day" {
					dateVal := (*datetime[0].Value).(string)
					after, err := time.Parse("2006-01-02T15:04:05Z07:00", dateVal)
					if err != nil {
						gymQuery.After = time.Now()
					} else {
						gymQuery.After = after
					}
					gymQuery.Before = after.AddDate(0, 0, 1)
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

		classes, err = QueryClasses(gymQuery, dbConfig)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to find classes")
			return []GymClass{}, errors.New("Failed to find classes")
		}
	} else {
		log.Info("Failed to get a response from wit.ai")
		classes := []GymClass{}
		return classes, errors.New("Failed to find any classes")
	}

	return classes, nil
}

// QueryClasses will query the classes from the stored database and return the results
func QueryClasses(query GymQuery, dbConfig *Config) ([]GymClass, error) {
	var err error
	var stmt *sql.Stmt
	stmt, err = dbConfig.DB.Prepare("SELECT * FROM class WHERE gym LIKE ? AND name LIKE ? and start_datetime > ? and start_datetime < ? ORDER BY uuid DESC limit ?")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to prepare select statement")
		return []GymClass{}, err
	}

	// TODO: Refactor this
	likeGym := "%" + query.Gym.Name + "%"
	likeName := "%" + query.Class + "%"
	var limit int
	if query.Limit != 0 {
		limit = query.Limit
	} else {
		limit = -1

	}
	rows, err := stmt.Query(
		strings.ToLower(likeGym),
		strings.ToLower(likeName),
		query.After,
		query.Before,
		limit,
	)
	log.Infof("Executing query with args gym: %s name: %s, start_datetime: %s, end_datetime: %s limit %s", likeGym, likeName, query.After, query.Before, limit)
	results := make([]GymClass, 0)
	for rows.Next() {
		var result GymClass
		err := rows.Scan(&result.UUID, &result.Gym, &result.Name, &result.Location, &result.StartDateTime, &result.EndDateTime, &result.InsertDateTime)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to marshal result")
		}
		translateName(&result.Name)
		results = append(results, result)
	}
	log.Infof("Returning %d gym classes", len(results))
	sort.Sort(ByStartDateTime(results))
	return results, nil
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
