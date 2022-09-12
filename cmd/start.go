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

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/controller"
	_ "optrispace.com/work/pkg/docs"
	"optrispace.com/work/pkg/service"
	"optrispace.com/work/pkg/service/ethsvc"
	"optrispace.com/work/pkg/web"
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

		cc.PersistentFlags().StringP(settDBURL, "d", "postgres://postgres:postgres@localhost:5432/optrwork?sslmode=disable", "database connection URL")

		cc.PersistentFlags().Bool(settHideBanner, false, "hide banner")

		cc.PersistentFlags().Bool(settServerCors, false, "enable CORS with allow any (*) origin")

		cc.PersistentFlags().StringP(settNotificationTgToken, "T", "", "telegram bot token for send notifications")
		cc.PersistentFlags().Int64SliceP(settNotificationTgChats, "C", nil, "telegram chat list for send notifications")
	})
}

// doStart creates and run echo instance with appropriate middlewares
func doStart(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	e := echo.New()

	e.HideBanner = viper.GetBool(settHideBanner)

	e.Pre(clog.PrepareContext)
	e.Use(sentryecho.New(sentryecho.Options{
		Repanic:         false,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	}))
	e.Pre(middleware.RemoveTrailingSlash())

	if viper.GetBool(settServerTrace) {
		e.Use(middleware.Logger())
	}

	if viper.GetBool(settServerCors) {
		e.Pre(middleware.CORS())
	}

	e.HTTPErrorHandler = web.GetErrorHandler(e.HTTPErrorHandler)
	e.Use(middleware.Recover())

	e.GET("info", func(c echo.Context) error {
		return c.JSON(http.StatusOK, c.Echo().Routes())
	})

	if err := addControllers(ctx, e); err != nil {
		log.Error().Err(err).Msg("Unable to create application")
	}

	listenHost := viper.GetString(settServerHost)
	if strings.TrimSpace(listenHost) == "" {
		return fmt.Errorf("%w: %s settings", errSpecifyServerHost, settServerHost)
	}
	go func() {
		errIn := e.Start(listenHost)
		log.Error().Err(errIn).Msg("Echo finished")
	}()

	<-ctx.Done()

	shutdownCtx, cancelTerminator := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelTerminator()

	return e.Shutdown(shutdownCtx)
}

// addControllers creates and registers controllers
func addControllers(ctx context.Context, e *echo.Echo) error {
	var rr []controller.Registerer

	db, err := sql.Open("postgres", viper.GetString(settDBURL))
	if err != nil {
		return fmt.Errorf("unable to open DB: %w", err)
	}
	go func() {
		<-ctx.Done()
		err := db.Close()
		log.Debug().Err(err).Msg("Closing postgres DB")
	}()

	sm := service.NewSecurity(db)

	e.Pre(web.Auth(sm))

	token := viper.GetString(settNotificationTgToken)
	chats := make([]int64, 0)

	for _, n := range viper.GetIntSlice(settNotificationTgChats) {
		chats = append(chats, int64(n))
	}

	eth := ethsvc.NewEthereum(viper.GetString(settEthereumURL))

	rr = append(rr,
		controller.NewAuth(sm, service.NewPerson(db)),
		controller.NewJob(sm, service.NewJob(db)),
		controller.NewApplication(sm, service.NewApplication(db)),
		controller.NewPerson(sm, service.NewPerson(db)),
		controller.NewContract(sm, service.NewContract(db, eth)),
		controller.NewNotification(service.NewNotification(token, chats...)),
		controller.NewStats(sm, service.NewStats(db)),
		controller.NewChat(sm, service.NewChat(db)),
	)

	controller.SwaggerRegister(e)
	for _, r := range rr {
		r.Register(e)
	}

	return nil
}
