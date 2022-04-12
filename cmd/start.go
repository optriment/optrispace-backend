/*
Copyright © 2022 Optriment

*/
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"optrispace.com/work/pkg/controller"
	"optrispace.com/work/pkg/service"
)

const (
	shutdownTimeout = 30 * time.Second
)

var errSpecifyServerHost = errors.New("specify server host to listen")

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"run"},
	Short:   "Runs server",
	Long:    `Runs Optriment server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doStart(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	registerFlags(startCmd, func(cc *cobra.Command) {
		cc.PersistentFlags().StringP(settServerHost, "s", ":8080", "server:host for listen incoming HTTP requests")
		cc.PersistentFlags().Bool(settServerTrace, false, "trace all requests to server")

		cc.PersistentFlags().StringP(settDbUrl, "d", "postgres://postgres:postgres@localhost:5432/optrwork?sslmode=disable", "database connection URL")

		cc.PersistentFlags().Bool(settHideBanner, false, "hide banner")
	})
}

func doStart(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	e := echo.New()

	e.HideBanner = viper.GetBool(settHideBanner)
	e.Pre(middleware.RemoveTrailingSlash())

	if viper.GetBool(settServerTrace) {
		e.Use(middleware.Logger())
	}

	standardHandler := e.HTTPErrorHandler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		standardHandler(err, c)
		log.Error().Err(err).Int("status", c.Response().Status).Stringer("url", c.Request().URL).
			Msg("Failed to process request")
	}
	e.Use(middleware.Recover())

	// testing stuff
	e.Any("echo", func(c echo.Context) error {
		return c.String(http.StatusOK, "Ok!")
	})

	// stopping
	controller.AddStop(e, cancel)

	if err := addControllers(ctx, e); err != nil {
		log.Error().Err(err).Msg("Unable to create application")
	}

	listenHost := viper.GetString(settServerHost)
	if strings.TrimSpace(listenHost) == "" {
		return fmt.Errorf("%w: %s settings", errSpecifyServerHost, settServerHost)
	}
	go e.Start(listenHost)

	<-ctx.Done()

	shutdownCtx, _ := context.WithTimeout(context.Background(), shutdownTimeout)

	return e.Shutdown(shutdownCtx)
}

func addControllers(ctx context.Context, e *echo.Echo) error {
	var rr []controller.Registerer

	db, err := sql.Open("postgres", viper.GetString(settDbUrl))
	if err != nil {
		return fmt.Errorf("unable to open DB: %w", err)
	}
	go func() {
		<-ctx.Done()
		err := db.Close()
		log.Debug().Err(err).Msg("Closing postgres DB")
	}()

	rr = append(rr,
		controller.NewJob(service.NewJob(db)),
		controller.NewApplication(service.NewApplication(db)),
		controller.NewPerson(service.NewPerson(db)),
	)

	for _, r := range rr {
		r.Register(e)
	}

	return nil
}
