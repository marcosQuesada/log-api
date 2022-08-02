package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/marcosQuesada/log-api/internal/immudb"
	"github.com/marcosQuesada/log-api/internal/jwt"
	"github.com/marcosQuesada/log-api/internal/proto"
	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"github.com/marcosQuesada/log-api/internal/service"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const maxReceivedMessageSize = 1024 * 1024 * 20

var (
	grpcPort  int
	httpPort  int
	jwtSecret string

	immudbUserName string
	immudbPassword string
	immudbDatabase string
	immudbPort     int
	immudbHost     string
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
		defer lis.Close()

		jwtProc := jwt.NewProcessor(jwtSecret)
		auth := proto.NewJWTAuthAdapter(jwtProc)

		cl := buildClient()
		repo := immudb.NewRepository(cl)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := repo.Initialize(ctx); err != nil {
			cancel()
			log.Fatalf("Unable to initialize log lines repository, error %v", err)
		}
		cancel()

		opt := grpc.UnaryInterceptor(auth.Interceptor)
		s := grpc.NewServer(opt)
		svc := service.NewLogService(repo)
		v1.RegisterLogServiceServer(s, svc)
		v1.RegisterAuthServiceServer(s, service.NewAuth(jwtProc, service.NewAuthFakeRepository()))

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
		if err = v1.RegisterLogServiceHandler(context.Background(), mux, conn); err != nil {
			log.Fatalln("Failed to register log service http grpc gateway:", err)
		}

		if err = v1.RegisterAuthServiceHandler(context.Background(), mux, conn); err != nil {
			log.Fatalln("Failed to register auth service http grpc gateway:", err)
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
	serverCmd.PersistentFlags().IntVar(&grpcPort, "grpc-port", 9000, "grpc port")
	serverCmd.PersistentFlags().IntVar(&httpPort, "http-port", 9090, "http grpc gateway port")
	serverCmd.PersistentFlags().StringVar(&jwtSecret, "jwt-secret", "jwt-secret", "jwt secret signature")
	serverCmd.PersistentFlags().StringVar(&immudbUserName, "immudb-user-name", "immudb", "immudb user name")
	serverCmd.PersistentFlags().StringVar(&immudbPassword, "immudb-password", "immudb", "immudb password")
	serverCmd.PersistentFlags().StringVar(&immudbDatabase, "immudb-database", "defaultdb", "immudb database")
	serverCmd.PersistentFlags().StringVar(&immudbHost, "immudb-host", "localhost", "immudb host")
	serverCmd.PersistentFlags().IntVar(&immudbPort, "immudb-port", 3322, "immudb port")

	if p := os.Getenv("jwt-secret"); p != "" {
		jwtSecret = p
	}
	if p := os.Getenv("immudb-user-name"); p != "" {
		immudbUserName = p
	}
	if p := os.Getenv("immudb-password"); p != "" {
		immudbPassword = p
	}
	if p := os.Getenv("immudb-database"); p != "" {
		immudbDatabase = p
	}
	if p := os.Getenv("immudb-host"); p != "" {
		immudbHost = p
	}

	if p := os.Getenv("grpc-port"); p != "" {
		gp, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			log.Fatalf("unable to parse grpc port, got %s error %v", p, err)
		}
		grpcPort = int(gp)
	}
	if p := os.Getenv("http-port"); p != "" {
		hp, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			log.Fatalf("unable to parse http port, got %s error %v", p, err)
		}
		httpPort = int(hp)
	}
	if p := os.Getenv("immudb-port"); p != "" {
		ip, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			log.Fatalf("unable to parse immudb port, got %s error %v", p, err)
		}
		immudbPort = int(ip)
	}
}

func buildClient() client.ImmuClient {
	o := client.DefaultOptions()
	o.Username = immudbUserName
	o.Password = immudbPassword
	o.Database = immudbDatabase
	o.Port = immudbPort
	o.Address = immudbHost

	cl := client.NewClient().WithOptions(o)
	ctx := context.Background()
	if err := cl.OpenSession(ctx, []byte(o.Username), []byte(o.Password), o.Database); err != nil {
		log.Fatalln("Failed to OpenSession on Immudb server, Reason:", err.Error())
	}

	if _, err := cl.UseDatabase(ctx, &schema.Database{DatabaseName: o.Database}); err != nil {
		log.Fatalln("Failed to use the database  on Immudb server, Reason:", err)
	}

	return cl
}
