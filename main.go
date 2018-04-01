package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "slack-files-cli"
	app.Usage = "Manege Slack files"
	app.Version = "0.0.1"
	app.Author = "xkumiyu"
	app.Email = "xkumiyu@gmail.com"
	app.Commands = []cli.Command{
		ConfigCommand,
		FilesCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
