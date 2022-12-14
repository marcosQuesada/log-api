package service

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Repository interface {
	Add(ctx context.Context, line *LogLine) error
	AddBatch(ctx context.Context, lines []*LogLine) error
	History(ctx context.Context, key string) (*LogLineHistory, error)
	Count(ctx context.Context) (uint64, error)
	GetByKey(ctx context.Context, key string) (*LogLine, error)
	GetByPrefix(ctx context.Context, prefix string) ([]*LogLine, error)
	GetLastNLogLines(ctx context.Context, n int) ([]*LogLine, error)

	GetByBucket(ctx context.Context, bucket string) ([]*LogLine, error)
}

type LogService struct {
	v1.UnimplementedLogServiceServer
	repository Repository
}

func NewLogService(r Repository) *LogService {
	return &LogService{
		repository: r,
	}
}

func (l *LogService) CreateLogLine(ctx context.Context, r *v1.CreateLogLineRequest) (*v1.CreateLogLineResponse, error) {
	log.Printf("Create Log Line %v", r)

	line := convertLogLineRequest(r)
	if err := l.repository.Add(ctx, line); err != nil {
		return nil, status.Error(codes.Internal, "Cannot add LoginLine on repository!")
	}

	return &v1.CreateLogLineResponse{
		Key: line.key,
	}, nil
}

func (l *LogService) BatchCreateLogLines(ctx context.Context, lines *v1.BatchCreateLogLinesRequest) (*v1.BatchCreateLogLinesResponse, error) {
	logs := []*LogLine{}
	ids := []string{}
	for _, r := range lines.Lines {
		line := convertLogLineRequest(r)

		logs = append(logs, line)
		ids = append(ids, line.key)
	}

	if err := l.repository.AddBatch(ctx, logs); err != nil {
		return nil, status.Error(codes.Internal, "Cannot process BatchCreateLogLines on repository!")
	}

	return &v1.BatchCreateLogLinesResponse{
		Key: ids,
	}, nil
}

func (l *LogService) GetAllLogLinesHistory(ctx context.Context, e *emptypb.Empty) (*v1.LogLineHistories, error) {
	all, err := l.repository.GetByPrefix(ctx, "")
	if err != nil {
		return nil, status.Error(codes.Internal, "Cannot process GetByPrefix on repository!")
	}

	return l.histories(ctx, all)
}

func (l *LogService) GetLastNLogLinesHistory(ctx context.Context, e *v1.LastNLogLinesHistoryRequest) (*v1.LogLineHistories, error) {
	all, err := l.repository.GetLastNLogLines(ctx, int(e.GetN()))
	if err != nil {
		return nil, status.Error(codes.Internal, "Cannot process GetByPrefix on repository!")
	}

	return l.histories(ctx, all)
}

func (l *LogService) GetLogLineCount(ctx context.Context, e *emptypb.Empty) (*v1.Count, error) {
	total, err := l.repository.Count(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Cannot count total LoginLine on repository!")
	}
	return &v1.Count{
		Total: total,
	}, nil
}

func (l *LogService) GetLogLineByKey(ctx context.Context, line *v1.LogLineByKeyRequest) (*v1.LogLine, error) {
	ll, err := l.repository.GetByKey(ctx, line.Key)
	if err != nil {
		return nil, status.Error(codes.Internal, "Cannot get by Key on repository!")
	}
	return convertLogLinesToProtocol(ll), nil
}

func (l *LogService) GetLogLinesByPrefix(ctx context.Context, line *v1.LogLineByPrefixRequest) (*v1.LogLines, error) {
	ll, err := l.repository.GetByPrefix(ctx, line.Prefix)
	if err != nil {
		return nil, status.Error(codes.Internal, "Cannot get by Prefix on repository!")
	}

	lines := []*v1.LogLine{}
	for _, logLine := range ll {
		lines = append(lines, convertLogLinesToProtocol(logLine))
	}
	return &v1.LogLines{LogLines: lines}, nil
}

func (l *LogService) GetLogLinesByBucket(ctx context.Context, req *v1.LogLineByBucketRequest) (*v1.LogLines, error) {
	ll, err := l.repository.GetByBucket(ctx, req.GetBucket())
	if err != nil {
		return nil, status.Error(codes.Internal, "Cannot get by Bucket on repository!")
	}

	lines := []*v1.LogLine{}
	for _, logLine := range ll {
		lines = append(lines, convertLogLinesToProtocol(logLine))
	}
	return &v1.LogLines{LogLines: lines}, nil
}

func (l *LogService) histories(ctx context.Context, all []*LogLine) (*v1.LogLineHistories, error) {
	lh := []*v1.LogLineHistory{}
	for _, line := range all {
		h, err := l.repository.History(ctx, string(line.Key()))
		if err != nil {
			return nil, status.Error(codes.Internal, "Cannot get History on repository!")
		}

		r := []*v1.LogLineRevision{}
		for _, i := range h.Revision {
			r = append(r, &v1.LogLineRevision{
				Tx:       int64(i.Tx),
				Value:    string(i.Value),
				Revision: int64(i.Revision),
			})
		}
		lh = append(lh, &v1.LogLineHistory{
			Key:      string(line.Key()),
			Revision: r,
		})
	}

	return &v1.LogLineHistories{
		Histories: lh,
	}, nil
}

func convertLogLineRequest(l *v1.CreateLogLineRequest) *LogLine {
	return &LogLine{
		key:   logLineKey(l.GetSource(), l.CreatedAt.AsTime()),
		value: l.Value,

		bucket: l.GetBucket(), // @TODO: Refactor on next iteration
		time:   l.CreatedAt.AsTime(),
	}

}

func convertLogLinesToProtocol(l *LogLine) *v1.LogLine {
	return &v1.LogLine{
		Key:   l.key,
		Value: l.value,
	}
}

func logLineKey(source string, t time.Time) string {
	return fmt.Sprintf("%s_%d", source, t.UnixNano())
}
