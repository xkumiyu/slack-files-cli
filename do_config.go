package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/urfave/cli"
)

type usersList struct {
	OK      bool     `json:"ok"`
	Members []member `json:"members"`
}

type member struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ConfigCommand is definition of config command
var ConfigCommand = cli.Command{
	Name:    "config",
	Aliases: []string{"c"},
	Usage:   "view/set config",
	Action:  doView,
	Subcommands: []cli.Command{
		setCommand,
		viewCommand,
	},
}

var viewCommand = cli.Command{
	Name:   "view",
	Usage:  "view current config valius",
	Action: doView,
}

var setCommand = cli.Command{
	Name:   "set",
	Usage:  "set config values",
	Action: doSet,
	Flags: []cli.Flag{
		// cli.BoolFlag{
		// 	Name:  "global, g",
		// 	Usage: "global",
		// },
		cli.StringFlag{
			Name:  "user, u",
			Usage: "slack username",
		},
		cli.StringFlag{
			Name:  "token, t",
			Usage: "authentication token (generate tokens with https://api.slack.com/custom-integrations/legacy-tokens)",
		},
	},
}

func doSet(c *cli.Context) error {
	userName := c.String("user")
	token := c.String("token")
	// global := c.Bool("global")

	cfg := new(Config)
	if Exists(ConfigFile) {
		cfg = ReadConfig()
	}

	if token != "" {
		cfg.Token = token
	}
	if userName != "" {
		cfg.UserName = userName
	}
	cfg.UserID = getUserID(cfg.UserName, cfg.Token)

	bytes, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(ConfigFile, bytes, 0644); err != nil {
		log.Fatal(err)
	}

	return nil
}

func getUserID(userName string, token string) string {
	values := url.Values{}
	values.Set("token", token)

	resp, err := http.Get("https://slack.com/api/users.list" + "?" + values.Encode())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	usersList := new(usersList)
	if err := json.Unmarshal(bytes, usersList); err != nil {
		log.Fatal(err)
	}

	userID := func(members []member, userName string) string {
		for _, member := range members {
			if member.Name == userName {
				return member.ID
			}
		}
		return ""
	}(usersList.Members, userName)

	return userID
}

func doView(c *cli.Context) error {
	cfg := ReadConfig()
	fmt.Printf("userid: %s\n", cfg.UserID)
	fmt.Printf("username: %s\n", cfg.UserName)
	fmt.Printf("token: %s\n", cfg.Token)
	return nil
}
