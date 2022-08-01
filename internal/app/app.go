package app

import (
	"context"
	"log"
	"time"

	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

type Repository interface {
	Add(ctx context.Context, line *LogLine) error
	AddBatch(ctx context.Context, lines []*LogLine) error
	History(ctx context.Context, key string) error
	Count(ctx context.Context) (uint64, error)
}

type LogService struct { // @TODO: Rethink project structure! THis will become an empty layer
	v1.UnimplementedLogServiceServer
	repository Repository
}

func NewLogService(r Repository) *LogService {
	return &LogService{
		repository: r,
	}
}

func (l *LogService) CreateLogLine(ctx context.Context, line *v1.LogLine) (*v1.LogLine, error) {
	log.Printf("Create Log Line %v", line)

	if err := l.repository.Add(ctx, convert(line)); err != nil {
		return nil, status.Error(codes.Internal, "Cannot add LoginLine on repository!")
	}

	return &v1.LogLine{
		Source:    line.Source,
		Bucket:    line.Bucket,
		Data:      line.Data,
		CreatedAt: line.CreatedAt,
	}, nil
}

func (l *LogService) CreateBatchLogLine(ctx context.Context, lines *v1.LogLines) (*v1.LogLines, error) {
	log.Printf("CreateBatchLogLine Log Lines %v", lines)

	logs := []*LogLine{}
	for _, line := range lines.LogLines {
		logs = append(logs, convert(line))
	}

	if err := l.repository.AddBatch(ctx, logs); err != nil {
		return nil, status.Error(codes.Internal, "Cannot add LoginLine on repository!")
	}

	// @TODO: SOlve response
	return &v1.LogLines{
		LogLines: []*v1.LogLine{
			{
				Source:    "foo.log",
				Bucket:    "FakeBucket",
				Data:      "FakeData",
				CreatedAt: timestamppb.New(time.Now()),
			},
			{
				Source:    "foo.log",
				Bucket:    "FakeBucket1",
				Data:      "FakeData1",
				CreatedAt: timestamppb.New(time.Now()),
			},
		},
	}, nil
}

// @TODO: Replace proper types
func (l *LogService) GetAllHistory(ctx context.Context, e *emptypb.Empty) (*v1.LogLines, error) {
	log.Printf("GetAllHistory Log Lines %v", e)

	// @TODO: GET ALL Log Lines, get History from them
	//if err := l.repository.History(ctx, key); err != nil {
	//	return nil, status.Error(codes.Internal, "Cannot add LoginLine on repository!")
	//}

	return &v1.LogLines{
		LogLines: []*v1.LogLine{
			{
				Source:    "foo.log",
				Bucket:    "FakeBucket",
				Data:      "FakeData",
				CreatedAt: timestamppb.New(time.Now()),
			},
			{
				Source:    "foo.log",
				Bucket:    "FakeBucket1",
				Data:      "FakeData1",
				CreatedAt: timestamppb.New(time.Now()),
			},
		},
	}, nil
}

func (l *LogService) GetLogCount(ctx context.Context, lines *v1.LogLines) (*v1.Count, error) {
	log.Printf("GetLogCount Log Lines %v", lines)
	total, err := l.repository.Count(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Cannot count total LoginLine on repository!")
	}
	return &v1.Count{
		Total: uint32(total), // @TODO: Solve it!
	}, nil
}

func (l *LogService) GetLogById(ctx context.Context, line *v1.LogLineById) (*v1.LogLine, error) {
	log.Printf("GetLogById Log Line %v", line)

	//l, err := l.repository.Grt(ctx, line.Key)
	return &v1.LogLine{
		Source:    "foo.log",
		Bucket:    "FakeBucket-XXX",
		Data:      "FakeData-XXX",
		CreatedAt: timestamppb.New(time.Now()),
	}, nil
}

func convert(l *v1.LogLine) *LogLine {
	return &LogLine{ // @TODO: Resolve key composition!
		key:   l.Source,
		tag:   l.Bucket,
		value: l.Data,
		time:  l.CreatedAt.AsTime(),
	}
}
