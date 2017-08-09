# Currency Ticker UI
This project is an extension of the [kyco.bitcoin.currency.tickers](https://github.com/kyco/kyco.bitcoin.currency.tickers.ui) project. This is a front end to display the currency ticker updates directly in your terminal.



## How To Run?
#### Dependencies
To install the required dependencies to compile your own binaries.
```
go get -u github.com/buger/goterm
go get -u github.com/fsnotify/fsnotify
go get -u github.com/gizak/termui
go get -u github.com/mattn/go-sqlite3
go get -u github.com/op/go-logging
go get -u github.com/spf13/viper
```

or

```
glide install
```

Build

```
go build main.go
./main
```

### How it works.
The sqlite database created by the [kyco.bitcoin.currency.tickers](https://github.com/kyco/kyco.bitcoin.currency.tickers.ui) is queried every 20 seconds for the latest BTC conversion rates.

### What Works
Querying the sqlite database and displaying these in your terminal

### Config File
Copy the config file to ```~/.config/bitcoin-ui/```.

### Future / TODO
 - Add graphs into the terminal as well
 - Move to a percentage based calculation in terms of where to place ui elements on the terminal so that it survives resizing