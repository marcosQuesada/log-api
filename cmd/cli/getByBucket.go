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
	bucket string
)

// getByBucketCmd represents the getByPrefix command
var getByBucketCmd = &cobra.Command{
	Use:   "get-by-bucket",
	Short: "get bucket log lines",
	Long:  "get bucket log lines",
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
		u, err := c.GetLogLinesByBucket(ctx, &v1.LogLineByBucketRequest{Bucket: bucket})
		if err != nil {
			log.Fatalf("could not get by bucket %s: %v", prefix, err)
		}
		for _, line := range u.LogLines {
			log.Printf("LogLine with key %s: %v\n", line.GetKey(), line)
		}
	},
}

func init() {
	ClientCmd.AddCommand(getByBucketCmd)
	getByBucketCmd.PersistentFlags().StringVar(&bucket, "bucket", "", "key bucket")
}
