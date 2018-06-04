package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
)

// JokesCache Init
var JokesCache = cache.New(cache.NoExpiration, cache.NoExpiration)

// CacheIndex Init
var CacheIndex = cache.New(-1, -1)

var fileName = "jokes.json"

// Joke struct holds the jSon structure
type Joke struct {
	Type  string
	Value struct {
		ID         int
		Joke       string
		Categories []string
	}
}

// Last downloaded joke ID
func lastJoke() int {
	var lastJoke int
	lastJokeCache, found := CacheIndex.Get("lastJoke")
	if found {
		lastJoke = lastJokeCache.(int)
	}
	CacheIndex.Set("lastJoke", lastJoke+1, cache.NoExpiration) // Updates the Joke
	// fmt.Println("JokeID is: ", lastJoke)
	return lastJoke
}

// Downloads the Joke from the API
func getJoke() Joke {
	lastJokeID := lastJoke() // Last joke id
	var m Joke
	// Downloading joke from the API
	url := "http://api.icndb.com/jokes/" + strconv.Itoa(lastJokeID)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	defer response.Body.Close()
	// Reading the response
	contents, err := ioutil.ReadAll(response.Body)
	check(err)
	// From byte to struct
	json.Unmarshal(contents, &m)
	// If the API return value is success we save the Joke in the API
	if m.Type == "success" {
		jsonString, err := json.Marshal(m.Value)
		if err == nil {
			if x, found := JokesCache.Get("cachedJokes"); found {
				cachedJokes := x.(string)
				JokesCache.Set("cachedJokes", cachedJokes+string(jsonString)+",\n", cache.DefaultExpiration)
				fmt.Println("Joke into cache: ", string(jsonString))
			} else {
				panic("There is a problem with the cache (cachedJokes)")
			}
		} else {
			panic(err)
		}
	} else {
		fmt.Println("API error: ", m)
	}
	return m
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Takes the jokes from the cache system and writes it to a file
func writeJokesToFile() {
	fmt.Println("Updating file and cleaning cache")
	if str, found := JokesCache.Get("cachedJokes"); found {
		if _, err := os.Stat(fileName); os.IsNotExist(err) { // Creating the file if not exist
			file, err := os.Create(fileName)
			check(err)
			file.Close()
		}
		fileJokes, err := ioutil.ReadFile(fileName) // Opening the file
		check(err)
		// Formating the string properly
		fileJokesString := strings.TrimRight(string(fileJokes), "]")
		fileJokesString = strings.TrimLeft(fileJokesString, "[")
		writeStr := fileJokesString + "," + str.(string)
		writeStr = strings.TrimRight(writeStr, ",\n")
		writeStr = strings.TrimLeft(writeStr, ",")
		writeStr = "[" + writeStr + "]"
		// Writing the jSon to the file
		err = ioutil.WriteFile(fileName, []byte(writeStr), 0666)
		check(err)
		// Cleaning the cache
		JokesCache.Set("cachedJokes", "", cache.NoExpiration)
	} else {
		panic("There is a problem with the cache (cachedJokes)")
	}
}

// The program keeps executing forever
func blockForever() {
	select {}
}

func main() {
	// Removes old file if exist
	os.Remove(fileName)
	// Inits cache into empty string
	JokesCache.Set("cachedJokes", "", cache.DefaultExpiration)

	// Ticker for the getJoke func executed as goroutine
	getJokeTicker := time.NewTicker(3 * time.Second)
	go func() {
		for range getJokeTicker.C {
			getJoke()
		}
	}()

	// Ticker for the writeJokesToFile func executed as goroutine
	writeJokeTicker := time.NewTicker(60 * time.Second)
	go func() {
		for range writeJokeTicker.C {
			writeJokesToFile()
		}
	}()
	// Execute progam "forever"
	blockForever()
}
