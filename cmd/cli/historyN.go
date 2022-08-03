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
)

var (
	n int
)

// historyNCmd represents the historyN command
var historyNCmd = &cobra.Command{
	Use:   "history-n",
	Short: "get last N transactions log lines history",
	Long:  "get last N transactions log lines history",
	Run: func(cmd *cobra.Command, args []string) {
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
		r, err := c.GetLastNLogLinesHistory(ctx, &v1.LastNLogLinesHistoryRequest{N: int64(n)})
		if err != nil {
			log.Fatalf("could not get all history: %v", err)
		}
		for _, i := range r.Histories {
			log.Printf("History Key %s Revision %v", i.Key, i.Revision)
		}
	},
}

func init() {
	ClientCmd.AddCommand(historyNCmd)
	historyNCmd.PersistentFlags().IntVar(&n, "number", 3, "last N transactions")
}
