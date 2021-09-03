package main

import (
	"compress/flate"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/retbrown/dynamic-tiles/config"
	"github.com/retbrown/dynamic-tiles/helpers"
	"github.com/retbrown/dynamic-tiles/v1/land"
)

func main() {
	logConfig := zap.NewProductionEncoderConfig()
	logConfig.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format(time.RFC3339))
	}
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(logConfig),
		os.Stdout,
		zapcore.DebugLevel,
	))
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	conf := config.Gather(logger)

	r := chi.NewRouter()

	// Some basic middleware for logging
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(helpers.NewZapMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(flate.DefaultCompression))
	r.Use(middleware.Timeout(60 * time.Second))

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-uuid"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	r.Use(cors.Handler)

	logger.Info("Connecting to the main database")
	mainDB, err := sqlx.Connect("pgx", conf.ConnStr)
	if err != nil {
		logger.Fatal("Unable to establish connection to Master database", zap.Error(err))
	}
	defer mainDB.Close()

	// health check url for AWS
	r.Get("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	landRepository := &land.Repository{Db: mainDB}
	land := &land.HTTP{Repository: landRepository}

	// API after middleware. These are all authenticated routes
	r.Get("/{z}/{x}/{y}.{extension}", land.GetTile)

	portNumber := ":8080"
	logger.Info("HTTP Server running and listening", zap.String("portNumber", portNumber))

	srv := &http.Server{
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler:           r,
		Addr:              portNumber,
	}
	logger.Fatal("HTTP serve failed", zap.Error(srv.ListenAndServe()))
}
