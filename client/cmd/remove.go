/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	data "zenith/models"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a service from zenith server",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint := strings.Join([]string{baseURL, "remove"}, "/")
		serviceName := args[0]
		//
		reqStruct := data.RemoveRequest{ServiceName: serviceName}
		body, err := json.Marshal(reqStruct)
		if err != nil {
			// log.Fatalf("Error marshaling request payload: %v", err)
			return fmt.Errorf("error marshaling request payload: %v", err)
		}
		//
		client := &http.Client{}
		req, err := http.NewRequest(
			http.MethodDelete,
			endpoint,
			strings.NewReader(string(body)),
		)
		if err != nil {
			return fmt.Errorf("\n[x] Error in DELETE request creation: %v", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("\n[x] Error in DELETE method: %v", err)
		}
		defer resp.Body.Close()
		//
		if resp.StatusCode != http.StatusNoContent {
			// log.Fatalf("received status code %d", resp.StatusCode)
			return fmt.Errorf("received status code %d", resp.StatusCode)
		}
		fmt.Printf("Status Code: %d", resp.StatusCode)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
