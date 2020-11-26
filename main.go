package main

import (
	"flag"
	"fmt"
	"os"

	"database/sql"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

// Stats : structure to hold our MySQL statistics values
type Stats struct {
	Queries int64 `json:"queries"`
	Selects int64 `json:"selects"`
	Inserts int64 `json:"inserts"`
	Updates int64 `json:"updates"`
	Deletes int64 `json:"deletes"`
	Uptime  int64 `json:"uptime"`
	Totals  int64 `json:"totals"`
}

func main() {
	// get command line flags

	var myPath string
	var icingaOut string
	var exitCode int

	myVersion := "0.0.2"

	hostAddress := flag.String("h", "127.0.0.1", "Host ")
	hostPort := flag.String("p", "3306", "Target port")
	userName := flag.String("u", "", "MySQL user")
	passWord := flag.String("P", "", "MySQL password for user")
	timeOut := flag.Int("t", 2, "MySQL connection timeout in seconds")
	version := flag.Bool("v", false, "Display version information")
	rWarn := flag.Int64("rwarn", 1000, "Warning level read operations")
	rCrit := flag.Int64("rcrit", 1500, "Critical level read operations")
	wWarn := flag.Int64("wwarn", 50, "Warning level write operations")
	wCrit := flag.Int64("wcrit", 100, "Critical level write operations")
	outPath := flag.String("o", "", "directory for temporary stats files")

	flag.Parse()

	if *version {
		fmt.Println(fmt.Sprintf("%s v%s", os.Args[0], myVersion))
		os.Exit(0)
	}

	if *outPath != "" {
		myPath = *outPath
	} else {
		myPath, _ = os.Getwd()
	}

	myName, err := os.Executable()

	// Get stats from last run
	lastStats, _ := getOldStats(myPath + "/" + filepath.Base(myName) + "." + *hostAddress + ".stats")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=%ds", *userName, *passWord, *hostAddress, *hostPort, *timeOut))
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// var result string
	var stats Stats

	// Get all the results in one run
	rows, err := db.Query("SELECT VARIABLE_NAME, VARIABLE_VALUE FROM information_schema.global_status WHERE VARIABLE_NAME IN ('QUESTIONS', 'COM_DELETE', 'COM_INSERT', 'COM_UPDATE', 'COM_SELECT', 'UPTIME')")
	if err != nil {
		panic(err.Error())
	}

	for rows.Next() {
		var key string
		var val int64

		if err :=  rows.Scan(&key, &val); err != nil {
			panic(err.Error())
		}
		switch key {
			case "QUESTIONS":
				stats.Queries = val
			case "COM_DELETE":
				stats.Deletes = val
			case "COM_INSERT":
				stats.Inserts = val
			case "COM_UPDATE":
				stats.Updates = val
			case "COM_SELECT":
				stats.Selects = val
			case "UPTIME":
				stats.Uptime = val
			default:
				continue
		}
	}

	stats.Totals = stats.Selects + stats.Deletes + stats.Inserts + stats.Updates

	// Check for MySQL restart - reset counters if there was
	if lastStats.Uptime > stats.Uptime {
		lastStats.Uptime = 0
		lastStats.Deletes = 0
		lastStats.Inserts = 0
		lastStats.Queries = 0
		lastStats.Selects = 0
		lastStats.Totals = 0
		lastStats.Updates = 0
	}

	// write the data to the SQLite store
	err = writeNewStats(myPath+"/"+filepath.Base(myName)+"."+*hostAddress+".stats", stats)

	diffTime := stats.Uptime - lastStats.Uptime

	curRps := (stats.Selects - lastStats.Selects) / diffTime
	curWps := (stats.Deletes + stats.Updates + stats.Inserts - lastStats.Deletes - lastStats.Updates - lastStats.Inserts) / diffTime

	icingaOut = "OK: Normal level of reads and writes"
	exitCode = 0
	if curRps > *rCrit {
		icingaOut = "CRITICAL: Reads above critical level (" + ")"
		exitCode = 2
	} else {
		if curRps > *rWarn {
			icingaOut = "WARNING: Reads above warning level"
			exitCode = 1
		}
	}

	if curWps > *wCrit {
		icingaOut = "CRITICAL: Writes above critical level"
		exitCode = 2
	} else {
		if curWps > *wWarn {
			icingaOut = "WARNING: Writes above warning level"
			exitCode = 1
		}
	}

	icingaOut += " | "
	icingaOut += fmt.Sprintf("'queries'=%d ", stats.Queries)
	icingaOut += fmt.Sprintf("'selects'=%d ", stats.Selects)
	icingaOut += fmt.Sprintf("'inserts'=%d ", stats.Inserts)
	icingaOut += fmt.Sprintf("'updates'=%d ", stats.Updates)
	icingaOut += fmt.Sprintf("'deletes'=%d ", stats.Deletes)
	icingaOut += fmt.Sprintf("'uptime'=%d ", stats.Uptime)
	icingaOut += fmt.Sprintf("'reads_per_second'=%d ", curRps)
	icingaOut += fmt.Sprintf("'writes_per_second'=%d ", curWps)

	fmt.Println(icingaOut)
	os.Exit(exitCode)
}

func writeNewStats(file string, stats Stats) error {
	var content []byte

	content, err := json.Marshal(stats)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = ioutil.WriteFile(file, content, 0644)
	return err
}

func getOldStats(file string) (Stats, error) {
	var stats Stats

	content, _ := ioutil.ReadFile(file)
	err := json.Unmarshal(content, &stats)

	return stats, err
}
