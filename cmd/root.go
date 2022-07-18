/*
Copyright © 2022 Optriment
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cmdName = os.Args[0]
	release = "n/a"
	env     = "developer-env"
	built   = "now"
)

const sentryFlushTimeout = 10 * time.Second

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   cmdName,
	Short: "Optriment work application",
	Long:  `Optriment work application. For full commands list please use --help flag`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if level, e := zerolog.ParseLevel(viper.GetString(settLogLevel)); e == nil {
			zerolog.SetGlobalLevel(level)
		} else {
			log.Warn().Err(e).Msgf("Unable to parse log level '%s'", viper.GetString(settLogLevel))
		}

		viper.SetDefault(settCfgRelease, release)
		viper.SetDefault(settCfgEnv, env)
		viper.SetDefault(settBuilt, built)

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              viper.GetString(settCfgSentryDSN),
			Environment:      viper.GetString(settCfgEnv),
			Release:          viper.GetString(settCfgRelease),
			Debug:            true,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed sentry.Init")
		} else {
			log.Info().Msg("Sentry was successfully initialized")
		}

		log.Logger = log.Hook(&goroutineID{Name: "GID"})
		if viper.GetBool(settLogCaller) {
			log.Logger = log.With().Caller().Logger()
		}
		pphp := viper.GetString(settPprofHostport)
		if len(pphp) > 0 {
			startPprof(pphp)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer func() {
		if sentry.Flush(sentryFlushTimeout) {
			log.Debug().Msg("Sentry successfully flushed")
		} else {
			log.Warn().Msg("Sentry failed to flush")
		}
	}()

	// context should be canceled while Int signal will be caught
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// main processing loop
	retChan := make(chan error, 1)
	go func() {
		err2 := rootCmd.ExecuteContext(ctx)
		if err2 != nil {
			retChan <- err2
		}
		close(retChan)
	}()

	// listening OS signals to terminate
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		log.Warn().Msgf("Signal '%s' was caught. Exiting", <-quit)
		cancel()
	}()

	// Listening for the main loop response
	for e := range retChan {
		log.Info().Err(e).Msg("Exiting.")
	}
}

// registerFlags registers specified flags for cobra and viper
func registerFlags(cc *cobra.Command, registerFunc func(cc *cobra.Command)) {
	registerFunc(cc)
	if err := viper.BindPFlags(cc.PersistentFlags()); err != nil {
		panic("unable to bind flags " + err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	registerFlags(rootCmd, func(cc *cobra.Command) {
		cc.PersistentFlags().StringVarP(&cfgFile, "config", "F", "", "config file (default is $HOME/.optrispace.yaml)")

		cc.PersistentFlags().StringP(settLogLevel, "L", "info", "log level (trace, debug, info, warn, error, fatal, panic)")
		cc.PersistentFlags().Bool(settLogCaller, false, "output caller function name in log (may impact performance if using)")

		cc.PersistentFlags().String(settPprofHostport, "", "host:port for start and exposing diagnostic info; not exposing, if unset; example — localhost:6060, and use http://localhost:6060/debug/pprof/ after start")

		cc.PersistentFlags().String(settCfgSentryDSN, "", "Sentry.io DSN. Sentry will be not available, of this value is not specified. The better way to specify this value, use appropriate environment variable.")
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".optrispace" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".optrispace")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("optr")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug().Msgf("Using config file: " + viper.ConfigFileUsed())
	}
}
