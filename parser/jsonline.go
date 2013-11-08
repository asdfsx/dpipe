package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
)

type JsonLineParser struct {
	AlsParser
}

func newJsonLineParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *JsonLineParser) {
	this = new(JsonLineParser)
	this.init(conf, chUpstream, chDownstream)
	return
}

func (this *JsonLineParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.AlsParser.ParseLine(line)
	var data *json.Json = this.msgToJson(msg)
	if dryRun {
		return
	}

	args := this.extractRowValues(data)
	if len(args) != this.keysCount() {
		return
	}

	args = append([]interface{}{area}, args...)
	this.colorPrintfLn(this.conf.PrintFormat, args...)
	this.beep()

	return
}
