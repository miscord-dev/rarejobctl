package librarejob

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"go.uber.org/zap"
)

// TODO(musaprg): dirty logic, needs to be refactored
func waitUntilElementLoaded(wd selenium.WebDriver, by, value string) error {
	var err error
	if waitErr := wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		elm, err := wd.FindElement(by, value)
		zap.L().Debug("checking if the element has been loaded", zap.String("by", by), zap.String("value", value))
		if err == nil {
			text, _ := elm.Text()
			zap.L().Debug("element has been loaded", zap.String("by", by), zap.String("value", value), zap.String("text", text))
		}
		return err == nil, nil
	}, defaultWaitTimeout, defaultWaitInterval); waitErr != nil {
		return err
	}
	return nil
}

func waitUntilURLChanged(wd selenium.WebDriver, url string) error {
	var err error
	u := ""
	if waitErr := wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		u, err = wd.CurrentURL()
		if err != nil {
			return false, err
		}
		zap.L().Debug("checking if the url has been changed", zap.String("url", u))
		return u == url, nil
	}, defaultWaitTimeout, defaultWaitInterval); waitErr != nil {
		return err
	}
	return nil
}

func generateTutorSearchQuery(from, by time.Time) (string, error) {
	s, err := strconv.Atoi(from.Format("1504"))
	if err != nil {
		return "", err
	}
	e, err := strconv.Atoi(by.Format("1504"))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(rarejobTutorSearchURL, from.Local().Year(), from.Local().Month(), from.Local().Day(), s, e), nil
}

func parseTime(s string) (h, m int, err error) {
	if t := strings.Split(s, ":"); len(t) == 2 {
		h, err = strconv.Atoi(t[0])
		if err != nil {
			return
		}
		m, err = strconv.Atoi(t[1])
		if err != nil {
			return
		}
	}
	return
}

func (c *client) getCurrentURL() string {
	url, err := c.wd.CurrentURL()
	if err != nil {
		zap.L().Debug("current url is empty", zap.Error(err))
	}
	return url
}
