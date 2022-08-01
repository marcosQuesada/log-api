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

var logSizeKeyPlaceHolder = []byte("log_size")

var errCounterNotInitialized = errors.New("log line size counter not initialized")

type repository struct {
	client client.ImmuClient
}

func NewRepository(c client.ImmuClient) *repository {
	return &repository{client: c}
}

// Initialize ensures total number of log lines Key initialization
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
	if _, err := r.client.Set(context.Background(), logSizeKeyPlaceHolder, id); err != nil {
		return fmt.Errorf("unable to initialize log lines size key %s error %w", logSizeKeyPlaceHolder, err)
	}

	return nil
}

// Add LogLine to repository, if it's a new line it will increment total Log Lines inside the transaction.
// if key already exists it just updates its value
func (r *repository) Add(ctx context.Context, line *app.LogLine) error {
	keySize, err := r.client.Get(context.Background(), logSizeKeyPlaceHolder)
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
			{Key: logSizeKeyPlaceHolder, Value: sizeValue},
		},
		Preconditions: []*schema.Precondition{
			schema.PreconditionKeyMustNotExist(line.Key()),
			schema.PreconditionKeyNotModifiedAfterTX(logSizeKeyPlaceHolder, keySize.Tx),
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

// AddBatch adds a batch of logLines in a unique transaction. Applies same logic from Add LogLines, if all keys are new it will increment total log lines too
// If any of the logLines already exists precondition will fail and to maintain consistency we will process entries one by one as Add does
func (r *repository) AddBatch(ctx context.Context, lines []*app.LogLine) error {
	// @TODO: Develop in transactional way too

	// on precondition failure process entries one by one
	return nil
}

// @TODO: History, decompose it on ALL and N last items (last N txs)
// History returns all History from all logLines
func (r *repository) History(ctx context.Context, key string) error {
	h, err := r.client.History(ctx, &schema.HistoryRequest{Key: []byte(key)})
	if err != nil {
		return fmt.Errorf("unable to Get key %s history, error %w", key, err)
	}
	spew.Dump(h)

	return nil
}

// Count returns total log lines, it's reading from total log lines key
func (r *repository) Count(ctx context.Context) (uint64, error) {
	raw, err := r.client.Get(ctx, logSizeKeyPlaceHolder)
	if err != nil && immuerrors.FromError(err) != nil {
		if errors.Is(immuerrors.FromError(err), store.ErrKeyNotFound) {
			return 0, errCounterNotInitialized
		}
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get logLines size, error: %w", err)
	}

	return binary.BigEndian.Uint64(raw.Value), nil
}
