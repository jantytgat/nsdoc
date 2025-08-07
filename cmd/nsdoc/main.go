package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jantytgat/go-kit/application"
	"github.com/jantytgat/go-kit/slogd"
	"github.com/spf13/cobra"

	"github.com/jantytgat/nsdoc/internal/nsdoc"
)

func main() {
	var err error
	slogd.Init(application.GetLogLevelFromArgs(os.Args), false)
	slogd.RegisterSink(slogd.HandlerText, slog.NewTextHandler(os.Stderr, slogd.HandlerOptions()), true)
	ctx := slogd.WithContext(context.Background())

	config := application.Config{
		Name:  "nsdoc",
		Title: "NetScaler Documentation Tool",
		// Banner @ https://patorjk.com/software/taag/#p=display&v=2&f=Slant&t=TLSMAN
		Banner:             "\n\n    _   _______ ____  ____  ______\n   / | / / ___// __ \\/ __ \\/ ____/\n  /  |/ /\\__ \\/ / / / / / / /     \n / /|  /___/ / /_/ / /_/ / /___   \n/_/ |_//____/_____/\\____/\\____/   \n                                  \n",
		PersistentPreRunE:  nil,
		PersistentPostRunE: nil,
		SubCommands: []application.Commander{
			nsdoc.AnalyzeCommand,
			nsdoc.DaemonCommand,
		},
		SubCommandInitializers: []func(cmd *cobra.Command){
			application.InitializeBannerOnSubCommands,
		},
		ValidArgs: nil,
	}

	var cmd *cobra.Command
	if cmd, err = config.BuildCommand(); err != nil {
		panic(err)
	}

	var app application.Application
	if app, err = application.New(cmd, application.NewDefaultQuitter(5*time.Second), slogd.Logger()); err != nil {
		panic(err)
	}

	if err = app.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
	}
}
