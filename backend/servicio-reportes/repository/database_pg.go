package repository

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"servicio-reportes/models"

	_ "github.com/lib/pq"
)

var Client *sql.DB

func InitDB() {
	var err error
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	for i := 1; i <= 5; i++ {
		Client, err = sql.Open("postgres", dsn)
		if err == nil {
			err = Client.Ping()
			if err == nil {
				fmt.Println("Conexion exitosa a DB_Reportes")
				break
			}
		}
		fmt.Printf("Esperando a la base de datos... (Intento %d/5)\n", i)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		fmt.Println("No se pudo conectar a la base de datos:", err)
		os.Exit(1)
	}

	query := `
	CREATE TABLE IF NOT EXISTS lecturas_sensores (
		trace_id VARCHAR(36) PRIMARY KEY,
		id_sensor VARCHAR(50) NOT NULL,
		latitud DOUBLE PRECISION NOT NULL,
		longitud DOUBLE PRECISION NOT NULL,
		magnitud DOUBLE PRECISION NOT NULL,
		profundidad_km INTEGER NOT NULL,
		confianza DOUBLE PRECISION NOT NULL,
		estado VARCHAR(20) DEFAULT 'RECIBIDO',
		timestamp TIMESTAMP NOT NULL
	);`

	_, err = Client.Exec(query)
	if err != nil {
		fmt.Println("Error al crear la tabla:", err)
	}
}

func GuardarSismo(traceID string, data models.SensorReport) error {

	query := `INSERT INTO lecturas_sensores 
	          (trace_id, id_sensor, latitud, longitud, magnitud, profundidad_km, confianza, timestamp) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := Client.Exec(query,
		traceID,
		data.IDSensor,
		data.Ubicacion.Lat,
		data.Ubicacion.Lon,
		data.Magnitud,
		data.ProfundidadKm,
		data.Confianza,
		data.Timestamp)

	return err
}
