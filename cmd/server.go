package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/marcosQuesada/log-api/internal/app"
	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	grpcPort int = 9000
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

		s := grpc.NewServer()

		v1.RegisterLogServiceServer(s, &app.LogService{})

		log.Fatalln(s.Serve(lis))
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
