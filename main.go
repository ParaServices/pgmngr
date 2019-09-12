package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ParaServices/errgo"
	"github.com/ParaServices/pgmngr/pgmngr"
	"github.com/gookit/color"
	"github.com/urfave/cli"
)

const appRevisionTag = "0.1.0"

func displayErrorOrMessage(err error) error {
	if err != nil {
		color.Error.Tips(err.Error())
		errx, ok := err.(*errgo.Error)
		if ok {
			b, err := json.Marshal(errx)
			if err != nil {
				color.Error.Sprintf(err.Error())
			}
			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, b, "", "\t")
			if err != nil {
				color.Error.Sprintf(err.Error())
				return cli.NewExitError(color.Error.Sprintf(""), 1)
			}
			fmt.Println(string(prettyJSON.Bytes()))
			return cli.NewExitError(color.Error.Sprintf(""), 1)

		}
		color.Error.Tips(err.Error())
		return cli.NewExitError(color.Error.Sprintf(err.Error()), 1)
	}

	return nil
}

func main() {
	app := cli.NewApp()

	app.Name = "pgmngr"
	app.Usage = "Manage your Postgres database"
	app.Version = appRevisionTag

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config-file, c",
			Value: ".pgmngr.json",
			Usage: "Configures the path of the config file to be used to perform DB management",
		},
	}

	config := &pgmngr.Config{}
	app.Before = func(c *cli.Context) error {
		return displayErrorOrMessage(pgmngr.LoadConfig(c, config))
	}

	app.Commands = []cli.Command{
		{
			Name:  "migration",
			Usage: "migration commands",
			Subcommands: []cli.Command{
				{
					Name:  "new",
					Usage: "generate a new migration file",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "no-txn",
							Usage: "generatea a new migration that will not be wrapped in a transaction when executed",
						},
					},
					Action: func(c *cli.Context) error {
						if len(c.Args()) == 0 {
							return cli.NewExitError("migration name not given! try `pgmgr migration NameGoesHere`", 1)
						}

						return displayErrorOrMessage(pgmngr.CreateMigration(config, c.Args()[0], c.Bool("no-txn")))
					},
				},
			},
		},
		{
			Name:  "db",
			Usage: "manage your database. use 'pgmngr db help' for more info",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Usage: "creates the database if it doesn't exist",
					Action: func(c *cli.Context) error {
						return displayErrorOrMessage(pgmngr.CreateDatabase(*config))
					},
				},
				{
					Name:  "drop",
					Usage: "drops the database (all sessions must be disconnected first. this command does not force it)",
					Action: func(c *cli.Context) error {
						return displayErrorOrMessage(pgmngr.DropDatabase(*config))
					},
				},
				{
					Name:  "migrate",
					Usage: "applies any un-applied migrations in the migration folder",
					Action: func(c *cli.Context) error {
						return displayErrorOrMessage(pgmngr.ApplyMigration(pgmngr.Forward, config))
					},
				},
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		app.Command("help").Run(c)
		return nil
	}

	app.Run(os.Args)
}
