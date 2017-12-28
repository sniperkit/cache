package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	// github api
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	// cache stores
	"github.com/sniperkit/cache"                                  // cacher interface for http transporter
	cachebadger "github.com/sniperkit/cache/backend/local/badger" // BadgerKV default implementation

	// http stats
	"github.com/segmentio/stats"
	"github.com/segmentio/stats/httpstats"
	"github.com/segmentio/stats/influxdb"

	// misc
	"github.com/sniperkit/cache/helpers"
)

var (
	influxConfig *influxdb.ClientConfig
	influxClient *influxdb.Client
)

func main() {

	http.DefaultClient.Transport = httpstats.NewTransport(http.DefaultClient.Transport)

	// initialize new influxdb client
	influxConfig = influxdb.ClientConfig{
		Database:   "stats2",
		Address:    "127.0.0.1:8086",
		BufferSize: 2 * 1024 * 1024,
		Timeout:    5 * time.Second,
	}
	influxClient = influxdb.NewClientWith(influxConfig)
	influxClient.CreateDB("stats")

	// register engine
	stats.Register(influxClient)
	defer stats.Flush()

	cacheStoragePrefixPath := filepath.Join("shared", "cacher.badger")
	helpers.EnsureDir(cacheStoragePrefixPath)
	hcache, err := cachebadger.New(
		&cachebadger.Config{
			ValueDir:    "api.github.com.v3.gzip", //gzip",
			StoragePath: cacheStoragePrefixPath,
			SyncWrites:  true,
			Debug:       false,
			Compress:    true,
		})
	if err != nil {
		panic(err)
	}

	t := cache.NewTransport(hcache)
	t.MarkCachedResponses = true
	t.Debug = true
	t.Transport = httpstats.NewTransport(&http.Transport{})

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			fmt.Println("No token provided")
			os.Exit(1)
		}
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	hc := &http.Client{
		Transport: &oauth2.Transport{
			Base:   t, // httpstats.NewTransport(t),
			Source: ts,
		},
	}

	gh := github.NewClient(hc)
	ctx := context.Background()

	var u string = ""
	var r string = ""

	if len(os.Args) > 2 {
		u = os.Args[1]
		r = os.Args[2]
	} else if len(os.Args) > 0 {
		log.Fatal("User and Repository name not specified!")
	}

	var opts = github.ListOptions{
		Page:    0,
		PerPage: 1000, /* This option is intended to just set the max of GitHub */
	}

	var maxPage int = 1

	for opts.Page <= maxPage {
		stars, response, e := gh.Activity.ListStargazers(ctx, u, r, &opts)
		log.Println("response.Header.Get(cacher.XFromCache)? ", response.Header.Get(cacher.XFromCache))
		if e != nil {
			log.Fatal(e)
			os.Exit(1)
		}

		for _, star := range stars {
			user := star.GetUser()
			time := star.GetStarredAt()

			fmt.Println(strconv.FormatInt(time.Time.Unix(), 10) + "\t" + user.GetLogin())
		}

		if opts.Page == maxPage {
			break
		}

		maxPage = response.LastPage
		opts.Page = response.NextPage
	}

}
