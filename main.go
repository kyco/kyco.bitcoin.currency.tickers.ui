package main

import (
	"database/sql"
	// "fmt"
	tm "github.com/buger/goterm"
	"github.com/fsnotify/fsnotify"
	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
	"time"
)

var err error

// Configure main config struct
type Config struct {
	Interval int
	Log      string
	SQLite   string
}

// Exchange data
type ExchangeData struct {
	Exchange     string
	Price        float64
	Ask          float64
	Bid          float64
	Volume       float64
	Date         string
	CurrencyCode string
}

var config Config
var logFile *os.File
var log = logging.MustGetLogger("bitcoin-ui")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// Setup global home and config folders
// File location
var fileLocation string = "/.config/bitcoin-ui/"
var bitoinStatsLocation string = "/.config/bitcoin-stats/"

// Get home folder
var home string = os.Getenv("HOME") + fileLocation
var sqliteDBLocation string = os.Getenv("HOME") + bitoinStatsLocation

// manages the UI
func bitcoin_prices() {

	// Tick on the minute
	// t := minuteTicker()

	for {

		tm.Clear() // Clear the screen
		// wait for the tick
		// <-t.C

		// // Grab SQLite data
		// 1 hourx
		hour := get_last_exchange_rates("1")
		day := get_last_exchange_rates("24")

		begin := 3
		// Underline
		underline := strings.Repeat("-", 99)

		// By moving cursor to top-left position we ensure that console output
		// will be overwritten each time, instead of adding new.
		tm.MoveCursor(begin, 40)

		tm.Println("Current Time:", time.Now().Format(time.RFC1123))

		tm.MoveCursor(begin+3, 60)
		tm.Println(tm.Bold("1 Hour"))
		tm.MoveCursorForward(10)
		tm.Println(underline)

		tm.MoveCursor(begin+5, 11)
		tm.Printf(tm.Bold("Exchange"))
		tm.MoveCursorForward(10)
		tm.Printf(tm.Bold("Average"))
		tm.MoveCursorForward(10)
		tm.Printf(tm.Bold("Ask"))
		tm.MoveCursorForward(13)
		tm.Printf(tm.Bold("Buy"))
		tm.MoveCursorForward(12)
		tm.Printf(tm.Bold("Volume"))
		tm.MoveCursorForward(8)
		tm.Printf(tm.Bold("Time"))

		// Reset to the beginning of the line
		tm.MoveCursor(begin+6, 10)

		// print hour based data
		for i := range hour {

			tm.Println()

			tm.MoveCursorForward(10)
			tm.Printf(hour[i].Exchange)
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((18 - len(hour[i].Exchange)))
			tm.Printf(strconv.FormatFloat(hour[i].Price, 'f', 2, 64))
			tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(hour[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((13 - len(strconv.FormatFloat(hour[i].Price, 'f', 2, 64))))
			tm.Printf(strconv.FormatFloat(hour[i].Ask, 'f', 2, 64))
			tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(hour[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((12 - len(strconv.FormatFloat(hour[i].Ask, 'f', 2, 64))))
			tm.Printf(strconv.FormatFloat(hour[i].Bid, 'f', 2, 64))
			tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(hour[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((11 - len(strconv.FormatFloat(hour[i].Bid, 'f', 2, 64))))
			tm.Printf(strconv.FormatFloat(hour[i].Volume, 'f', 2, 64))
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((14 - len(strconv.FormatFloat(hour[i].Volume, 'f', 2, 64))))
			tm.Printf(hour[i].Date)
		}

		// Reset row
		tm.MoveCursor(10+len(hour), 1)

		tm.MoveCursorForward(60)
		tm.MoveCursorDown(2)
		tm.Println(tm.Bold("24 Hour"))
		tm.MoveCursorForward(10)
		tm.Println(underline)

		tm.MoveCursorForward(10)
		tm.Printf(tm.Bold("Exchange"))
		tm.MoveCursorForward(10)
		tm.Printf(tm.Bold("Average"))
		tm.MoveCursorForward(10)
		tm.Printf(tm.Bold("Ask"))
		tm.MoveCursorForward(13)
		tm.Printf(tm.Bold("Buy"))
		tm.MoveCursorForward(12)
		tm.Printf(tm.Bold("Volume"))
		tm.MoveCursorForward(8)
		tm.Printf(tm.Bold("Time"))

		// Reset to the beginning of the line
		tm.MoveCursorDown(1)

		// print daily based data
		for i := range day {

			tm.Println()

			tm.MoveCursorForward(10)
			tm.Printf(day[i].Exchange)
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((18 - len(day[i].Exchange)))
			tm.Printf(strconv.FormatFloat(day[i].Price, 'f', 2, 64))
			tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(day[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((13 - len(strconv.FormatFloat(day[i].Price, 'f', 2, 64))))
			tm.Printf(strconv.FormatFloat(day[i].Ask, 'f', 2, 64))
			tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(day[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((12 - len(strconv.FormatFloat(day[i].Ask, 'f', 2, 64))))
			tm.Printf(strconv.FormatFloat(day[i].Bid, 'f', 2, 64))
			tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(day[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((11 - len(strconv.FormatFloat(day[i].Bid, 'f', 2, 64))))
			tm.Printf(strconv.FormatFloat(day[i].Volume, 'f', 2, 64))
			// tm.MoveCursorDown(1)
			tm.MoveCursorForward((14 - len(strconv.FormatFloat(day[i].Volume, 'f', 2, 64))))
			tm.Printf(day[i].Date)
		}

		tm.Flush() // Call it every time at the end of rendering

		time.Sleep(20 * time.Second)

	}

}

// Grab sqlite data from database
func get_last_exchange_rates(hours string) []ExchangeData {

	response := []ExchangeData{}

	// Open SQLite
	db := sqlite_open()
	result, err := db.Query(`select exchange, printf("%.2f", AVG(ask)) as ask, printf("%.2f", AVG(bid)) as bid,
				printf("%.2f", AVG((ask + bid) / 2)) as price,
				AVG(volume) as volume, datetime(timestamp, 'unixepoch') as timestamp, currencyCode
				from exchanges
				where (currencyCode = "USD" or currencyCode = "btcusd" or currencyCode = "btc_usd")
					and datetime(timestamp, 'unixepoch') >= datetime('now', '-` + hours + ` Hour')
				group by exchange
				order by volume desc;`)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	// Loop through records
	for result.Next() {
		// Scan records into a struct
		tmpStruct := ExchangeData{}
		result.Scan(&tmpStruct.Exchange, &tmpStruct.Ask, &tmpStruct.Bid, &tmpStruct.Price, &tmpStruct.Volume,
			&tmpStruct.Date, &tmpStruct.CurrencyCode)

		response = append(response, tmpStruct)

	}

	return response

}

// Open SQlite Connection
func sqlite_open() *sql.DB {
	db, err := sql.Open("sqlite3", sqliteDBLocation+"data.db")
	if err != nil {
		log.Error(err.Error())
	}
	return db
}

// Waits for the minute to tick over
func minuteTicker() *time.Ticker {
	// Return new ticker that triggers on the minute
	return time.NewTicker(time.Second * time.Duration(int(60*config.Interval)-time.Now().Second()))
}

// Configure logging
func config_log() {

	// Check if it is already open
	logFile.Close()

	// Configure logging
	logFile, err := os.OpenFile(config.Log, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Info("error opening file: %v", err)
	}

	// For demo purposes, create two backend for os.Stderr.
	loggingFile := logging.NewLogBackend(logFile, "", 0)

	// For messages written to loggingFile we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	loggingFileFormatter := logging.NewBackendFormatter(loggingFile, format)

	// Set the backends to be used.
	logging.SetBackend(loggingFileFormatter)
}

// Configure configs
func config_init() {

	// Config File
	viper.SetConfigName("config") // no need to include file extension
	viper.AddConfigPath(home)     // set the path of your config file

	err := viper.ReadInConfig()
	if err != nil {
		log.Info("Config file not found... Error %s\n", err)
	} else {

		// ========= CONFIG ================================================================
		interval := viper.GetInt("interval")
		logLocation := viper.GetString("log")
		sqliteLocation := viper.GetString("sqliteLocation")
		// Main Config
		config = Config{
			Interval: interval,
			Log:      logLocation,
			SQLite:   sqliteLocation,
		}
	}

	// Monitor the config file for changes and reload
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {

		// Re-configure config
		config_init()

		// Re-configure logging
		config_log()

		log.Info("Config file changed:", e.Name)
	})
}

func main() {
	// Initialise config file and settings
	config_init()

	// Configure logging
	config_log()

	// don't forget to close the log file
	defer logFile.Close()

	// Start UI process
	go bitcoin_prices()

	// Notify log that we are up and running
	log.Info("started Bitcoin UI")

	select {}
}
