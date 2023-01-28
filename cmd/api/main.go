package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// undescore (alias) is used to avoid go compiler complaining or erasing this
	// library.
	_ "github.com/lib/pq"
	"github.com/shynggys9219/greenlight/internal/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const version = "1.0.0"

// Add a db struct field to hold the configuration settings for our database connection
// pool. For now this only holds the DSN, which we will read in from a command-line flag.
type config struct {
	port int
	env  string
	db   struct {
		dsn string // a conenction string to a sql server
		// maxOpenConns int    // limit on the number of ‘open’ connections
		// maxIdleConns int    // limit on the number of idle connections in the pool
		// maxIdleTime  string // the maximum length of time that a connection can be idle
		// maxLifetime  string //optional here; maximum length of time that a connection can be reused for
	}
	jwt struct {
		secret string // Add a new field to store the JWT signing secret.
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models // hold new models in app
	// mailer mailer.Mailer
}

// key VxSmQ7LHd9d7vfgTEoIoQ9RHe0zmjVTJ4rCYsWDJH2HA6F39pxBlhwpl7TNFbA6
func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// flag.StringVar(&cfg.db.dsn, "db-dsn", "mongodb://localhost:27017", "Mongo DSN")

	// // Setting restrictions on db connections
	// flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	// flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	// flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle time")

	flag.StringVar(&cfg.jwt.secret, "jwt-secret", "pei3einoh0Beem6uM6Ungohn2heiv5lah1ael4joopie5JaigeikoozaoTew2Eh6", "JWT secret")

	// flag.StringVar(&cfg.db.maxLifetime, "db-max-lifetime", "1h", "PostgreSQL max idle time")

	flag.Parse()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		logger.Fatalf("Connection failed. Error is: %s", err)
	}

	// db will be closed before main function is completed.
	logger.Printf("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db), // data.NewModels() function to initialize a Models struct
	}
	// Use the httprouter instance returned by app.routes() as the server handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	// reuse defined variable err
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// func openDB(cfg config) (*sql.DB, error) {

// 	db, err := sql.Open("postgres", cfg.db.dsn)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// db.SetMaxIdleConns(cfg.db.maxIdleConns)
// 	// db.SetMaxOpenConns(cfg.db.maxOpenConns)

// 	// duration, err := time.ParseDuration(cfg.db.maxIdleTime)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// db.SetConnMaxIdleTime(duration)

// 	// optional lifetime limit, to use this, uncomment db substruct field and corresponding flag stringvar
// 	// lifetime, err := time.ParseDuration(cfg.db.maxIdleTime)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// db.SetConnMaxLifetime(lifetime)

// 	//context with a 5 second timeout deadline
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	err = db.PingContext(ctx) //create a connection and verify that everything is set up correctly.

// 	if err != nil {
// 		return nil, err
// 	}

// 	return db, nil
// }
