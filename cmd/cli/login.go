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
)

var (
	user     string
	password string
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login obtains JWT token",
	Long:  "login obtains JWT token",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("login called")
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

		c := v1.NewAuthServiceClient(conn)
		u, err := c.Login(ctx, &v1.LoginRequest{Username: user, Password: password})
		if err != nil {
			log.Fatalf("could not get by ID: %v", err)
		}
		log.Printf("User: %v", u)
	},
}

func init() {
	ClientCmd.AddCommand(loginCmd)
	loginCmd.PersistentFlags().StringVar(&user, "user", "fake_user", "login user name")
	loginCmd.PersistentFlags().StringVar(&password, "password", "fake_password", "login password")
}
