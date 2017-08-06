package main

import (
	"database/sql"
	tm "github.com/buger/goterm"
	"github.com/fsnotify/fsnotify"
	ui "github.com/gizak/termui"
	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"
)

var err error

// Configure main config struct
type Config struct {
	Interval           int
	Log                string
	SQLite             string
	CurrencyCodeIgnore string
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

		// wait for the tick
		// <-t.C

		var data []ExchangeData

		// Grab SQLite data
		data = get_last_exchange_rates("15")

		// Pop graph in
		// ui_create_graph(data)

		// Update UI
		ui_top_table(data)

		time.Sleep(20 * time.Second)

	}

}

// manage the UI
func ui_top_table(data []ExchangeData) {

	tm.Clear() // Clear the screen

	// Underline
	underline := strings.Repeat("-", 99)

	// By moving cursor to top-left position we ensure that console output
	// will be overwritten each time, instead of adding new.
	tm.MoveCursor(3, 20)

	tm.Println("Current Time:", time.Now().Format(time.RFC1123))

	tm.MoveCursor(6, 40)
	tm.Println(tm.Bold("Current Exchange Rate"))
	tm.MoveCursorForward(5)
	tm.Println(underline)

	tm.MoveCursor(8, 6)
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
	tm.MoveCursor(9, 10)

	// print hour based data
	for i := range data {

		tm.Println()

		tm.MoveCursorForward(5)
		tm.Printf(data[i].Exchange)
		// tm.MoveCursorDown(1)
		tm.MoveCursorForward((18 - len(data[i].Exchange)))
		tm.Printf(strconv.FormatFloat(data[i].Price, 'f', 2, 64))
		tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(data[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
		// tm.MoveCursorDown(1)
		tm.MoveCursorForward((13 - len(strconv.FormatFloat(data[i].Price, 'f', 2, 64))))
		tm.Printf(strconv.FormatFloat(data[i].Ask, 'f', 2, 64))
		tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(data[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
		// tm.MoveCursorDown(1)
		tm.MoveCursorForward((12 - len(strconv.FormatFloat(data[i].Ask, 'f', 2, 64))))
		tm.Printf(strconv.FormatFloat(data[i].Bid, 'f', 2, 64))
		tm.Printf(" " + strings.ToUpper(strings.Replace(strings.Replace(data[i].CurrencyCode, "btc", "", -1), "_", "", -1)))
		// tm.MoveCursorDown(1)
		tm.MoveCursorForward((11 - len(strconv.FormatFloat(data[i].Bid, 'f', 2, 64))))
		tm.Printf(strconv.FormatFloat(data[i].Volume, 'f', 2, 64))
		// tm.MoveCursorDown(1)
		tm.MoveCursorForward((14 - len(strconv.FormatFloat(data[i].Volume, 'f', 2, 64))))
		tm.Printf(data[i].Date)
	}

	tm.Flush() // Call it every time at the end of rendering
}

// Create a graph
func ui_create_graph(data []ExchangeData) {
	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	sinps := (func() []float64 {
		n := 220
		ps := make([]float64, n)
		for i := range ps {
			ps[i] = 1 + math.Sin(float64(i)/5)
		}
		return ps
	})()

	lc := ui.NewLineChart()
	lc.BorderLabel = "Bitcoin Price USD"
	lc.Data = sinps
	lc.Width = 100
	lc.Height = 11
	lc.X = 0
	lc.Y = 20
	lc.AxesColor = ui.ColorWhite
	lc.LineColor = ui.ColorRed | ui.AttrBold
	// lc.Mode = "dot"

	draw := func(t int) {
		lc.Data = sinps[t/2%220:]
		ui.Render(lc)
	}
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/timer/1s", func(e ui.Event) {
		t := e.Data.(ui.EvtTimer)
		draw(int(t.Count))
	})
	ui.Loop()
}

// Grab sqlite data from database
func get_last_exchange_rates(minutes string) []ExchangeData {

	response := []ExchangeData{}

	// Open SQLite
	db := sqlite_open()
	result, err := db.Query(`select exchange, printf("%.2f", AVG(ask)) as ask, printf("%.2f", AVG(bid)) as bid,
				printf("%.2f", AVG((ask + bid) / 2)) as price,
				AVG(volume) as volume, datetime(timestamp, 'unixepoch') as timestamp, currencyCode
				from exchanges
				where currencyCode in ("USD", "ZAR") and datetime(timestamp, 'unixepoch') >= datetime('now', '-` + minutes + ` Minute')
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
		currencyCodeIgnore := viper.GetString("currencyCodeIgnore")
		// Main Config
		config = Config{
			Interval:           interval,
			Log:                logLocation,
			SQLite:             sqliteLocation,
			CurrencyCodeIgnore: currencyCodeIgnore,
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

	http.ListenAndServe(":8081", http.DefaultServeMux)
}
