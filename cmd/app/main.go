package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"maragu.dev/env"
	"maragu.dev/errors"
	gluehttp "maragu.dev/glue/http"
	gluejobs "maragu.dev/glue/jobs"
	"maragu.dev/glue/log"
	"maragu.dev/glue/sql"

	"app/html"
	"app/http"
	"app/jobs"
	"app/sqlite"
)

func main() {
	_ = env.Load(".env")
	_ = env.Load("/run/secrets/env")

	log := log.NewLogger(log.NewLoggerOptions{
		JSON:   env.GetBoolOrDefault("LOG_JSON", true),
		Level:  log.StringToLevel(env.GetStringOrDefault("LOG_LEVEL", "info")),
		NoTime: env.GetBoolOrDefault("LOG_NO_TIME", false),
	})

	if err := start(log); err != nil {
		log.Error("Error starting app", "error", err)
		os.Exit(1)
	}
}

func start(log *slog.Logger) error {
	appName := env.GetStringOrDefault("APP_NAME", "App")
	log.Info("Starting app", "name", appName)

	// Catch SIGTERM and SIGINT from the terminal, so we can do clean shutdowns.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	databaseLog := log.With("component", "sql.Database")

	jobTimeout := env.GetDurationOrDefault("JOB_QUEUE_TIMEOUT", 30*time.Second)

	db := sqlite.NewDatabase(sqlite.NewDatabaseOptions{
		H: sql.NewHelper(sql.NewHelperOptions{
			JobQueue: sql.JobQueueOptions{
				Timeout: jobTimeout,
			},
			Log: databaseLog,
			SQLite: sql.SQLiteOptions{
				Path: env.GetStringOrDefault("DATABASE_PATH", "app.db"),
			},
		}),
		Log: databaseLog,
	})
	if err := db.H.Connect(ctx); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	if err := db.H.MigrateUp(ctx); err != nil {
		return errors.Wrap(err, "error migrating database")
	}

	runner := gluejobs.NewRunner(gluejobs.NewRunnerOpts{
		Log:   log.With("component", "jobs.Runner"),
		Queue: db.H.JobsQ,
	})

	baseURL := env.GetStringOrDefault("BASE_URL", "http://localhost:8080")

	jobs.Register(runner, jobs.RegisterOpts{
		Log: log.With("component", "jobs"),
	})

	server := gluehttp.NewServer(gluehttp.NewServerOptions{
		Address:            env.GetStringOrDefault("SERVER_ADDRESS", ":8080"),
		BaseURL:            baseURL,
		CSP:                http.CSP(env.GetBoolOrDefault("CSP_ALLOW_UNSAFE_INLINE", false)),
		HTMLPage:           html.Page,
		HTTPRouterInjector: http.InjectHTTPRouter(log, db),
		Log:                log.With("component", "http.Server"),
		SecureCookie:       env.GetBoolOrDefault("SECURE_COOKIE", true),
	})

	// An error group is used to start and wait for multiple goroutines that can each fail with an error.
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return server.Start()
	})

	eg.Go(func() error {
		runner.Start(ctx)
		return nil
	})

	// Wait for the context to be done, which happens when the user sends a SIGTERM or SIGINT signal,
	// or if an error occurs in the errgroup.
	<-ctx.Done()
	log.Info("Stopping app")

	eg.Go(func() error {
		return server.Stop(ctx)
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	log.Info("Stopped app")

	return nil
}
