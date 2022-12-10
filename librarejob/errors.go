package librarejob

import "errors"

var (
	ErrSpreadAcrossTwoDays = errors.New("specified duration are spreading across 2 days")
)