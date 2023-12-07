package librarejob

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"go.uber.org/zap"
)

// TODO(musaprg): dirty logic, needs to be refactored
func waitUntilElementLoaded(wd selenium.WebDriver, by, value string) error {
	if waitErr := wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		elm, err := wd.FindElement(by, value)
		zap.L().Debug("checking if the element has been loaded", zap.String("by", by), zap.String("value", value))
		if err == nil {
			text, _ := elm.Text()
			zap.L().Debug("element has been loaded", zap.String("by", by), zap.String("value", value), zap.String("text", text), zap.Error(err))
		}
		return err == nil, nil
	}, defaultWaitTimeout, defaultWaitInterval); waitErr != nil {
		return waitErr
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
	// TODO(musaprg): make this configurable via flag
	onlyFilipinoTutor := 1
	// TODO(musaprg): make this configurable via flag
	characteristics := "4"
	return fmt.Sprintf(rarejobTutorSearchURL, from.Local().Year(), from.Local().Month(), from.Local().Day(), s, e, int(onlyFilipinoTutor), characteristics), nil
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

func (c *client) saveCurrentScreenshot(dirPath string, name string) error {
	zap.L().Debug("saving screenshot", zap.String("dir_path", dirPath), zap.String("name", name))
	if c.debug {
		ss, err := c.wd.Screenshot()
		if err != nil {
			return fmt.Errorf("failed to take screenshot: %w", err)
		}
		zap.L().Debug("took screenshot", zap.String("dir_path", dirPath), zap.String("name", name))
		path := filepath.Join(dirPath, name)
		if err := ioutil.WriteFile(path, ss, fs.FileMode(0644)); err != nil {
			return fmt.Errorf("failed to write screenshot: %w", err)
		}
		zap.L().Debug("saved screenshot", zap.String("dir_path", dirPath), zap.String("name", name))
	}
	return nil
}
