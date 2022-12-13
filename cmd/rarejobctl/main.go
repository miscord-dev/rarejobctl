package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	seleniumHost        = flag.String("selenium-host", "127.0.0.1", "Remote Selenium Hostname")
	seleniumBrowserName = flag.String("selenium-browser-name", "firefox", "Remote Selenium Browser name")
)

const (
	// maxRetryReservation is the number of retry when the reservation failes somehow.
	maxRetryReservation = 5
)

func init() {
	flag.Parse()
}

// This is test main function
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	api := slack.New(os.Getenv("SLACK_API_TOKEN"))

	logger.Info("start initialization of rarejob client")

	opts := librarejob.ClientOpts{
		SeleniumHost:        *seleniumHost,
		SeleniumPort:        seleniumPort,
		SeleniumBrowserName: *seleniumBrowserName,
	}
	rc, err := librarejob.NewClient(opts)
	if err != nil {
		api.PostMessage(os.Getenv("SLACK_CHANNEL"), slack.MsgOptionText("something went wrong... I failed to reserve your tutor. try again later.", false), slack.MsgOptionAsUser(true))
		logger.Fatal("failed to create rarejob client", zap.Error(err))
	}
	defer rc.Teardown()

	logger.Info("initialized rarejob client")

	logger.Info("attempting to login rarejob...")

	if err := rc.Login(context.TODO(), os.Getenv("RAREJOB_EMAIL"), os.Getenv("RAREJOB_PASSWORD")); err != nil {
		api.PostMessage(os.Getenv("SLACK_CHANNEL"), slack.MsgOptionText("something went wrong... I failed to reserve your tutor. try again later.", false), slack.MsgOptionAsUser(true))
		logger.Fatal("failed to login", zap.Error(err))
	}

	logger.Info("logged in to rarejob")

	tt := strings.Split(*t, ":")
	if len(tt) != 2 {
		logger.Fatal("invalid time format", zap.String("input", *t))
	}
	hour, _ := strconv.Atoi(tt[0])
	minute, _ := strconv.Atoi(tt[1])
	from := time.Date(*year, time.Month(*month), *day, hour, minute, 0, 0, time.Local)

	var r *librarejob.Reserve

	logger.Info("start reserving tutor", zap.Int("year", *year), zap.Int("month", *month), zap.Int("day", *day), zap.String("time", *t))
	for attempt := 0; attempt < maxRetryReservation; attempt++ {
		r, err = rc.ReserveTutor(context.TODO(), from, time.Minute*time.Duration(*margin))
		if r != nil {
			break
		}
		if err != nil {
			logger.Warn("failed to reserve tutor. retrying...", zap.Error(err))
		}
	}
	if err != nil {
		api.PostMessage(os.Getenv("SLACK_CHANNEL"), slack.MsgOptionText("something went wrong... I failed to reserve your tutor. try again later.", false), slack.MsgOptionAsUser(true))
		logger.Fatal("failed to reserve tutor", zap.Error(err))
	}

	logger.Info("completed, posting status to slack")

	api.PostMessage(os.Getenv("SLACK_CHANNEL"), slack.MsgOptionText(fmt.Sprintf(`Reservation completed! Enjoy your EIKAIWA lesson yay.

Tutor Name: %s
Start: %s
End: %s
`, r.Name, r.StartAt, r.EndAt), false), slack.MsgOptionAsUser(true))

	logger.Info(fmt.Sprint("reserved tutor:", r))
}
