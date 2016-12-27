package lm

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/PuloV/ics-golang"
	log "github.com/Sirupsen/logrus"
	"github.com/jsgoecke/go-wit"
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

// Gymclass describes a class at Les Mills
type GymClass struct {
	Gym            string    `json:"gym" db:"gym"`
	Name           string    `json:"name" db:"class"`
	Location       string    `json:"location" db:"location"`
	StartDateTime  time.Time `json:"startdatetime" db:"start_datetime"`
	EndDateTime    time.Time `json:"enddatetime" db:"end_datetime"`
	InsertDateTime time.Time `json:"insertdatetime" db:"insert_datetime"`
}

// GymQuery describes a query for GymClasses
type GymQuery struct {
	Gym    Gym
	Class  string
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
        CREATE TABLE IF NOT EXISTS timetable (
           gym VARCHAR(9) NOT NULL,
           class VARCHAR(45) NOT NULL,
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
	CREATE UNIQUE INDEX IF NOT EXISTS unique_class ON timetable(gym, location, start_datetime);`

	_, err = dbConfig.DB.Exec(indexTable)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to index table")
		return err
	}

	// Save all classes
	for _, class := range classes {
		// Prepare insert query
		stmt, err := dbConfig.DB.Prepare("INSERT OR IGNORE INTO timetable (gym, class, location, start_datetime, end_datetime, insert_datetime) values(?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to create row")
		}
		_, err = stmt.Exec(class.Gym, class.Name, class.Location, class.StartDateTime, class.EndDateTime, time.Now())
		if err != nil {
			log.WithFields(log.Fields{"error": err, "row": class}).Error("Failed to insert row")
		}

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

		gymQuery.Limit = "-1"
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
	stmt, err = dbConfig.DB.Prepare("SELECT * FROM timetable WHERE gym LIKE ? AND class LIKE ? and start_datetime > ? and start_datetime < ? limit ?")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to prepare select statement")
		return []GymClass{}, err
	}

	// TODO: Refactor this
	likeGym := "%" + query.Gym.Name + "%"
	likeName := "%" + query.Class + "%"
	rows, err := stmt.Query(
		strings.ToLower(likeGym),
		strings.ToLower(likeName),
		query.After,
		query.Before,
		query.Limit,
	)
	log.Infof("Executing query with args gym: %s class: %s, start_datetime: %s, end_datetime: %s limit %s", likeGym, likeName, query.After, query.Before, query.Limit)
	results := make([]GymClass, 0)
	for rows.Next() {
		var result GymClass
		err := rows.Scan(&result.Gym, &result.Name, &result.Location, &result.StartDateTime, &result.EndDateTime, &result.InsertDateTime)
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
