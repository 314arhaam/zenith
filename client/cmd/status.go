/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of a service on zenith server",
	Long:  ``,
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint := strings.Join([]string{baseURL, "status"}, "/")
		if len(args) == 1 {
			endpoint = endpoint + "?service=" + args[0]
		} else if len(args) > 1 {
			return fmt.Errorf("Too many args")
		}
		// log.Printf("[*] Calling endpoint %s", endpoint)
		request, err := http.NewRequest(
			http.MethodGet,
			endpoint,
			nil,
		)
		if err != nil {
			return fmt.Errorf("Error in Request create %v", err)
		}
		request.Header["Content-Type"] = []string{"application/json"}
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return fmt.Errorf("Error in GET %s: %v", endpoint, err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error in response parsing %v", err)
		}
		fmt.Printf("%v", string(body))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
