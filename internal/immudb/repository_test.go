package immudb

import (
	"context"
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
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/log-api/internal/app"
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
	//defer reset()
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
	r := NewRepository(cl)
	key := "foo_0"
	if err := r.Add(context.Background(), app.NewLogLine(key, "fake value")); err != nil {
		log.Fatalf("unable to add , error %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.Initialize(ctx); err != nil {
		t.Fatalf("unable to initialize repository error %v", err)
	}

	keyB := "foo_1"
	if err := r.Add(ctx, app.NewLogLine(keyB, "fake value B")); err != nil {
		log.Fatalf("unable to add , error %v", err)
	}

	v, err := r.Count(ctx)
	if err != nil {
		t.Fatalf("unable to get repository size, error %v", err)
	}

	if expected, got := uint64(2), v; expected != got {
		t.Fatalf("values do not match, expected %d got %d", expected, got)
	}
}

func TestItGetHistoryFromMultipleAddedLogLines(t *testing.T) {
	defer reset()
	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.Initialize(ctx); err != nil {
		t.Fatalf("unable to initialize repository error %v", err)
	}

	key := "foo_9"
	_ = r.Add(ctx, app.NewLogLine(key, "fake value"))
	_ = r.Add(ctx, app.NewLogLine(key, "fake value X"))
	_ = r.Add(ctx, app.NewLogLine(key, "fake value XX"))
	_ = r.Add(ctx, app.NewLogLine(key, "fake value XXX"))

	h, err := r.History(ctx, key)
	if err != nil {
		log.Fatalf("unable to add, error %v", err)
	}

	if expected, got := 4, len(h.Revision); expected != got {
		t.Errorf("unexpected total Revisions")
	}

}

func TestItAddMultipleTimesSameSingleLogLineAsNewAndIncrementsTotalLogLinesCounterJustOnce(t *testing.T) {
	defer reset()
	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.Initialize(ctx); err != nil {
		t.Fatalf("unable to initialize repository error %v", err)
	}

	key := "foo_0"
	if err := r.Add(context.Background(), app.NewLogLine(key, "fake value")); err != nil {
		log.Fatalf("unable to add , error %v", err)
	}

	if err := r.Add(ctx, app.NewLogLine(key, "fake value B")); err != nil {
		log.Fatalf("unable to add , error %v", err)
	}

	v, err := r.Count(ctx)
	if err != nil {
		t.Fatalf("unable to get repository size, error %v", err)
	}

	if expected, got := uint64(1), v; expected != got {
		t.Fatalf("values do not match, expected %d got %d", expected, got)
	}
}

func TestItGetsByIDPreviousInsertedLogLine(t *testing.T) {
	defer reset()
	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.Initialize(ctx); err != nil {
		t.Fatalf("unable to initialize repository error %v", err)
	}
	key := "foo_0"
	value := "fake value"
	_ = r.Add(context.Background(), app.NewLogLine(key, value))

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
	if err := r.Initialize(ctx); err != nil {
		t.Fatalf("unable to initialize repository error %v", err)
	}
	key := "foo_0"
	_ = r.Add(context.Background(), app.NewLogLine(key, "fake value"))
	keyB := "foo_1"
	_ = r.Add(context.Background(), app.NewLogLine(keyB, "fake value b"))

	prefix := "foo"
	all, err := r.GetByPrefix(ctx, prefix)
	if err != nil {
		log.Fatalf("unable to get logs by prefix, error %v", err)
	}

	if expected, got := 2, len(all); expected != got {
		t.Fatalf("expectation does not match, expected %d got %d", expected, got)
	}
}

func TestItGetsLastNInsertedLogLines(t *testing.T) {
	defer reset()
	r := NewRepository(cl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := r.Initialize(ctx); err != nil {
		t.Fatalf("unable to initialize repository error %v", err)
	}
	key := "foo_0"
	_ = r.Add(context.Background(), app.NewLogLine(key, "fake value"))
	keyB := "foo_1"
	_ = r.Add(context.Background(), app.NewLogLine(keyB, "fake value b"))
	keyC := "foo_x"
	_ = r.Add(context.Background(), app.NewLogLine(keyC, "fake value c"))

	size := 3
	all, err := r.GetLastNLogLines(ctx, size)
	if err != nil {
		log.Fatalf("unable to get last N logs, error %v", err)
	}

	if expected, got := size, len(all); expected != got {
		t.Fatalf("expectation does not match, expected %d got %d", expected, got)
	}
}

func setup() {
	log.Println("SETUP")
	options = server.DefaultOptions()
	bs := servertest.NewBufconnServer(options)

	_ = bs.Start()

	var err error
	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	go bs.GrpcServer.Serve(listener)
	immudbServer = bs

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
	log.Println("RESET")
	r := NewRepository(cl)
	all, err := r.GetByPrefix(context.Background(), "")
	if err != nil {
		log.Fatalf("unexpected error %v", err)
	}

	keys := [][]byte{}
	for _, line := range all {
		if string(line.Key()) == string(logSizeKeyPlaceHolder) {
			continue
		}
		keys = append(keys, line.Key())
	}
	if len(keys) == 0 {
		return
	}

	tx, _ := cl.Delete(context.Background(), &schema.DeleteKeysRequest{
		Keys:    keys,
		SinceTx: 0,
		NoWait:  false,
	})

	spew.Dump(tx)
}

func shutdown() {
	log.Println("EXIT")

	_ = listener.Close()
	_ = immudbServer.Stop()

	_ = os.RemoveAll(options.Dir)
	_ = os.Remove(".state-")

	// @TODO: RESET STATE!
}
