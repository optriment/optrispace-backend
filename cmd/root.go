/*
Copyright © 2022 Optriment
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cmdName = os.Args[0]
)

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
	// context should be canceled while Int signal will be caught
	ctx, cancel := context.WithCancel(context.Background())

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
		cc.PersistentFlags().StringVarP(&cfgFile, "config", "F", "", "config file (default is $HOME/.work.yaml)")

		cc.PersistentFlags().StringP(settLogLevel, "L", "info", "log level (trace, debug, info, warn, error, fatal, panic)")
		cc.PersistentFlags().Bool(settLogCaller, false, "output caller function name in log (may impact performance if using)")

		cc.PersistentFlags().String(settPprofHostport, "", "host:port for start and exposing diagnostic info; not exposing, if unset; example — localhost:6060, and use http://localhost:6060/debug/pprof/ after start")
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

		// Search config in home directory with name ".work" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".work")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug().Msgf("Using config file: " + viper.ConfigFileUsed())
	}
}
