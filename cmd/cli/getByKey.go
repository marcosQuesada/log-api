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
	key string
)

// getByKeyCmd represents the getByKey command
var getByKeyCmd = &cobra.Command{
	Use:   "get-by-key",
	Short: "Get By log Line Key",
	Long:  "Get By log Line Key",
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

		u, err := c.GetLogLineByKey(ctx, &v1.LogLineByKeyRequest{Key: key})
		if err != nil {
			log.Fatalf("could not get by ID: %v", err)
		}
		log.Printf("User: %v", u)
	},
}

func init() {
	ClientCmd.AddCommand(getByKeyCmd)
	getByKeyCmd.PersistentFlags().StringVar(&key, "key", "", "key name")
}
