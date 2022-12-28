package librarejob

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NOTE: Cookie analysis
//  rarejob_auto_login == rarejob_onetime_key
//  PHPSESSID and PHPSESSID_HIGH are session ids
//  once rarejob_onetime_key and PHPSESSID are deleted, session is closed and we're redirected to login page.

type Reserve struct {
	Name    string
	StartAt time.Time
	EndAt   time.Time
}

type Tutor struct {
	Name           string
	AvailableSlots []time.Time
}

func (t Tutor) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", t.Name)
	// TODO(musaprg): output availableslots
	return nil
}

type Tutors []Tutor

func (ts Tutors) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, t := range ts {
		enc.AppendObject(t)
	}
	return nil
}

type Client interface {
	Login(ctx context.Context, username, password string) error
	ReserveTutor(ctx context.Context, from time.Time, by time.Duration) (*Reserve, error)
	Teardown() error
}

type client struct {
	s  *selenium.Service
	wd selenium.WebDriver
}

type ClientOpts struct {
	SeleniumHost        string
	SeleniumPort        *int
	SeleniumBrowserName string
	SeleniumDebug       bool
}

func NewClient(opts ClientOpts) (Client, error) {
	defer zap.L().Sync()

	var s *selenium.Service
	var err error
	url := "127.0.0.1"
	port := 4444
	browserName := "firefox"
	if opts.SeleniumHost == "" {
		if opts.SeleniumPort != nil {
			port = *opts.SeleniumPort
		}
		s, err = startLocalSelenium(port, opts.SeleniumDebug)
		if err != nil {
			return nil, err
		}
	}
	if opts.SeleniumHost != "" {
		url = opts.SeleniumHost
	}
	if opts.SeleniumPort != nil {
		port = *opts.SeleniumPort
	}
	urlPrefix := fmt.Sprintf("http://%s:%d/wd/hub", url, port)

	if opts.SeleniumBrowserName != "" {
		browserName = opts.SeleniumBrowserName
	}

	caps := selenium.Capabilities{"browserName": browserName}

	// Connect to the WebDriver instance running locally.
	var wd selenium.WebDriver
	for i := 0; i < maxSeleniumHealthCheckBackoffLimit; i++ {
		wd, err = selenium.NewRemote(caps, urlPrefix)
		if err != nil {
			zap.L().Warn("failed to access to the selenium server, retrying...", zap.Error(err))
			time.Sleep(time.Second * seleniumHealthCheckRetrySecond)
		}
	}
	if err != nil {
		return nil, err
	}

	return &client{
		s:  s,
		wd: wd,
	}, nil
}

func startLocalSelenium(port int, debug bool) (*selenium.Service, error) {
	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	const (
		// These paths will be different on your system.
		seleniumPath    = "/opt/selenium/selenium-server-standalone.jar"
		geckoDriverPath = "/usr/bin/geckodriver"
	)
	so := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
	}
	if debug {
		so = append(so, selenium.Output(os.Stdout))
		selenium.SetDebug(debug)
	}
	return selenium.NewSeleniumService(seleniumPath, port, so...)
}

func (c *client) Login(ctx context.Context, username, password string) error {
	// TODO(musaprg): Cache SESSIONID and reuse
	if err := c.wd.Get(rarejobLoginURL); err != nil {
		return fmt.Errorf("failed to access rarejob login page: %w", err)
	}

	_ = waitUntilElementLoaded(c.wd, selenium.ByCSSSelector, loginPageEmailSelector)

	if emailInput, err := c.wd.FindElement(selenium.ByCSSSelector, loginPageEmailSelector); err != nil {
		return fmt.Errorf("failed to find the email input box: %w", err)
	} else {
		emailInput.SendKeys(os.Getenv("RAREJOB_EMAIL"))
	}

	_ = waitUntilElementLoaded(c.wd, selenium.ByCSSSelector, loginPagePasswordSelector)

	if passwordInput, err := c.wd.FindElement(selenium.ByCSSSelector, loginPagePasswordSelector); err != nil {
		return fmt.Errorf("failed to find the password input box: %w", err)
	} else {
		passwordInput.SendKeys(os.Getenv("RAREJOB_PASSWORD"))
	}

	if submit, err := c.wd.FindElement(selenium.ByName, "yt0"); err != nil {
		return fmt.Errorf("failed to find submit button: %w", err)
	} else {
		submit.Click()
	}

	if err := c.wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		return wd.SessionID() != "", nil
	}); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	return nil
}

func (c *client) ReserveTutor(ctx context.Context, from time.Time, margin time.Duration) (*Reserve, error) {
	defer zap.L().Sync()

	// TODO(musaprg): split this function into two

	// -- Search available tutors --

	by := from.Local().Add(margin)
	if !(margin < 24*time.Hour && from.Hour() <= by.Hour()) {
		return nil, ErrSpreadAcrossTwoDays
	}

	queryURL, err := generateTutorSearchQuery(from, by)
	if err != nil {
		return nil, fmt.Errorf("failed to generate search query: %w", err)
	}
	if err := c.wd.Get(queryURL); err != nil {
		return nil, fmt.Errorf("failed to get availabe tutor list: %w", err)
	}

	waitUntilElementLoaded(c.wd, selenium.ByCSSSelector, tutorListSelector)
	tutorList, err := c.wd.FindElements(selenium.ByCSSSelector, tutorListSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to get tutor info: %w", err)
	}

	var tutors Tutors
	// TODO(musaprg): parallelize with goroutine and use errgroup to aggregate error
	for tnum := 1; tnum <= len(tutorList); tnum++ {
		zap.L().Debug("getting tutor info", zap.Int("number", tnum), zap.String("url", c.getCurrentURL()))
		nameElm, _ := c.wd.FindElement(selenium.ByCSSSelector, fmt.Sprintf(tutorNameSelector, tnum))
		name, _ := nameElm.Text()
		slotElms, err := c.wd.FindElements(selenium.ByCSSSelector, fmt.Sprintf(tutorTimeSlotSelector, tnum))
		if err != nil {
			return nil, fmt.Errorf("failed to get time slots for tutor #%d: %w", tnum, err)
		}
		var slots []time.Time
		for snum := 1; snum <= len(slotElms); snum++ {
			slotElm, err := c.wd.FindElement(selenium.ByCSSSelector, fmt.Sprintf(tutorTimeSlotButtonSelector, tnum, snum))
			if err != nil { // if err, fill zero time to preserve index
				slots = append(slots, time.Time{})
			}
			slotText, _ := slotElm.Text()
			h, m, err := parseTime(slotText)
			if err != nil {
				slots = append(slots, time.Time{})
			}
			slots = append(slots, time.Date(from.Year(), from.Month(), from.Day(), h, m, 0, 0, time.Local))
		}
		tutors = append(tutors, Tutor{
			Name:           name,
			AvailableSlots: slots,
		})
	}

	zap.L().Info("found tutors", zap.Array("tutors", tutors))

	// -- Do reservation --

	timeSlotButtonSelector := fmt.Sprintf(tutorTimeSlotButtonSelector, 1, 1)
	waitUntilElementLoaded(c.wd, selenium.ByCSSSelector, timeSlotButtonSelector)
	// TODO(musaprg): Implement to select tutor, not hard-coded
	timeSlot, err := c.wd.FindElement(selenium.ByCSSSelector, timeSlotButtonSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to find time slot button: %w", err)
	}
	{
		text, _ := timeSlot.Text()
		zap.L().Debug("found time slot button", zap.String("button_text", text))
	}
	if err := timeSlot.Click(); err != nil {
		return nil, fmt.Errorf("failed to click time slot button: %w", err)
	}

	zap.L().Debug("loading reservation page", zap.String("url", c.getCurrentURL()))
	waitUntilElementLoaded(c.wd, selenium.ByLinkText, "予約する")
	zap.L().Debug("loaded reservation page", zap.String("url", c.getCurrentURL()))
	reserveButton, err := c.wd.FindElement(selenium.ByLinkText, "予約する")
	if err != nil {
		zap.L().Debug("failed to get reserve button", zap.Error(err), zap.String("url", c.getCurrentURL()))
		return nil, fmt.Errorf("failed to get reserve button: %w", err)
	}
	if err := reserveButton.Click(); err != nil {
		return nil, fmt.Errorf("failed to click reserve button: %w", err)
	}

	zap.L().Debug("waiting for completion of reservation")
	waitUntilURLChanged(c.wd, rarejobReservationFinishURL)
	zap.L().Debug("reservation completed")

	return &Reserve{
		Name:    tutors[0].Name,
		StartAt: tutors[0].AvailableSlots[0],
		EndAt:   tutors[0].AvailableSlots[0].Add(25 * time.Minute),
	}, nil
}

func (c *client) Teardown() error {
	if c.wd != nil {
		if err := c.wd.Quit(); err != nil {
			return fmt.Errorf("failed to quit current webdriver session: %w", err)
		}
	}
	if c.s != nil {
		if err := c.s.Stop(); err != nil {
			return fmt.Errorf("failed to quit current webdriver session: %w", err)
		}
	}
	return nil
}
