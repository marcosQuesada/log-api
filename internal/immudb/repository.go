package immudb

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/codenotary/immudb/embedded/store"
	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	immuerrors "github.com/codenotary/immudb/pkg/client/errors"
	"github.com/marcosQuesada/log-api/internal/service"
)

// logSizeKeyPlaceHolder defines immudb key to store logLines count
var logSizeKeyPlaceHolder = []byte("log_size")

var errCounterNotInitialized = errors.New("log line size counter not initialized")

type repository struct {
	client client.ImmuClient
}

// NewRepository instantiates new Immudb repository
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

	id := initBinaryCounter()
	if _, err := r.client.Set(context.Background(), logSizeKeyPlaceHolder, id); err != nil {
		return fmt.Errorf("unable to initialize log lines size key %s error %w", logSizeKeyPlaceHolder, err)
	}

	return nil
}

// Add LogLine to repository, if it's a new line it will increment total Log Lines inside the transaction.
// if key already exists it just updates its value
func (r *repository) Add(ctx context.Context, line *service.LogLine) error {
	keySize, err := r.client.Get(context.Background(), logSizeKeyPlaceHolder)
	if err != nil {
		return fmt.Errorf("unable to get log line index %w", err)
	}

	if keySize == nil {
		return errCounterNotInitialized
	}

	sizeValue := incBinaryCounter(keySize.Value)
	_, err = r.client.SetAll(ctx, &schema.SetRequest{
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
	if err := r.addZset(ctx, line.Bucket(), string(line.Key()), line.Time().UnixNano()); err != nil {
		return fmt.Errorf("unexpected error adding zset on key %s error %v", string(line.Key()), err)
	}
	return nil
}

// AddBatch adds a batch of logLines in a unique transaction. Applies same logic from Add LogLines, if all keys are new it will increment total log lines too
// If any of the logLines already exists precondition will fail and to maintain consistency we will process entries one by one as Add does
func (r *repository) AddBatch(ctx context.Context, lines []*service.LogLine) error {
	kv := []*schema.KeyValue{}
	pre := []*schema.Precondition{}
	for _, line := range lines {
		kv = append(kv, &schema.KeyValue{Key: line.Key(), Value: line.Value()})
		pre = append(pre, schema.PreconditionKeyMustNotExist(line.Key()))
	}

	keySize, err := r.client.Get(context.Background(), logSizeKeyPlaceHolder)
	if err != nil {
		return fmt.Errorf("unable to get log line index %w", err)
	}

	if keySize == nil {
		return errors.New("unexpected error key Size is Nil")
	}

	sizeValue := incBinaryCounter(keySize.Value)
	kv = append(kv, &schema.KeyValue{Key: logSizeKeyPlaceHolder, Value: sizeValue})
	pre = append(pre, schema.PreconditionKeyNotModifiedAfterTX(logSizeKeyPlaceHolder, keySize.Tx))
	_, err = r.client.SetAll(ctx, &schema.SetRequest{KVs: kv, Preconditions: pre})

	// On Batch insertion premises failure, try to store lines one by one
	if err != nil && immuerrors.FromError(err) != nil && immuerrors.FromError(err).Code() == immuerrors.CodIntegrityConstraintViolation {
		for _, line := range lines {
			_ = r.Add(ctx, line)
		}

		return nil
	}

	for _, line := range lines {
		if err := r.addZset(ctx, line.Bucket(), string(line.Key()), line.Time().UnixNano()); err != nil {
			return fmt.Errorf("unexpected error adding zset on key %s error %v", string(line.Key()), err)
		}
	}

	if err != nil {
		return fmt.Errorf("unable to BatchLogLines, error %w", err)
	}

	return nil
}

// History returns all revisions from a key
func (r *repository) History(ctx context.Context, key string) (*service.LogLineHistory, error) {
	key = cleanKey([]byte(key))

	h, err := r.client.History(ctx, &schema.HistoryRequest{Key: []byte(key)})
	if err != nil {
		return nil, fmt.Errorf("unable to Get key %s history, error %w", key, err)
	}

	rv := []*service.LogLineRevision{}
	for _, entry := range h.Entries {
		key := cleanKey(entry.GetKey())
		if filterSelfSystemKey(key) {
			continue
		}

		rv = append(rv, &service.LogLineRevision{
			Value:    entry.Value,
			Tx:       entry.Tx,
			Revision: entry.Revision,
		})
	}
	return &service.LogLineHistory{Key: key, Revision: rv}, nil
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

	return binaryCounter(raw.Value), nil
}

// GetByKey returns logLine by Key
func (r *repository) GetByKey(ctx context.Context, key string) (*service.LogLine, error) {
	l, err := r.client.Get(ctx, []byte(key))
	if err != nil {
		return nil, fmt.Errorf("unable to get key %s error %v", key, err)
	}

	return service.NewLogLine(string(l.Key), string(l.Value)), nil
}

// GetByPrefix gets logLines with prefixed key
func (r *repository) GetByPrefix(ctx context.Context, prefix string) ([]*service.LogLine, error) {
	all, err := r.client.Scan(ctx, &schema.ScanRequest{
		Prefix: []byte(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get keys by prefix, error %v", err)
	}

	logs := []*service.LogLine{}
	for _, entry := range all.Entries {
		logs = append(logs, service.NewLogLine(string(entry.Key), string(entry.Value)))
	}

	return logs, nil
}

// GetByBucket gets logLines with bucket
func (r *repository) GetByBucket(ctx context.Context, bucket string) ([]*service.LogLine, error) {
	all, err := r.client.ZScan(ctx, &schema.ZScanRequest{Set: []byte(bucket)})
	if err != nil {
		return nil, fmt.Errorf("unable to get keys by bucket, error %v", err)
	}

	logs := []*service.LogLine{}
	for _, entry := range all.Entries {
		ln, err := r.client.Get(ctx, entry.GetKey())
		if err != nil {
			return nil, fmt.Errorf("unable to get keys by bucket, error %v", err)
		}
		logs = append(logs, service.NewLogLine(string(entry.Key), string(ln.Value)))
	}

	return logs, nil
}

// GetLastNLogLines gets logLines from last N transactions
func (r *repository) GetLastNLogLines(ctx context.Context, n int) ([]*service.LogLine, error) {
	st, err := r.client.CurrentState(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get immudb current state, error %w", err)
	}

	txs, err := r.client.TxScan(ctx, &schema.TxScanRequest{
		InitialTx: st.TxId,
		Limit:     uint32(n),
		Desc:      true,
	})

	logs := []*service.LogLine{}
	for _, tx := range txs.GetTxs() {
		for _, entry := range tx.Entries {
			key := cleanKey(entry.GetKey()) // @TODO: Dirty Key Values here! WHY?
			if filterSelfSystemKey(key) {
				continue
			}

			item, err := r.client.Get(ctx, []byte(key))
			if err != nil {
				return nil, fmt.Errorf("unable to get client, key %s error %v", key, err)
			}
			logs = append(logs, service.NewLogLine(string(item.Key), string(item.Value)))
		}
	}

	return logs, nil
}

func (r *repository) addZset(ctx context.Context, bucket string, key string, score int64) error {
	log.Printf("Add Zset on Key %s bucket %s \n", key, bucket)
	_, err := r.client.ZAdd(ctx, []byte(bucket), float64(score), []byte(key))
	if err != nil {
		return fmt.Errorf("unexpected error adding zset entry, error %v", err)
	}

	return nil
}

// filterSelfSystemKey returns true on our logLines counter key
func filterSelfSystemKey(key string) bool {
	return key == string(logSizeKeyPlaceHolder)
}
