package main

import "database/sql"

func InsertMeter(db *sql.DB, meter MeterData) error {
	createSQL := `
	create table if not exists meters (
    	timestamp TEXT NOT NULL PRIMARY KEY,
    	temperature REAL NOT NULL,
    	humidity INTEGER NOT NULL,
    	battery INTEGER NOT NULL
	);
	`
	_, err := db.Exec(createSQL)
	if err != nil {
		return err
	}

	// insert
	insertSQL := `
	insert or replace into meters 
	       (timestamp, temperature, humidity, battery) values (?, ?, ?, ?);
	`
	_, err = db.Exec(insertSQL,
		meter.Timestamp.Format("2006-01-02 15:04:05"),
		meter.Temperature, meter.Humidity, meter.Battery)
	if err != nil {
		return err
	}
	return nil
}
