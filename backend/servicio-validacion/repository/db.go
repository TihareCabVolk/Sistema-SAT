package repository

import (
	"database/sql"
	"fmt"
	"time"
)

func InitDB(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 1; i <= 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		fmt.Printf("Esperando a la base de datos... (Intento %d/5)\n", i)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS senales (
		id VARCHAR(36) PRIMARY KEY,
		id_sensor VARCHAR(50) NOT NULL,
		lat DOUBLE PRECISION NOT NULL,
		lon DOUBLE PRECISION NOT NULL,
		magnitud DOUBLE PRECISION NOT NULL,
		profundidad_km INTEGER NOT NULL,
		confianza DOUBLE PRECISION NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		validada BOOLEAN DEFAULT false
	);`

	if _, err = db.Exec(query); err != nil {
		return nil, err
	}

	return db, nil
}
