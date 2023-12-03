package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/webhook"
	"github.com/musaprg/rarejobctl/librarejob"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

var (
	year                = flag.Int("year", time.Now().Local().Year(), "year")
	month               = flag.Int("month", int(time.Now().Local().Month()), "month")
	day                 = flag.Int("day", time.Now().Day(), "day")
	t                   = flag.String("time", "10:30", "time formatted in HH:MM")
	margin              = flag.Int("margin", 30, "allowed margin, unit is minute")
	seleniumPort        = flag.Int("selenium-port", 4444, "Remote Selenium port")
	seleniumHost        = flag.String("selenium-host", "", "Remote Selenium Hostname")
	seleniumBrowserName = flag.String("selenium-browser-name", "firefox", "Remote Selenium Browser name")
	debug               = flag.Bool("debug", false, "enable debug mode")
	maxRetryReservation = flag.Int("max-retry", 5, "max number of attempts for reservation")

	// via Slack API
	slackAPIToken = os.Getenv("SLACK_API_TOKEN")
	slackChannel  = os.Getenv("SLACK_CHANNEL")

	slackAPI = slack.New(slackAPIToken)

	// via Slack incoming webhook
	slackWebhookURL = os.Getenv("SLACK_WEBHOOK_URL")

	// via Discord incoming webhook
	discrdWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")

	discordWebhookClient, _ = webhook.NewWithURL(discrdWebhookURL)
)

func init() {
	flag.Parse()
}

// This is test main function
func main() {
	var l *zap.Logger
	var err error
	if *debug {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}
	defer l.Sync()
	if err != nil {
		zap.L().Fatal("failed to initialize logger", zap.Error(err))
	}
	zap.ReplaceGlobals(l)

	zap.L().Info("start initialization of rarejob client")

	opts := librarejob.ClientOpts{
		SeleniumHost:        *seleniumHost,
		SeleniumPort:        seleniumPort,
		SeleniumBrowserName: *seleniumBrowserName,
		ClientDebug:         *debug,
	}
	rc, err := librarejob.NewClient(opts)
	if err != nil {
		postMessage("something went wrong... I failed to reserve your tutor. try again later.")
		zap.L().Fatal("failed to create rarejob client", zap.Error(err))
	}
	defer rc.Teardown()

	zap.L().Info("initialized rarejob client")

	tt := strings.Split(*t, ":")
	if len(tt) != 2 {
		zap.L().Fatal("invalid time format", zap.String("input", *t))
	}
	hour, _ := strconv.Atoi(tt[0])
	minute, _ := strconv.Atoi(tt[1])
	from := time.Date(*year, time.Month(*month), *day, hour, minute, 0, 0, time.Local)

	var r *librarejob.Reserve

	zap.L().Info("start reserving tutor", zap.Int("year", *year), zap.Int("month", *month), zap.Int("day", *day), zap.String("time", *t))
	for attempt := 0; attempt < *maxRetryReservation; attempt++ {
		zap.L().Info("attempting to login rarejob...")

		if err := rc.Login(context.TODO(), os.Getenv("RAREJOB_EMAIL"), os.Getenv("RAREJOB_PASSWORD")); err != nil {
			postMessage("something went wrong... I failed to reserve your tutor. try again later.")
			zap.L().Fatal("failed to login", zap.Error(err))
		}

		zap.L().Info("attempting to reserve tutor", zap.Int("attempt", attempt+1))
		r, err = rc.ReserveTutor(context.TODO(), from, time.Minute*time.Duration(*margin))
		if r != nil {
			break
		}
		if err != nil {
			zap.L().Warn("failed to reserve tutor. retrying...", zap.Error(err), zap.Int("attempt", attempt+1))
		}
	}
	if err != nil {
		postMessage("something went wrong... I failed to reserve your tutor. try again later.")
		zap.L().Fatal("failed to reserve tutor", zap.Error(err))
	}

	zap.L().Info("completed, posting status to slack")

	postMessage(fmt.Sprintf(`Reservation completed! Enjoy your EIKAIWA lesson yay.

Tutor Name: %s
Start: %s
End: %s
`, r.Name, r.StartAt, r.EndAt))

	zap.L().Info(fmt.Sprint("reserved tutor:", r))
}

func postMessage(text string) {
	switch {
	case slackAPIToken != "":
		slackAPI.PostMessage(slackChannel, slack.MsgOptionText(text, false), slack.MsgOptionAsUser(true))

	case slackWebhookURL != "":
		slack.PostWebhook(slackWebhookURL, &slack.WebhookMessage{
			Text: text,
		})

	case discordWebhookClient != nil:
		discordWebhookClient.CreateContent(text)
	}
}
