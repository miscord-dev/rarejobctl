package librarejob

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

// TODO(musaprg): dirty logic, needs to be refactored
func waitUntilElementLoaded(wd selenium.WebDriver, by, value string) error {
	var err error
	if waitErr := wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		_, err = wd.FindElement(by, value)
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

func parseTime(s string) (h,m int, err error) {
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