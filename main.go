package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//Written as a quick response to the realization that a mission-critical information source would be lost.
//Resulted in ~250MB of useful data. This script doesn't work anymore against the original target, but might be useful with some adaptions in the future.

//SQLUsername is the SQL username
var SQLUsername = "sourceoftruth"

//SQLPassword is the SQL password
var SQLPassword = "sourceoftruth"

//SQLDatabase is the SQL database
var SQLDatabase = "sourceoftruth"

type Entry struct {
	MagicAlpha   string `json:"MagicAlpha"`
	IP           string `json:"IP"`
	MagicBravo   string `json:"MagicBravo"`
	MAC          string `json:"MAC"`
	MagicCharlie string `json:"MagicCharlie"`
	Port         string `json:"Port"`
	MagicDelta   string `json:"MagicDelta"`
	VLAN         string `json:"VLAN"`
}

func compileSQLPassword() string {
	return SQLUsername + ":" + SQLPassword + "@/" + SQLDatabase
}

func establishDB() *sql.DB {
	db, err := sql.Open("mysql", compileSQLPassword()) //Open a database connection
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func readURLs() []string {
	file, err := ioutil.ReadFile("URLLIST.txt") //Previously enumerated get URLs. Best to not make iteration easy.
	if err != nil {
		log.Fatal(err)
	}
	return strings.Fields(string(file))
}

func getData(url string) []byte { //Use a HTTP get to request JSON data
	httpClient := http.Client{
		Timeout: time.Second * 120,
	}
	request, errreq := http.NewRequest(http.MethodGet, url, nil)
	if errreq != nil {
		log.Fatal(errreq)
	}
	response, geterr := httpClient.Do(request)
	if geterr != nil {
		log.Fatal(geterr)
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	text, texterr := ioutil.ReadAll(response.Body)
	if texterr != nil {
		log.Fatal(texterr)
	}
	fmt.Println(url)
	return text
}

func fetchURL(url string, data *[]Entry) {
	text := getData(url)
	jsonerr := json.Unmarshal(text, &data)
	if jsonerr != nil {
		log.Fatal(jsonerr)
	}
}

func IC(input *string) *string {
	if *input == "" {
		input = nil //Nil out empty return values
	}
	return input
}

func insertIntoDB(indata *[]Entry, db *sql.DB) {
	data := *indata
	for i := range data {
		fmt.Println(data[i].IP)
		db.Exec("INSERT INTO sourceoftruth VALUES (?,?,?,?,?,?,?,?)", *IC(&data[i].MagicAlpha), *IC(&data[i].IP), *IC(&data[i].MagicBravo), *IC(&data[i].MAC), *IC(&data[i].MagicCharlie), *IC(&data[i].Port), *IC(&data[i].MagicDelta), *IC(&data[i].VLAN))
	}
}

func saveURLs(urls []string, db *sql.DB) {
	for i := range urls {
		if urls[i] == "" {
			fmt.Println("BREAK ON " + urls[i])
			break
		}
		fmt.Println(urls[i])
		data := []Entry{}
		fetchURL(urls[i], &data)
		insertIntoDB(&data, db)
	}
}

func main() {
	db := establishDB()
	defer db.Close()
	urls := readURLs()
	saveURLs(urls, db)
}
