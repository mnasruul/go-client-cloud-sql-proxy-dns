package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Fatal Error: %s environment variable not set.\n", k)
		}
		return v
	}
	ctx := context.Background()

	// Initialize a new Cloud SQL Connector
	connector, err := cloudsqlconn.NewDialer(ctx)
	if err != nil {
		log.Fatalf("Unable to initialize connector: %v", err)
	}
	defer connector.Close()
	var (
		dbUser = mustGetenv("DB_USER")  // e.g. 'my-db-user'
		dbPwd  = mustGetenv("DB_PASS")  // e.g. 'my-db-password'
		dbName = mustGetenv("DB_NAME")  // e.g. 'my-database'
		host   = mustGetenv("DNS_NAME") // e.g. 'project:region:instance'
	)
	// Connection string using Private Service Connect DNS
	dsn := fmt.Sprintf("user=%s password=%s host=%s database=%s", dbUser, dbPwd, host, dbName)
	// println(dsn)
	// Open the database connection with pgx driver
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Unable to open connection: %v", err)
	} 
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	log.Println("Successfully connected to Cloud SQL using Private Service Connect!")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		stat, err := json.Marshal(db.Stats())
		if err != nil {
			log.Fatal(err)
		}
		w.Write(stat)
	})

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, value FROM public.employee")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(err.Error()))
			return
		}
		defer rows.Close()
		type employee struct {
			ID    int
			Name  string
			Value string
		}
		employees := make([]employee, 0)
		for rows.Next() {
			var employee employee
			if err := rows.Scan(&employee.ID, &employee.Name, &employee.Value); err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			employees = append(employees, employee)
		}
		data, err := json.Marshal(employees)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
	log.Fatal(http.ListenAndServe(":2000", nil))

}

// func PostgresConnectionPool(Connection interface{}) {
// 	Con := Connection.(*sql.DB)
// 	var buffer bytes.Buffer
// 	buffer.WriteString("postgres://")
// 	buffer.WriteString(env.Username + ":" + env.Password)
// 	buffer.WriteString("@")
// 	buffer.WriteString(env.Host + ":" + env.Port + "/")
// 	buffer.WriteString(env.Name)
// 	buffer.WriteString("?sslmode=disable")
// 	connection_string := buffer.String()
// 	Connection, err := sql.Open(Constant.POSTGRES, connection_string)
// 	if err != nil {
// 		//panic err
// 		panic(err.Error())
// 	}
// 	Connection.(*sql.DB).SetMaxOpenConns(env.Maximum_connection)
// 	*Con = *Connection.(*sql.DB)
// 	err = Con.Ping()
// 	if err != nil {
// 		log.Fatal().Msgf("Couldn't connect to the Postgres %v", err)
// 		panic(err.Error())
// 	} else {
// 		log.Info().Msg("Postgres Connected!")
// 	}
// }
