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

// Joke have the content of the jSon coming from the API
type Joke struct {
	Type  string
	Value struct {
		ID         int
		Joke       string
		Categories []string
	}
}

// Gets the last joke from the cache system
func lastJoke() int {
	var lastJoke int
	lastJokeCache, found := CacheIndex.Get("lastJoke")
	if found {
		lastJoke = lastJokeCache.(int)
	}
	CacheIndex.Set("lastJoke", lastJoke+1, cache.NoExpiration) // Updates the Joke
	fmt.Println("JokeID is: ", lastJoke)
	return lastJoke
}

func getJoke() Joke {
	lastJokeID := lastJoke() + 1
	var m Joke
	url := "http://api.icndb.com/jokes/" + strconv.Itoa(lastJokeID)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		json.Unmarshal(contents, &m)
		if m.Type == "success" {
			jsonString, err := json.Marshal(m.Value)
			if err == nil {
				if x, found := JokesCache.Get("cachedJokes"); found {
					cachedJokes := x.(string)
					JokesCache.Set("cachedJokes", cachedJokes+string(jsonString)+",\n", cache.DefaultExpiration)
					// if x, found := JokesCache.Get("cachedJokes"); found {
					// 	fmt.Println("ddd")
					// 	fmt.Println("x is: ", x)
					// } else {
					// 	fmt.Println("found is: ", found)
					// }
					//writeJokesToFile()
				} else {
					fmt.Println("found not found: ", found)
				}
			} else {
				fmt.Println("jSon error: ", err)
			}
		}
	}
	fmt.Println(m)
	return m
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeJokesToFile() {
	fmt.Println("lets going to create the file")
	if str, found := JokesCache.Get("cachedJokes"); found {
		if _, err := os.Stat("jokes.json"); os.IsNotExist(err) { // Creating the file
			fmt.Println("file created")
			file, err := os.Create("jokes.json")
			check(err)
			file.Close()
		}
		fileJokes, err := ioutil.ReadFile("jokes.json")
		check(err)
		fileJokesString := strings.TrimRight(string(fileJokes), "]")
		fileJokesString = strings.TrimLeft(fileJokesString, "[")
		writeStr := fileJokesString + "," + str.(string)
		writeStr = strings.TrimRight(writeStr, ",\n")
		writeStr = strings.TrimLeft(writeStr, ",")
		writeStr = "[" + writeStr + "]"
		err = ioutil.WriteFile("jokes.json", []byte(writeStr), 0666)
		check(err)
		JokesCache.Set("cachedJokes", "", cache.NoExpiration)
	} else {
		fmt.Println("found is: ", found)
	}
}

func blockForever() {
	select {}
}

func main() {
	JokesCache.Set("cachedJokes", "", cache.DefaultExpiration)
	getJokeTicker := time.NewTicker(3 * time.Second)
	go func() {
		for range getJokeTicker.C {
			getJoke()
		}
	}()
	writeJokeTicker := time.NewTicker(60 * time.Second)
	go func() {
		for range writeJokeTicker.C {
			writeJokesToFile()
		}
	}()
	blockForever()
}

