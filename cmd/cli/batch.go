package cli

import (
	"context"
	"encoding/json"
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
	logLinesData string
)

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "batch log lines",
	Long:  "batch log lines",
	Run: func(cmd *cobra.Command, args []string) {
		req := &v1.BatchCreateLogLinesRequest{}
		if err := json.Unmarshal([]byte(logLinesData), req); err != nil {
			log.Fatalf("unable to unmarshall BatchCreateLogLinesRequest, data %s error %v", logLinesData, err)
		}

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
		res, err := c.BatchCreateLogLines(ctx, req)
		if err != nil {
			log.Fatalf("could not batch log lines: %v", err)
		}

		log.Printf("Created Log Lines with keys: %s", res.Key)
	},
}

func init() {
	ClientCmd.AddCommand(batchCmd)
	ClientCmd.PersistentFlags().StringVar(&logLinesData, "lines-data", `{"lines":[{"source":"fake_source_a","bucket":"fake_bucket","value":"fake data value xxx","created_at":{"seconds":1659469108,"nanos":710408961}},{"source":"fake_source_b","bucket":"fake_bucket","value":"fake data value xaxaxax","created_at":{"seconds":1659469108,"nanos":710409242}}]}`, "raw json encoded data")
}
