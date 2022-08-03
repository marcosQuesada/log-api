package immudb

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/server/servertest"
	"github.com/marcosQuesada/log-api/internal/service"
	"google.golang.org/grpc"
)

var (
	cl           client.ImmuClient
	grpcPort     = 3333
	immudbServer staticServer
	options      *server.Options
	listener     net.Listener
)

type staticServer interface {
	Start() error
	Stop() error
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestInitializeRepositoryItCreatesLogLIneSizeKeyEntry(t *testing.T) {
	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.Initialize(ctx); err != nil {
		t.Fatalf("unable to initialize repository error %v", err)
	}

	v, err := r.Count(ctx)
	if err != nil {
		t.Fatalf("unable to get repository size, error %v", err)
	}

	if expected, got := uint64(0), v; expected != got {
		t.Fatalf("values do not match, expected %d got %d", expected, got)
	}
}

func TestItAddTwoLogLineAsNewAndIncrementsTotalLogLinesCounter(t *testing.T) {
	defer reset()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r := NewRepository(cl)
	_ = r.Add(ctx, service.NewLogLine("foo_0", "fake value"))
	_ = r.Add(ctx, service.NewLogLine("foo_1", "fake value B"))

	v, err := r.Count(ctx)
	if err != nil {
		t.Fatalf("unable to get repository size, error %v", err)
	}

	if expected, got := uint64(2), v; expected != got {
		t.Fatalf("values do not match, expected %d got %d", expected, got)
	}
}

func TestItGetHistoryFromMultipleUpdatedLogLine(t *testing.T) {
	defer reset()
	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "foo_9"
	_ = r.Add(ctx, service.NewLogLine(key, "fake value"))
	_ = r.Add(ctx, service.NewLogLine(key, "fake value X"))
	_ = r.Add(ctx, service.NewLogLine(key, "fake value XX"))
	_ = r.Add(ctx, service.NewLogLine(key, "fake value XXX"))

	h, err := r.History(ctx, key)
	if err != nil {
		log.Fatalf("unable to add, error %v", err)
	}

	if expected, got := 4, len(h.Revision); expected != got {
		t.Errorf("unexpected total Revisions, expected %d got %d", expected, got)
	}
}

func TestItAddMultipleTimesSameSingleLogLineAsNewAndIncrementsTotalLogLinesCounterJustOnce(t *testing.T) {
	defer reset()
	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "foo_0"
	_ = r.Add(context.Background(), service.NewLogLine(key, "fake value"))
	_ = r.Add(ctx, service.NewLogLine(key, "fake value B"))

	v, err := r.Count(ctx)
	if err != nil {
		t.Fatalf("unable to get repository size, error %v", err)
	}

	if expected, got := uint64(1), v; expected != got {
		t.Fatalf("values do not match, expected %d got %d", expected, got)
	}
}

func TestItGetsByKeyPreviousInsertedLogLine(t *testing.T) {
	defer reset()

	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "foo_0"
	value := "fake value"
	_ = r.Add(context.Background(), service.NewLogLine(key, value))

	ll, err := r.GetByKey(ctx, key)
	if err != nil {
		log.Fatalf("unable to get logs by prefix, error %v", err)
	}

	if expected, got := value, string(ll.Value()); expected != got {
		t.Fatalf("values do not match, expected %s got %s", expected, got)
	}
}

func TestItGetsByPrefixPreviousInsertedLogLines(t *testing.T) {
	defer reset()

	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "foo_0"
	_ = r.Add(context.Background(), service.NewLogLine(key, "fake value"))
	keyB := "foo_1"
	_ = r.Add(context.Background(), service.NewLogLine(keyB, "fake value b"))

	prefix := "foo"
	all, err := r.GetByPrefix(ctx, prefix)
	if err != nil {
		log.Fatalf("unable to get logs by prefix, error %v", err)
	}

	if expected, got := 2, len(all); expected != got {
		t.Fatalf("expectation does not match, expected %d got %d", expected, got)
	}

	// @TODO: Validate result composition
}

func TestItGetsLastNInsertedLogLines(t *testing.T) {
	defer reset()

	r := NewRepository(cl)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	ctx := context.Background()
	key := "foo_0"
	_ = r.Add(context.Background(), service.NewLogLine(key, "fake value"))
	keyB := "foo_1"
	_ = r.Add(context.Background(), service.NewLogLine(keyB, "fake value b"))
	keyC := "foo_x"
	_ = r.Add(context.Background(), service.NewLogLine(keyC, "fake value c"))

	size := 3
	all, err := r.GetLastNLogLines(ctx, size)
	if err != nil {
		log.Fatalf("unable to get last N logs, error %v", err)
	}

	if expected, got := size, len(all); expected != got {
		t.Fatalf("expectation does not match, expected %d got %d", expected, got)
	}

	// @TODO: Validate result composition
}

func TestItInsertsMultipleLogLinesInBatch(t *testing.T) {
	defer reset()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r := NewRepository(cl)
	lines := []*service.LogLine{
		service.NewLogLine("bar_00", "fake value"),
		service.NewLogLine("bar_01", "fake value B"),
	}
	if err := r.AddBatch(ctx, lines); err != nil {
		log.Fatalf("unexpected error adding batch, error %v", err)
	}

	all, err := r.GetLastNLogLines(ctx, 1)
	if err != nil {
		log.Fatalf("unable to get last N logs, error %v", err)
	}

	if expected, got := len(lines), len(all); expected != got {
		t.Fatalf("values do not match, expected %d got %d", expected, got)
	}

}

func TestItInsertsOnZsetOnBatchAdditionDevelopmentTest(t *testing.T) { // @TODO: consolidate on next iteration
	defer reset()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r := NewRepository(cl)
	bucket := "fake_bucket_foo"
	lines := []*service.LogLine{
		service.NewLogLineWithBucket(bucket, "bar_00", "fake value", time.Now()),
		service.NewLogLineWithBucket(bucket, "bar_01", "fake value 0", time.Now().Add(time.Nanosecond)),
	}
	if err := r.AddBatch(ctx, lines); err != nil {
		log.Fatalf("unexpected error adding batch, error %v", err)
	}

	res, err := r.GetByBucket(ctx, bucket)
	if err != nil {
		t.Fatalf("unexpected error getting entries by bucket, error %v", err)
	}

	if expected, got := 2, len(res); expected != got {
		t.Fatalf("expectation does not match, expected %d got %d", expected, got)
	}
}

func setup() {
	log.Println("SETUP")
	options = server.DefaultOptions()
	bs := servertest.NewBufconnServer(options)
	_ = bs.Start()
	immudbServer = bs

	var err error
	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %d, error %v", grpcPort, err)
	}
	go bs.GrpcServer.Serve(listener)

	opts := client.DefaultOptions().WithDialOptions(
		[]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()},
	)
	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"
	opts.Port = grpcPort

	cl = client.NewClient().WithOptions(opts)
	ctx := context.Background()
	if err := cl.OpenSession(ctx, []byte(opts.Username), []byte(opts.Password), opts.Database); err != nil {
		log.Fatalln("Failed to OpenSession. Reason:", err.Error())
	}

	if _, err := cl.UseDatabase(ctx, &schema.Database{DatabaseName: opts.Database}); err != nil {
		log.Fatalln("Failed to use the database. Reason:", err)
	}

	r := NewRepository(cl)
	if err := r.Initialize(ctx); err != nil {
		log.Fatalf("unable to initialize repository, error %v", err)
	}
}

func reset() {
	r := NewRepository(cl)
	all, err := r.GetByPrefix(context.Background(), "")
	if err != nil {
		log.Fatalf("unexpected error %v", err)
	}

	keys := [][]byte{}
	for _, line := range all {
		keys = append(keys, line.Key())
	}

	// Soft delete all keys
	_, _ = cl.Delete(context.Background(), &schema.DeleteKeysRequest{
		Keys:    keys,
		SinceTx: 0,
		NoWait:  false,
	})

	// Reset log line counter
	var sizeValue = make([]byte, 8)
	binary.BigEndian.PutUint64(sizeValue, 0)
	_, _ = cl.Set(context.Background(), logSizeKeyPlaceHolder, sizeValue)
}

func shutdown() {
	_ = listener.Close()
	_ = immudbServer.Stop()

	_ = os.RemoveAll(options.Dir)
	_ = os.Remove(".state-")
}
