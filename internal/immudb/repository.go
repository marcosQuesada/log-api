package immudb

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/codenotary/immudb/embedded/store"
	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	immuerrors "github.com/codenotary/immudb/pkg/client/errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/log-api/internal/app"
)

var logSizeKey = []byte("log_size")

type repository struct {
	client client.ImmuClient
}

func NewRepository(c client.ImmuClient) *repository {
	return &repository{client: c}
}

func (r *repository) Initialize(ctx context.Context) error {
	_, err := r.Count(ctx)
	if err == nil {
		return nil
	}
	if !errors.Is(err, errCounterNotInitialized) {
		return fmt.Errorf("unexpected error initializing log line counter, error %w", err)
	}

	var id = make([]byte, 8)
	binary.BigEndian.PutUint64(id, 0)
	if _, err := r.client.Set(context.Background(), logSizeKey, id); err != nil {
		return fmt.Errorf("unable to initialize log lines size key %s error %w", logSizeKey, err)
	}

	return nil
}

//func (r *repository) SetLogLine(ctx context.Context, line *app.LogLine) error { // @TODO Provisional
//	tx, err := r.client.Set(ctx, line.Key(), line.Value())
//	if err != nil {
//		return fmt.Errorf("unable to Set key %s, error %w", line.Key(), err)
//	}
//	spew.Dump(tx)
//	return nil
//}

func (r *repository) Add(ctx context.Context, line *app.LogLine) error {
	keySize, err := r.client.Get(context.Background(), logSizeKey)
	if err != nil {
		return fmt.Errorf("unable to get log line index %w", err)
	}

	if keySize == nil {
		return errors.New("unexpected error key Size is Nil")
	}

	var size = binary.BigEndian.Uint64(keySize.Value)
	size++
	var sizeValue = make([]byte, 8)
	binary.BigEndian.PutUint64(sizeValue, size)

	tx, err := r.client.SetAll(ctx, &schema.SetRequest{
		KVs: []*schema.KeyValue{
			{Key: line.Key(), Value: line.Value()},
			{Key: logSizeKey, Value: sizeValue},
		},
		Preconditions: []*schema.Precondition{
			schema.PreconditionKeyMustNotExist(line.Key()),
			schema.PreconditionKeyNotModifiedAfterTX(logSizeKey, keySize.Tx),
		},
	})

	if err != nil && immuerrors.FromError(err) != nil && immuerrors.FromError(err).Code() == immuerrors.CodIntegrityConstraintViolation {
		if _, err := r.client.Set(context.Background(), line.Key(), line.Value()); err != nil {
			return fmt.Errorf("unable to Update key %s error %w", line.Key(), err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to LogLine key %s, error %w", line.Key(), err)
	}
	spew.Dump(tx)
	return nil
}

func (r *repository) AddBatch(ctx context.Context, lines []*app.LogLine) error {
	// @TODO: Develop in transactional way too

	// on precondition failure process entries one by one
	return nil
}

// @TODO: History, decompose it on ALL and N last items (last N txs)
func (r *repository) History(ctx context.Context, key string) error {
	h, err := r.client.History(ctx, &schema.HistoryRequest{Key: []byte(key)})
	if err != nil {
		return fmt.Errorf("unable to Get key %s history, error %w", key, err)
	}
	spew.Dump(h)

	return nil
}

var errCounterNotInitialized = errors.New("log line size counter not initialized")

func (r *repository) Count(ctx context.Context) (uint64, error) {
	raw, err := r.client.Get(ctx, logSizeKey)
	if err != nil && immuerrors.FromError(err) != nil {
		if errors.Is(immuerrors.FromError(err), store.ErrKeyNotFound) {
			return 0, errCounterNotInitialized
		}
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get logLines size, error: %w", err)
	}

	size := binary.BigEndian.Uint64(raw.Value)
	return size, nil
}
