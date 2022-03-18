# L.E.R.T. 
(stands for Live Exchange Rates Today)

The application is designed to monitor exchange rates.

## Requirements:
* go 1.16+ installed
* gcc (to build sqlite3 driver)

## Setup and run locally

### Edit .env according to your needs
```
cp .env.sample .env
```

### Init database schema
```
go run init.go
```

### Run REST API server
```
go run main.go
```

### Visit REST API endpoint
http://localhost:9000/api/rate

where 9000 - port you set up in .env
