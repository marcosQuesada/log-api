package app

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

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

type LogService struct {
	v1.UnimplementedLogServiceServer
}

func (l *LogService) CreateLogLine(ctx context.Context, line *v1.LogLine) (*v1.LogLine, error) {
	log.Printf("Create Log Line %v", line)

	return &v1.LogLine{
		Id:        0,
		Bucket:    "FakeBucket",
		Data:      "FakeData",
		CreatedAt: &timestamp.Timestamp{Seconds: time.Now().Unix()},
	}, nil
}

func (l *LogService) CreateBatchLogLine(ctx context.Context, lines *v1.LogLines) (*v1.LogLines, error) {
	log.Printf("CreateBatchLogLine Log Lines %v", lines)

	return &v1.LogLines{
		LogLines: []*v1.LogLine{
			{
				Id:        0,
				Bucket:    "FakeBucket",
				Data:      "FakeData",
				CreatedAt: &timestamp.Timestamp{Seconds: time.Now().Unix()},
			},
			{
				Id:        1,
				Bucket:    "FakeBucket1",
				Data:      "FakeData1",
				CreatedAt: &timestamp.Timestamp{Seconds: time.Now().Unix()},
			},
		},
	}, nil
}

func (l *LogService) GetLogById(ctx context.Context, line *v1.LogLineById) (*v1.LogLine, error) {
	log.Printf("GetLogById Log Line %v", line)

	return &v1.LogLine{
		Id:        99,
		Bucket:    "FakeBucket-XXX",
		Data:      "FakeData-XXX",
		CreatedAt: &timestamp.Timestamp{Seconds: time.Now().Unix()},
	}, nil
}

func (l *LogService) GetAllHistory(ctx context.Context, e *emptypb.Empty) (*v1.LogLines, error) {
	log.Printf("GetAllHistory Log Lines %v", e)
	return &v1.LogLines{
		LogLines: []*v1.LogLine{
			{
				Id:        0,
				Bucket:    "FakeBucket",
				Data:      "FakeData",
				CreatedAt: &timestamp.Timestamp{Seconds: time.Now().Unix()},
			},
			{
				Id:        1,
				Bucket:    "FakeBucket1",
				Data:      "FakeData1",
				CreatedAt: &timestamp.Timestamp{Seconds: time.Now().Unix()},
			},
		},
	}, nil
}

func (l *LogService) GetLogCount(ctx context.Context, lines *v1.LogLines) (*v1.Count, error) {
	log.Printf("GetLogCount Log Lines %v", lines)

	return &v1.Count{
		Total: 199,
	}, nil
}
