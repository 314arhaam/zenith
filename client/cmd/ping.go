/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "A brief description of your command",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint := strings.Join([]string{baseURL, "ping"}, "/")
		//
		resp, err := http.Get(
			endpoint,
		)
		if err != nil {
			log.Fatalf("Error making GET /ping: %v", err)
			return fmt.Errorf("error making GET /ping: %v", err)
		}
		defer resp.Body.Close()
		//
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("received status code %d", resp.StatusCode)
			return fmt.Errorf("received status code %d", resp.StatusCode)
		}
		if val, err := io.ReadAll(resp.Body); err != nil {
			log.Fatalf("response error: %s", err)
			return fmt.Errorf("response error: %s", err)
		} else {
			fmt.Print(string(val))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
