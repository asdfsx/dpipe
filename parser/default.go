package parser

import (
	"strings"
)

type DefaultParser struct {
	name string
}

func (this *DefaultParser) parseLine(line string) {
	parts := strings.SplitN(line, ",", 3)
	area, ts, js := parts[0], parts[1], parts[2]
	println(area, ts, js)
}
