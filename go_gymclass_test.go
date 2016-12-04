package lm

import (
	"database/sql"
	"testing"
)

func TestNewConfigDefault(t *testing.T) {
	expected := Config{"sqlite3", "admin", "password", "./gym.db", nil}
	db, _ := sql.Open(expected.DBDriver, expected.DBPath)
	expected.DB = db
	actual, err := NewConfig()
	if err != nil {
		t.Errorf("Got an error while creating config: %s", err)
	}
	if expected.DB != actual.DB {
		t.Errorf("Failed to create config expected: %s but got: %s", expected.DB, actual.DB)
	}

}
