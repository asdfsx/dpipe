package parser

import (
    json "github.com/bitly/go-simplejson"
    "time"
)

// Parser prototype
type Parser interface {
    ParseLine(line string) (area string, ts uint64, data *json.Json)
    GetStats(duration time.Duration)
}

func NewParsers(parsers []string, chAlarm chan <- Alarm) {
	for _, p := range parsers {
		switch p {
		case "MemcacheFailParser":
			allParsers["MemcacheFailParser"] = newMemcacheFailParser(chAlarm)
		case "ErrorLogParser":
			allParsers["ErrorLogParser"] = newErrorLogParser(chAlarm)
		case "PaymentParser":
			allParsers["PaymentParser"] = newPaymentParser(chAlarm)
		default:
			logger.Println("invalid parser:", p)
		}
	}
}
