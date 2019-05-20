package main

import (
	"errors"
	"os"
	"sort"
	"time"

	application "github.com/MindHunter86/icqdumper/app"
	"github.com/rs/zerolog"
	"gopkg.in/urfave/cli.v1"
)

// todo
// -- add recover in mongodb.UpdateOne

var log zerolog.Logger

func main() {

	// log initialization:
	zerolog.ErrorFieldName = "ERROR"
	log = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr}).With().Timestamp().Logger()

	// define app metadata:
	app := cli.NewApp()
	app.Name = "icqdumper"
	app.Version = "0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Vadimka Komissarov",
			Email: "v.komissarov@corp.mail.ru, v.komissarov@corp.vk.com, vadimka_kom@mail.ru"}}
	app.Copyright = "(c) 2019 Mindhunter and CO"
	app.Usage = "Dump history from ICQ chats"

	// define global flags:
	var globAppFlags []cli.Flag = []cli.Flag{
		cli.StringFlag{
			Name:  "loglevel, l",
			Value: "debug",
			Usage: "Dumper log level (debug, info, warn, error, fatal, panic) [debug]",
		},
		cli.BoolFlag{
			Name:  "silent, s",
			Usage: "",
		},
		cli.StringFlag{
			Name:   "aimsid, a",
			Value:  "",
			EnvVar: "ICQ_AIMSID",
			Usage:  "Bot or client AIMSID (megabot(70001) can help you)",
		},
		cli.StringFlag{
			Name:   "mongodb, m",
			Value:  "",
			EnvVar: "ICQ_MONGODB",
			Usage:  "mongodb connection string",
		},
		cli.IntFlag{
			Name:  "workers",
			Value: 128,
			Usage: "Workers count for parsing and saving chats and messages",
		},
		cli.IntFlag{
			Name:  "queuebuffer",
			Value: 1024000,
			Usage: "Number of unassigned buffered jobs",
		},
		cli.IntFlag{
			Name:  "workercapacity",
			Value: 1024,
			Usage: "Number of assigned buffered jobs",
		},
	}

	// commands define:
	app.Commands = []cli.Command{
		{
			Name:    "getHistory",
			Aliases: []string{"gh"},
			Usage:   "get chat history",
			Flags: append(globAppFlags, cli.StringFlag{
				Name:  "chat, c",
				Value: "all",
				Usage: "Chat for histroy dumping (default all)",
			}),
			Action: func(c *cli.Context) (e error) {

				log.Debug().Str("chat", c.String("chat")).Msg("Given ChatID")

				if len(c.String("aimsid")) == 0 {
					log.Info().Str("aimsid", c.String("aimsid")).Msg("Given AIMSID")
					return errors.New("AIMSID is undefined!")
				}

				if len(c.String("mongodb")) == 0 {
					return errors.New("MONGODB connection string is empty!")
				}

				if !c.Bool("silent") {
					switch c.String("loglevel") {
					case "off":
						zerolog.SetGlobalLevel(zerolog.NoLevel)
					case "debug":
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					case "info":
						zerolog.SetGlobalLevel(zerolog.InfoLevel)
					case "warn":
						zerolog.SetGlobalLevel(zerolog.WarnLevel)
					case "error":
						zerolog.SetGlobalLevel(zerolog.ErrorLevel)
					case "fatal":
						zerolog.SetGlobalLevel(zerolog.FatalLevel)
					case "panic":
						zerolog.SetGlobalLevel(zerolog.PanicLevel)
					}
				} else {
					zerolog.SetGlobalLevel(zerolog.NoLevel)
				}

				var app *application.App = application.NewApp(&log, &application.AppParams{
					Silent:         c.Bool("silent"),
					AimSid:         c.String("aimsid"),
					MongoConn:      c.String("mongodb"),
					Workers:        c.Int("workers"),
					QueueBuffer:    c.Int("queuebuffer"),
					WorkerCapacity: c.Int("workercapacity"),
				})

				return app.Bootstrap(c.String("chat"))
			},
		},
		{
			Name:    "sendIM",
			Aliases: []string{"sim"},
			Usage:   "send message",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "chat, c",
					Value: "all",
					Usage: "Chat for histroy dumping (default all)",
				},
			},
		},
		{
			Name:    "fetchEvents",
			Aliases: []string{"fe"},
			Usage:   "start event fetcher",
		},
		{
			Name:    "listenHistory",
			Aliases: []string{"lh"},
			Usage:   "start histroy listener",
		},
	}

	// sort all falgs && cmds:
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	// parse all given arguments:
	if e := app.Run(os.Args); e != nil {
		log.Fatal().Err(e).Msg("Could not run the App!")
	}
}
