package app

import "time"

type LogLine struct {
	key   string
	value string
	tag   string
	time  time.Time
}

func NewLogLine(tag, key, value string, t time.Time) *LogLine {
	return &LogLine{
		key:   key,
		value: value,
		tag:   tag,
		time:  t,
	}
}

func (l *LogLine) Key() []byte {
	return []byte(l.key)
}

func (l *LogLine) Value() []byte {
	return []byte(l.value)
}

func (l *LogLine) Tag() []byte {
	return []byte(l.tag)
}

func (l *LogLine) Time() time.Time {
	return l.time // @TODO: Format t0 string and use them as key
}
