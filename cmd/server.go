package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/marcosQuesada/log-api/internal/immudb"
	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"github.com/marcosQuesada/log-api/internal/service"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const maxReceivedMessageSize = 1024 * 1024 * 20

var (
	grpcPort int = 9000
	httpPort int = 9090

	immudbUserName = "immudb"
	immudbPassword = "immudb"
	immudbDatabase = "defaultdb"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "api server",
	Long:  `api server`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("API server started, gRPC port %d HTTP gRPC-Gateway %d", grpcPort, httpPort)

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			log.Fatalln("Unable to start grpc listener, error:", err)
		}
		defer lis.Close() // @TODO: Move it to shutdown

		cl := buildClient()
		repo := immudb.NewRepository(cl)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := repo.Initialize(ctx); err != nil {
			cancel()
			log.Fatalf("Unable to initialize log lines repository, error %v", err)
		}
		cancel()

		svc := service.NewLogService(repo)
		s := grpc.NewServer()
		v1.RegisterLogServiceServer(s, svc)

		// @TODO: Signal chan, add ordered shutdown
		go func() {
			if err := s.Serve(lis); err != nil {
				log.Fatalf("error serving %v", err)
			}
		}()

		// Create a client connection to the gRPC server bind to gRPC-Gateway proxy
		conn, err := grpc.DialContext(
			context.Background(),
			fmt.Sprintf("%s:%d", "localhost", grpcPort),
			grpc.WithBlock(),
			grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxReceivedMessageSize), grpc.MaxCallSendMsgSize(maxReceivedMessageSize)),
		)
		if err != nil {
			log.Fatalln("Failed to dial server:", err)
		}

		mux := runtime.NewServeMux()
		err = v1.RegisterLogServiceHandler(context.Background(), mux, conn)
		if err != nil {
			log.Fatalln("Failed to register http grpc gateway:", err)
		}

		gws := &http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%d", httpPort),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		log.Fatalln(gws.ListenAndServe())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func buildClient() client.ImmuClient {
	o := client.DefaultOptions()
	o.Username = immudbUserName
	o.Password = immudbPassword
	o.Database = immudbDatabase
	o.Port = 3322

	cl := client.NewClient().WithOptions(o)
	ctx := context.Background()
	if err := cl.OpenSession(ctx, []byte(o.Username), []byte(o.Password), o.Database); err != nil {
		log.Fatalln("Failed to OpenSession. Reason:", err.Error())
	}

	if _, err := cl.UseDatabase(ctx, &schema.Database{DatabaseName: o.Database}); err != nil {
		log.Fatalln("Failed to use the database. Reason:", err)
	}

	return cl
}
