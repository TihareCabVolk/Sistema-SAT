package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type SafeDB struct {
	DB *sql.DB
}

func InitDB(dsn string) (*SafeDB, error) {
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
	CREATE TABLE IF NOT EXISTS alertas (
		id VARCHAR(36) PRIMARY KEY,
		id_validacion VARCHAR(36) NOT NULL,
		magnitud DOUBLE PRECISION NOT NULL,
		epicentro_lat DOUBLE PRECISION NOT NULL,
		epicentro_lon DOUBLE PRECISION NOT NULL,
		nivel_alerta VARCHAR(20) NOT NULL,
		costo_emergencia DOUBLE PRECISION NOT NULL,
		zonas_afectadas JSONB DEFAULT '[]',
		estado VARCHAR(20) DEFAULT 'EMITIDA',
		creado_en TIMESTAMP NOT NULL
	);`

	if _, err = db.Exec(query); err != nil {
		return nil, err
	}

	return &SafeDB{DB: db}, nil
}

func (s *SafeDB) Close() error {
	return s.DB.Close()
}
