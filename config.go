package easyworker

import "fmt"

var (
	// use to store last id of supervisor. id is auto_increment.
	taskLastId int
)

/*
Config is shared between EasyTask & EasyStream struct.
It's store options input by user.
*/
type Config struct {
	// fun stores function add by user.
	fun any

	// number of workers (goroutines)
	worker int

	// retry times, if function was failed, worker will try again.
	retry int

	// sleep time before rerun
	retrySleep int
}

/*
Make a configuration holder for EasyTask or EasyStream.
fun: This is func you need to run task.
numWorkers: Number of goroutine you want to run task.
retryTimes: Number of retry if func is failed.

Example:

	fn = func(n int, prefix string) string {
		return log.Sprintf("%s_%d", prefix, n)
	}

	config,_ := NewConfig(fn, 3, 0, 0)
*/
func NewConfig(fun any, numWorkers int, retryTimes int, retrySleep int) (ret Config, err error) {
	if err = verifyFunc(fun); err != nil {
		return
	}

	if numWorkers < 1 {
		err = fmt.Errorf("number of workers is incorrect, %d", numWorkers)
		return
	}

	if retryTimes < 0 {
		err = fmt.Errorf("retryTimes is incorrect, %d", numWorkers)
		return
	}

	if retryTimes > 0 && retrySleep < 0 {
		err = fmt.Errorf("retrySleep is incorrect, %d", numWorkers)
		return
	}

	ret = Config{
		fun:        fun,
		worker:     numWorkers,
		retry:      retryTimes,
		retrySleep: retrySleep,
	}

	return
}

/*
Turn on/off print log to logger.
enable = true, print log.
       = false, no print log.
*/
func EnableLog(enable bool) {
	printLog = enable
}
