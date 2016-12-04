package lm

import (
	"database/sql"
	"sort"
	"strings"
	"time"

	"github.com/PuloV/ics-golang"
	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
)

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

// GymClass describes a class at Les Mills
type GymClass struct {
	Gym           string    `json:"gym" db:"gym"`
	Name          string    `json:"name" db:"class"`
	Location      string    `json:"location" db:"location"`
	StartDateTime time.Time `json:"startdatetime" db:"start_datetime"`
	EndDateTime   time.Time `json:"enddatetime" db:"end_datetime"`
}

// GymQuery describes a query for GymClasses
type GymQuery struct {
	Gym    Gym
	Class  GymClass
	Before time.Time
	After  time.Time
	Limit  string
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

// GetClasses will return a list of classes as stored by LesMills for the next 7 days when passing one or more Gyms
func GetClasses(gyms []Gym) ([]GymClass, error) {

	baseURL := "https://www.lesmills.co.nz/timetable-calander.ashx?club="

	var foundClasses = []GymClass{}

	parser := ics.New()
	inputChan := parser.GetInputChan()

	for _, gym := range gyms {
		// Create the URL for the ICS based on the gym
		inputChan <- baseURL + gym.ID
		parser.Wait()

		cal, err := parser.GetCalendars()
		if err != nil {
			log.WithFields(log.Fields{"value": err}).Error("Failed to get calendars")
			return nil, err
		}
		var foundClass GymClass
		for _, c := range cal {

			loc, err := time.LoadLocation("Pacific/Auckland")
			if err != nil {
				log.WithFields(log.Fields{"value": err}).Error("Failed to get timezone")
				return nil, err
			}
			c.SetTimezone(*loc)
			for _, event := range c.GetEvents() {
				name := event.GetSummary()
				translateName(&name)
				foundClass = GymClass{
					Gym:           gym.Name,
					Name:          name,
					Location:      event.GetLocation(),
					StartDateTime: event.GetStart(),
					EndDateTime:   event.GetEnd(),
				}
				foundClasses = append(foundClasses, foundClass)
			}
		}
	}
	sort.Sort(ByStartDateTime(foundClasses))
	return foundClasses, nil
}

// StoreClasses will store a list of classes into a database based on the configuration provided
func StoreClasses(classes []GymClass, dbConfig *Config) error {

	// Create table
	createTable := `
        CREATE TABLE IF NOT EXISTS timetable (
           gym VARCHAR(9) NOT NULL,
           class VARCHAR(45) NOT NULL,
           location VARCHAR(27) NOT NULL,
           start_datetime DATETIME NOT NULL,
           end_datetime DATETIME NOT NULL);`

	_, err := dbConfig.DB.Exec(createTable)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create table")
		return err
	}

	// Create index on table
	indexTable := `
	CREATE UNIQUE INDEX IF NOT EXISTS unique_class ON timetable(gym, location, start_datetime);`

	_, err = dbConfig.DB.Exec(indexTable)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to index table")
		return err
	}

	// Save all classes
	for _, class := range classes {
		// Prepare insert query
		stmt, err := dbConfig.DB.Prepare("INSERT OR IGNORE INTO timetable (gym, class, location, start_datetime, end_datetime) values(?, ?, ?, ?, ?)")
		if err != nil {
			log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to create row")
		}
		_, err = stmt.Exec(class.Gym, class.Name, class.Location, class.StartDateTime, class.EndDateTime)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to insert row")
		}

	}
	return nil
}

// QueryClasses will query the classes from the stored database and return the results
func QueryClasses(query GymQuery, dbConfig *Config) ([]GymClass, error) {
	// Prepare the SELECT query
	var err error
	var stmt *sql.Stmt
	stmt, err = dbConfig.DB.Prepare("SELECT * FROM timetable WHERE gym LIKE ? AND class LIKE ? and start_datetime > ? and start_datetime < ? limit ?")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to prepare select statement")
		return []GymClass{}, err
	}

	// TODO: Refactor this
	likeGym := "%" + query.Gym.Name + "%"
	likeName := "%" + query.Class.Name + "%"
	rows, err := stmt.Query(
		strings.ToLower(likeGym),
		strings.ToLower(likeName),
		query.After,
		query.Before,
		query.Limit,
	)

	var results []GymClass
	for rows.Next() {
		var result GymClass
		err := rows.Scan(&result)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to marshal result")
		}
		translateName(&result.Name)
		results = append(results, result)
	}
	sort.Sort(ByStartDateTime(results))
	return results, nil
}
