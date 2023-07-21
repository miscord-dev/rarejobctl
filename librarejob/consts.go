package librarejob

import "time"

const (
	// rarejobTeacherQuery is the query URL to search available rarejob teachers.
	// query parameter:
	//	page: 検索結果のページ番号
	// 	characteristics: 講師条件（カンマ区切り）
	// 		1 -> フィリピン大学卒業
	//		2 -> 初心者向き
	//		3 -> ビデオレッスン多め
	//		4 -> ビジネス認定講師
	//		5 -> おすすめ講師
	// 	order: 並び替え
	//		1 -> 総合評価順
	//		2 -> 新着順
	//
	// example query URL: 2022/10/9 10:00~10:30で空いていて、ビジネス教材対応の講師。総合評価順。
	//	https://www.rarejob.com/reservation/?year=2022&month=10&day=9&page=1&lessonTime_from=1000&lessonTime_to=1030&characteristics=4&isSaveCookie=1&order=1
	rarejobTutorSearchURL = "https://www.rarejob.com/reservation/?year=%d&month=%d&day=%d&page=1&lessonTime_from=%d&lessonTime_to=%d&freeWord_target=1&characteristics=4&order=1"

	rarejobLoginURL             = "https://www.rarejob.com/account/login/"
	rarejobReservationFinishURL = "https://www.rarejob.com/reservation/reserve/finish/"
)

const (
	loginPageEmailSelector    = "#RJ_LoginForm_email"
	loginPagePasswordSelector = "#RJ_LoginForm_password"
	loginPageFormSelector     = "#rj--login-form"

	tutorListSelector           = ".o-listItem"
	tutorListItemSelector       = ".o-listItem:nth-child(%d)"
	tutorTimeSlotSelector       = ".o-listItem:nth-child(%d) .o-listItem__slot"
	tutorNameSelector           = ".o-listItem:nth-child(%d) .o-listItem__ttl"
	tutorTimeSlotButtonSelector = ".o-listItem:nth-child(%d) .o-listItem__slot:nth-child(%d) > .a-squareBtn"
	tutorReserveButtonSelector  = ".lessonReserve__tutorInfoBtn > div > a"
)

const (
	// defaultWaitInterval is the interval duration for checking conditions, which needs to set a little bit longer than library default to avoid DDoS.
	defaultWaitInterval = time.Millisecond * 500
	// defaultWaitTimeout is the timeout duration for checking conditions.
	defaultWaitTimeout = time.Second * 60
)

const (
	// maxSeleniumHealthCheckBackoffLimit is the timeout duration for checking the health of selenium server
	maxSeleniumHealthCheckBackoffLimit = 5
	// seleniumHealthCheckRetrySecond is the time until the next retry
	seleniumHealthCheckRetrySecond = 10
)

const (
	// rarejobctlTempDir is the temporary directory for rarejobctl to store files for debugging
	rarejobctlTempDir = "/tmp/rarejobctl"
)
