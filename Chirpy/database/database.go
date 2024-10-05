package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"sync"
)

type DB struct {
	path string
	mux  *sync.Mutex
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	fileContents, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("Creating database")
			err := os.WriteFile("database.json", []byte(`{
	"chirps": {
		"1": {
			"id": 1,
			"body": "This is the first chirp ever!"
		},
		"2": {
			"id": 2,
			"body": "Hello, world!"
		}
	}
}`), 0666)
			if err != nil {
				fmt.Println("Error writing to file")
			}
		} else {
			fmt.Println("uh oh")
			return nil, err
		}
	}
	fmt.Println(fileContents)
	mux := &sync.Mutex{}
	db := DB{
		path: path,
		mux:  mux,
	}
	return &db, nil
}

// func (db *DB) CreateChirp(body string) (Chirp, error) {

// }

func (db *DB) GetChirps() (DBStructure, error) {
	db.mux.Lock()
	contents, err := os.ReadFile(db.path)
	if err != nil {
		fmt.Println(err)
	}
	db.mux.Unlock()
	// Temporary structure to unmarshal into with string keys for chirps
	var temp struct {
		Chirps map[string]Chirp `json:"chirps"`
	}

	// Unmarshal the JSON data into the temp structure
	if err := json.Unmarshal(contents, &temp); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	// Initialize the final DBStructure with int keys
	c := DBStructure{
		Chirps: make(map[int]Chirp),
	}

	// Convert string keys to int keys
	for key, chirp := range temp.Chirps {
		intKey, err := strconv.Atoi(key)
		if err != nil {
			fmt.Println("Error converting key to int:", err)
		}
		c.Chirps[intKey] = chirp
	}

	return c, nil
}

// func (db *DB) ensureDB() error {

// }

// func (db *DB) loadDB() (DBStructure, error) {

// }

// func (db *DB) writeDB(dbStructure DBStructure) error {

// }
