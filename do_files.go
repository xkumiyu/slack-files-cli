package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/urfave/cli"
)

type filesList struct {
	OK    bool   `json:"ok"`
	Files []file `json:"files"`
}

type file struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FileType string `json:"filetype"`
	Size     int    `json:"size"`
	URL      string `json:"url_private_download"`
	Created  int64  `json:"created"`
}

type filesDelete struct {
	OK bool `json:"ok"`
}

type filesListParam struct {
	User  string
	From  string
	To    string
	Types []string
}

func newfilesListParam(user string, from string, to string) *filesListParam {
	param := new(filesListParam)

	param.User = user
	if from, err := time.Parse("20060102", from); err == nil {
		param.From = strconv.FormatInt(from.Unix(), 10)
	}
	if to, err := time.Parse("20060102", to); err == nil {
		param.To = strconv.FormatInt(to.Unix(), 10)
	}

	return param
}

// FilesCommand is definition of files command
var FilesCommand = cli.Command{
	Name:    "files",
	Aliases: []string{"f"},
	Usage:   "list/donwload/delete files",
	Subcommands: []cli.Command{
		listCommand,
		downloadCommand,
		deleteCommand,
	},
}

var listCommand = cli.Command{
	Name:  "list",
	Usage: "show files list",
	Flags: append(
		filesFlags,
		cli.BoolFlag{
			Name:  "long, l",
			Usage: "list in long format",
		},
	),
	Action: doList,
}

var downloadCommand = cli.Command{
	Name:   "download",
	Usage:  "download files",
	Action: doDownload,
	Flags: append(
		filesFlags,
		cli.StringFlag{
			Name:  "path, p",
			Value: ".",
			Usage: "directory where files will be downloaded",
		},
	),
}

var deleteCommand = cli.Command{
	Name:   "delete",
	Usage:  "delete files",
	Flags:  filesFlags,
	Action: doDelete,
}

var filesFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "to",
		Usage: "filter files created before date (%Y%m%d)",
	},
	cli.StringFlag{
		Name:  "from",
		Usage: "filter files created after date (%Y%m%d)",
	},
	// cli.StringSliceFlag{
	// 	Name:  "types",
	// 	Value: "all",
	// 	Usage: "filter files by type",
	// },
}

func doList(c *cli.Context) error {
	long := c.Bool("long")
	cfg := ReadConfig()
	param := newfilesListParam(cfg.UserID, c.String("from"), c.String("to"))
	files := getFiles(*param, cfg.Token)

	if long {
		fmt.Printf("fileID\t\tcreatedDate\t\t\tfileName\n")
	}
	for _, file := range files {
		if long {
			created := time.Unix(file.Created, 0)
			fmt.Printf("%s\t%s\t%s\n", file.ID, created, file.Name)
		} else {
			fmt.Println(file.Name)
		}
	}

	return nil
}

func doDownload(c *cli.Context) error {
	downloadDir := c.String("path")
	cfg := ReadConfig()
	param := newfilesListParam(cfg.UserID, c.String("from"), c.String("to"))
	files := getFiles(*param, cfg.Token)

	if len(files) > 0 {
		fmt.Printf("Found %d files created by %s\n", len(files), cfg.UserName)
		downloadFiles(files, downloadDir, cfg.Token)
	} else {
		fmt.Printf("There is no files created by %s\n", cfg.UserName)
	}

	return nil
}

func doDelete(c *cli.Context) error {
	cfg := ReadConfig()
	param := newfilesListParam(cfg.UserID, c.String("from"), c.String("to"))
	files := getFiles(*param, cfg.Token)

	if len(files) > 0 {
		fmt.Printf("Found %d files created by %s.\n", len(files), cfg.UserName)
		deleteFiles(files, cfg.Token)
	} else {
		fmt.Printf("There is no files created by %s.\n", cfg.UserName)
	}

	return nil
}

func getFiles(param filesListParam, token string) []file {
	values := url.Values{}
	values.Set("token", token)
	values.Set("user", param.User)
	if param.From != "" {
		values.Set("ts_from", param.From)
	}
	if param.To != "" {
		values.Set("ts_to", param.To)
	}

	resp, err := http.Get("https://slack.com/api/files.list" + "?" + values.Encode())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	filesList := new(filesList)
	if err := json.Unmarshal(bytes, filesList); err != nil {
		log.Fatal(err)
	}

	return filesList.Files
}

func downloadFiles(files []file, downloadDir string, token string) {
	if !Exists(downloadDir) {
		if err := os.Mkdir(downloadDir, 0777); err != nil {
			log.Fatal(err)
		}
	}

	for _, file := range files {
		donwloadFilePath := filepath.Join(downloadDir, file.Name)

		req, _ := http.NewRequest("GET", file.URL, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := new(http.Client)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		fp, err := os.Create(donwloadFilePath)
		if err != nil {
			panic(err)
		}
		defer fp.Close()

		io.Copy(fp, resp.Body)
		fmt.Printf("Downloaded %s to %s\n", file.Name, downloadDir)
	}
}

func deleteFiles(files []file, token string) {
	values := url.Values{}
	values.Set("token", token)

	for _, file := range files {
		values.Set("file", file.ID)

		resp, err := http.PostForm("https://slack.com/api/files.delete", values)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		filesDelete := new(filesDelete)
		if err := json.Unmarshal(bytes, filesDelete); err != nil {
			log.Fatal(err)
		}

		if filesDelete.OK {
			fmt.Printf("Deleted the file %s\n", file.Name)
		} else {
			fmt.Printf("Could not delete the file %s\n", file.Name)
		}
	}
}
