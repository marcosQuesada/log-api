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
	logLineData string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add single log line",
	Long:  "add single log line",
	Run: func(cmd *cobra.Command, args []string) {
		req := &v1.CreateLogLineRequest{}
		if err := json.Unmarshal([]byte(logLineData), req); err != nil {
			log.Fatalf("unable to unmarshall CreateLogLineRequest, data %s error %v", logLineData, err)
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
		res, err := c.CreateLogLine(ctx, req)
		if err != nil {
			log.Fatalf("could not create log line: %v", err)
		}

		log.Printf("Created Log Line key %s", res.Key)
	},
}

func init() {
	ClientCmd.AddCommand(addCmd)
	ClientCmd.PersistentFlags().StringVar(&logLineData, "line-data", `{"source":"fake_source_a","bucket":"fake_bucket","value":"fake data value xxx","created_at":{"seconds":1659469226,"nanos":165084420}}`, "raw json encoded data")
}
