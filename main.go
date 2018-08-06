package main

import (
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
)

func main() {
	app := getApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getApp() *cli.App {
	app := cli.NewApp()
	app.Name = "cfn-stack-launcher"
	app.HelpName = "cfn-stack-launcher"
	app.Version = "0.0.1"
	app.Usage = "Command Line Tools for CloudFormation Stacks"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		{
			Name:      "deploy",
			Usage:     "Deploy CloudFormation Stack",
			UsageText: "cfn-stack-launcher deploy FILE",
			Action: func(c *cli.Context) error {
				configFile := c.Args().Get(0)
				if len(configFile) < 1 {
					return cli.NewExitError("No config file specified", 1)
				}
				config, err := loadConfig(configFile)
				if err != nil {
					return err
				}

				launcher := newStackLauncher(config)
				err = launcher.Launch(config)
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:      "get-change-set",
			Usage:     "List of changes that will be applied to a stack",
			UsageText: "cfn-stack-launcher get-change-set FILE",
			Action: func(c *cli.Context) error {
				configFile := c.Args().Get(0)
				if len(configFile) < 1 {
					return cli.NewExitError("No config file specified", 1)
				}
				config, err := loadConfig(configFile)
				if err != nil {
					return err
				}

				launcher := newStackLauncher(config)
				err = launcher.GetChangeSet(config)
				if err != nil {
					return err
				}

				return nil
			},
		},
	}

	return app
}
