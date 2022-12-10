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
	year = flag.Int("year", time.Now().Local().Year(), "year")
	month = flag.Int("month", int(time.Now().Local().Month()), "month")
	day = flag.Int("day", time.Now().Day(), "day")
	t = flag.String("time", "10:30", "time formatted in HH:MM")
	margin = flag.Int("margin", 30, "allowed margin, unit is minute")
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

	rc, err := librarejob.NewClient()
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
	margin := 30 * time.Minute

	logger.Info("start reserving tutor", zap.Int("year", *year), zap.Int("month", *month), zap.Int("day", *day), zap.String("time", *t))
	r, err := rc.ReserveTutor(context.TODO(), from, margin)
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
