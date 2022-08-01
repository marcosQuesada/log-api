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

func TestAddSingleLogLine(t *testing.T) {
	defer reset()

	r := NewRepository(cl)
	key := "foo_0"
	if err := r.Add(context.Background(), app.NewLogLine("fakeTag", key, "fake value", time.Now())); err != nil {
		log.Fatalf("unable to add , error %v", err)
	}

	keyB := "foo_1"
	if err := r.Add(context.Background(), app.NewLogLine("fakeTag", keyB, "fake value B", time.Now())); err != nil {
		log.Fatalf("unable to add , error %v", err)
	}

	if err := r.History(context.Background(), key); err != nil {
		log.Fatalf("unable to add, error %v", err)
	}
}

func TestLogLinesSize(t *testing.T) {
	defer reset()
	r := NewRepository(cl)
	size, err := r.Count(context.Background())
	if err != nil {
		t.Fatalf("unable to count entries , error %v", err)
	}

	spew.Dump(size)
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
}

func shutdown() {
	log.Println("EXIT")

	_ = listener.Close()
	_ = immudbServer.Stop()

	_ = os.RemoveAll(options.Dir)
	_ = os.Remove(".state-")
}
