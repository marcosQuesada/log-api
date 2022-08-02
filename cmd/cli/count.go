package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// countCmd represents the count command
var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Count all created log lines",
	Long:  "Count all created log lines",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("count called")
		addr := fmt.Sprintf("localhost:%d", grpcPort)
		conn, err := grpc.Dial(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("client unable to connect, error: %v", err)
		}
		defer conn.Close()

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", fmt.Sprintf("Bearer %s", jwtToken))
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		c := v1.NewLogServiceClient(conn)
		u, err := c.GetLogLineCount(ctx, &emptypb.Empty{})
		if err != nil {
			log.Fatalf("could not get by ID: %v", err)
		}
		log.Printf("User: %v", u)
	},
}

func init() {
	ClientCmd.AddCommand(countCmd)
}
