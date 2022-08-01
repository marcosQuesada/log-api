package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/marcosQuesada/log-api/internal/app"
	"github.com/marcosQuesada/log-api/internal/immudb"
	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	grpcPort       int = 9000
	immudbUserName     = "immudb"
	immudbPassword     = "immudb"
	immudbDatabase     = "defaultdb"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "api server",
	Long:  `api server`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("API server started, grpc port %d", grpcPort)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			log.Fatalln("Unable to start grpc listener, error:", err)
		}

		cl := buildClient()
		repo := immudb.NewRepository(cl)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := repo.Initialize(ctx); err != nil {
			cancel()
			log.Fatalf("Unable to initialize log lines repository, error %v", err)
		}
		cancel()

		svc := app.NewLogService(repo)
		s := grpc.NewServer()
		v1.RegisterLogServiceServer(s, svc)

		// @TODO: Signal chan, add ordered shutdown
		log.Fatalln(s.Serve(lis))
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
