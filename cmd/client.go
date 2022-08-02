package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "client command",
	Long:  `client command description`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("client called, arguments %v \n", args)

		if len(args) != 3 {
			log.Fatalf("unexpected total arguments, expected: source, bucket, data, got %v", args)
		}

		addr := fmt.Sprintf("localhost:%d", grpcPort)
		conn, err := grpc.Dial(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("client unable to connect, error: %v", err)
		}
		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		c := v1.NewLogServiceClient(conn)
		res, err := c.CreateLogLine(ctx, &v1.CreateLogLineRequest{
			Source:    strings.Trim(args[0], " "),
			Bucket:    strings.Trim(args[1], " "),
			Value:     strings.Trim(args[2], " "),
			CreatedAt: timestamppb.New(time.Now()),
		})
		if err != nil {
			log.Fatalf("could not create log line: %v", err)
		}

		log.Printf("Created Log Line key %s", res.Key)

		r, err := c.GetLastNLogLinesHistory(ctx, &v1.LastNLogLinesHistoryRequest{N: 1})
		if err != nil {
			log.Fatalf("could not get all history: %v", err)
		}
		for _, i := range r.Histories {
			log.Printf("History Key %s Revision %v", i.Key, i.Revision)

		}

		//u, err := c.GetLogById(ctx, &v1.LogLineById{Id: 12112})
		//if err != nil {
		//	log.Fatalf("could not get by ID: %v", err)
		//}
		//log.Printf("User: %v", u)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
