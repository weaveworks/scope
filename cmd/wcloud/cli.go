package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

func env(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

var (
	token   = env("SERVICE_TOKEN", "")
	baseURL = env("BASE_URL", "https://cloud.weave.works")
)

func usage() {
	fmt.Println(`Usage:
	deploy <image>:<version>   Deploy image to your configured env
	list                       List recent deployments
	config (<filename>)        Get (or set) the configured env
	logs <deploy>              Show lots for the given deployment`)
}

func main() {
	if len(os.Args) <= 1 {
		usage()
		os.Exit(1)
	}

	c := NewClient(token, baseURL)

	switch os.Args[1] {
	case "deploy":
		deploy(c, os.Args[2:])
	case "list":
		list(c, os.Args[2:])
	case "config":
		config(c, os.Args[2:])
	case "logs":
		logs(c, os.Args[2:])
	case "help":
		usage()
	default:
		usage()
	}
}

func deploy(c Client, args []string) {
	if len(args) != 1 {
		usage()
		return
	}
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) < 2 {
		usage()
		return
	}
	deployment := Deployment{
		ImageName: parts[0],
		Version:   parts[1],
	}
	if err := c.Deploy(deployment); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func list(c Client, args []string) {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	page := flags.Int("page", 0, "Zero based index of page to list.")
	pagesize := flags.Int("page-size", 10, "Number of results per page")
	if err := flags.Parse(args); err != nil {
		usage()
		return
	}
	deployments, err := c.GetDeployments(*page, *pagesize)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Created", "ID", "Image", "Version", "State"})
	table.SetBorder(false)
	table.SetColumnSeparator(" ")
	for _, deployment := range deployments {
		table.Append([]string{
			deployment.CreatedAt.Format(time.RFC822),
			deployment.ID,
			deployment.ImageName,
			deployment.Version,
			deployment.State,
		})
	}
	table.Render()
}

func loadConfig(filename string) (*Config, error) {
	extension := filepath.Ext(filename)
	var config Config
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if extension == ".yaml" || extension == ".yml" {
		if err := yaml.Unmarshal(buf, &config); err != nil {
			return nil, err
		}
	} else {
		if err := json.NewDecoder(bytes.NewReader(buf)).Decode(&config); err != nil {
			return nil, err
		}
	}
	return &config, nil
}

func config(c Client, args []string) {
	if len(args) > 1 {
		usage()
		return
	}

	if len(args) == 1 {
		config, err := loadConfig(args[0])
		if err != nil {
			fmt.Println("Error reading config:", err)
			os.Exit(1)
		}

		if err := c.SetConfig(config); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else {
		config, err := c.GetConfig()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		buf, err := yaml.Marshal(config)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println(string(buf))
	}
}

func logs(c Client, args []string) {
	if len(args) != 1 {
		usage()
		return
	}

	output, err := c.GetLogs(args[0])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(string(output))
}
