package service

type LogLine struct {
	key   string
	value string
}

func NewLogLine(key, value string) *LogLine {
	return &LogLine{
		key:   key,
		value: value,
	}
}

func (l *LogLine) Key() []byte {
	return []byte(l.key)
}

func (l *LogLine) Value() []byte {
	return []byte(l.value)
}

type LogLineHistory struct {
	Key      string
	Revision []*LogLineRevision
}

type LogLineRevision struct {
	Value    []byte
	Tx       uint64
	Revision uint64
}
