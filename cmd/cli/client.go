package cli

import (
	"github.com/spf13/cobra"
)

var (
	jwtToken string
	grpcPort int
)

var ClientCmd = &cobra.Command{
	Use:   "client",
	Short: "client command",
	Long:  `client command description`,
}

func init() {
	ClientCmd.PersistentFlags().StringVar(&jwtToken, "token", "", "jwt jwtSecret")
	ClientCmd.PersistentFlags().IntVar(&grpcPort, "grpc-port", 9000, "grpc port")
}
