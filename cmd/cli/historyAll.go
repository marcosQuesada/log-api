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

// historyAllCmd represents the historyAll command
var historyAllCmd = &cobra.Command{
	Use:   "history-all",
	Short: "get all log lines history",
	Long:  `get all log lines history`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := fmt.Sprintf("localhost:%d", grpcPort)
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("client unable to connect, error: %v", err)
		}
		defer conn.Close()

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", fmt.Sprintf("Bearer %s", jwtToken))
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		c := v1.NewLogServiceClient(conn)
		r, err := c.GetAllLogLinesHistory(ctx, &emptypb.Empty{})
		if err != nil {
			log.Fatalf("could not get all history: %v", err)
		}
		for _, i := range r.Histories {
			log.Printf("History Key %s Revision %v", i.Key, i.Revision)
		}
	},
}

func init() {
	ClientCmd.AddCommand(historyAllCmd)
}
