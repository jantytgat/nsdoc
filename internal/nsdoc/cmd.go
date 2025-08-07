package nsdoc

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/Oudwins/zog"
	"github.com/jantytgat/go-kit/application"
	"github.com/jantytgat/go-kit/flagzog"
	"github.com/jantytgat/go-netscaleradc-nitro/pkg/nitro"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	environmentNameFlag       = flagzog.NewStringFlag("environment-name", zog.String().Min(1).Not().ContainsSpecial(), "Environment Name")
	addressFlag               = flagzog.NewStringFlag("address", zog.String().Min(1).Not().ContainsSpecial(), "IP Address or FQDN")
	usernameFlag              = flagzog.NewStringFlag("username", zog.String().Min(1).Not().ContainsSpecial(), "Username")
	passwordFlag              = flagzog.NewStringFlag("password", zog.String().Min(1).Not().ContainsSpecial(), "Password")
	useInsecureProtocol       = flagzog.NewBoolFlag("http", zog.Bool(), "Use HTTP instead of HTTPS")
	skipCertificateValidation = flagzog.NewBoolFlag("skip-certificate-validation", zog.Bool(), "Skip certificate validation")
)

var AnalyzeCommand = application.Command{
	Command: &cobra.Command{
		Use:   "analyze",
		Short: "Analyze a single environment",
		Long:  "Analyze a single environment",
		RunE:  AnalyzeFuncE,
	},
	SubCommands: []application.Commander{},
	Configure: func(c *cobra.Command) {
		c.Flags().StringVarP(&environmentNameFlag.Value, environmentNameFlag.Name(), "", "", environmentNameFlag.Usage())
		c.Flags().StringVarP(&addressFlag.Value, addressFlag.Name(), "", "", addressFlag.Usage())
		c.Flags().StringVarP(&usernameFlag.Value, usernameFlag.Name(), "", "", usernameFlag.Usage())
		c.Flags().StringVarP(&passwordFlag.Value, passwordFlag.Name(), "", "", passwordFlag.Usage())
		c.Flags().BoolVarP(&useInsecureProtocol.Value, useInsecureProtocol.Name(), "", false, useInsecureProtocol.Usage())
		c.Flags().BoolVarP(&skipCertificateValidation.Value, skipCertificateValidation.Name(), "", false, skipCertificateValidation.Usage())
	},
}

var DaemonCommand = application.Command{
	Command: &cobra.Command{
		Use:   "daemon",
		Short: "Run a daemon",
		Long:  "Run a daemon",
		RunE:  application.HelpFuncE,
	},
	SubCommands: []application.Commander{},
	Configure:   nil,
}

func AnalyzeFuncE(cmd *cobra.Command, args []string) error {
	fmt.Println("Environment:", environmentNameFlag.Value)
	fmt.Println("Address:", addressFlag.Value)
	fmt.Println("Username:", usernameFlag.Value)
	if passwordFlag.Value == "" {
		fmt.Print("Password: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("\nFailed to read password:")
			fmt.Println(err.Error())
			return err
		}
		passwordFlag.Value = string(password)
		fmt.Println()
	}

	nc, err := nitro.NewClient(
		environmentNameFlag.Value,
		addressFlag.Value,
		nitro.Credentials{
			Username: usernameFlag.Value,
			Password: passwordFlag.Value,
		},
		nitro.ConnectionSettings{
			UseSsl:                    !useInsecureProtocol.Value,
			Timeout:                   3000,
			UserAgent:                 "nsdoc",
			ValidateServerCertificate: !skipCertificateValidation.Value,
			LogTlsSecrets:             false,
			LogTlsSecretsDestination:  "",
			AutoLogin:                 true,
		},
	)
	if err != nil {
		fmt.Println(errors.Unwrap(errors.Unwrap(err)))
		return err
	}

	defer func() {
		e := nc.Logout()
		if e != nil {
			err = e
		}
	}()

	fmt.Println("Connection - OK")
	time.Sleep(1 * time.Second)
	lbvs, err := nc.LbVserver.List(context.Background(), nil, nil)
	if err != nil {
		return err
	}
	if lbvs == nil {
		return errors.New("no lb vservers found")
	}
	for _, lbv := range lbvs {
		fmt.Println("LB VServer:", lbv.Name)
	}
	return nil
}
