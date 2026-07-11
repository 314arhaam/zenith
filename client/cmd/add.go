/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	data "zenith/models"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a service to zenith server",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint := strings.Join([]string{baseURL, "add"}, "/")
		serviceName := args[0]
		//
		req := data.AddRequest{ServiceName: serviceName}
		body, err := json.Marshal(req)
		if err != nil {
			// log.Fatalf("Error marshaling request payload: %v", err)
			return fmt.Errorf("error marshaling request payload: %v", err)
		}
		//
		resp, err := http.Post(
			endpoint,
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			// log.Fatalf("Error making POST request: %v, payload: %s", err, body)
			return fmt.Errorf("error making POST request: %v", err)
		}
		defer resp.Body.Close()
		//
		if resp.StatusCode != http.StatusCreated {
			// log.Fatalf("received status code %d", resp.StatusCode)
			return fmt.Errorf("received status code %d", resp.StatusCode)
		}
		fmt.Printf("Status Code: %d", resp.StatusCode)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
