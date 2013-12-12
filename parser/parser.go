package parser

import (
	"github.com/funkygao/alser/rule"
	"log"
)

type Stopable interface {
	Stop()
}

type Waitable interface {
	Wait()
}

// Parser prototype
type Parser interface {
	ParseLine(line string) (area string, ts uint64, msg string)
	Stopable
	Waitable
}

type AlarmCollector interface {
	CollectAlarms()
}

// Pass through logger
func SetLogger(l *log.Logger) {
	logger = l
}

// Enable/disable debug mode
func SetDebug(d bool) {
	debug = d
}

// Enable verbose or not
func SetVerbose(v bool) {
	verbose = v
}

func SetDryRun(dr bool) {
	dryRun = dr
}

func SetBackground(b bool) {
	background = b
}

// Stop all parsers and they will do their cleanup automatically
func StopAll() {
	for _, parser := range allParsers {
		parser.Stop()
	}
}

func WaitAll() {
	for _, parser := range allParsers {
		parser.Wait()
	}

	close(chParserAlarm)
}

func ParsersCount() int {
	return len(allParsers)
}

func Parsers() map[string]Parser {
	return allParsers
}

func createParser(conf *config.ConfParser, chUpstreamAlarm chan<- Alarm, chDownstreamAlarm chan<- string) Parser {
	mutex.Lock()
	defer mutex.Unlock()

	if conf.Class == "HostLineParser" {
		return newHostLineParser(conf, chUpstreamAlarm, chDownstreamAlarm)
	} else if conf.Class == "RegexCollectorParser" {
		return newRegexCollectorParser(conf, chUpstreamAlarm, chDownstreamAlarm)
	}

	return newJsonCollectorParser(conf, chUpstreamAlarm, chDownstreamAlarm)
}

// pid: only run this single parser id
func InitParsers(pid string, conf *config.Config, chUpstreamAlarm chan<- Alarm) {
	go runSendAlarmsWatchdog(conf)

	indexer = newIndexer(conf)
	go indexer.mainLoop()

	for _, g := range conf.Guards {
		for _, parserId := range g.Parsers {
			if pid != "" && pid != parserId {
				continue
			}

			if _, present := allParsers[parserId]; present {
				continue
			}

			confParser := conf.ParserById(parserId)
			if confParser == nil {
				panic("invalid parser id: " + parserId)
			}

			if debug {
				logger.Printf("create parser[%s] for %s\n", parserId, g.TailLogGlob)
			}

			allParsers[parserId] = createParser(confParser, chUpstreamAlarm, chParserAlarm)
		}
	}
}

// Dispatch a line of log entry to target parser by name
func Dispatch(parserId, line string) {
	if p, present := allParsers[parserId]; present {
		if debug {
			logger.Printf("%s will parse line: %s\n", parserId, line)
		}

		p.ParseLine(line)
	} else {
		logger.Printf("parser[%s] not found\n", parserId)
	}
}
