package service

import "time"

type LogLine struct {
	key   string
	value string

	bucket string
	time   time.Time
}

func NewLogLine(key, value string) *LogLine {
	return &LogLine{
		key:   key,
		value: value,
	}
}

func NewLogLineWithBucket(bucket, key, value string, ts time.Time) *LogLine {
	return &LogLine{
		key:    key,
		value:  value,
		bucket: bucket,
		time:   ts,
	}
}

func (l *LogLine) Key() []byte {
	return []byte(l.key)
}

func (l *LogLine) Value() []byte {
	return []byte(l.value)
}

func (l *LogLine) Bucket() string {
	return l.bucket
}

func (l *LogLine) Time() time.Time {
	return l.time
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
