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
	"time"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "A brief description of your command",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint := strings.Join([]string{baseURL, "status"}, "/")
		if len(args) == 1 {
			endpoint = endpoint + "?service=" + args[0]
		} else if len(args) > 1 {
			return fmt.Errorf("Too many args")
		}
		retry := 0
		maxRetry, err := cmd.Flags().GetInt("max-retry")
		if err != nil {
			log.Fatalf("Error in flag: %v", err)
			return nil
		}
		dt, err := cmd.Flags().GetInt("interval")
		if err != nil {
			log.Fatalf("Error in flag: %v", err)
			return nil
		}
		sleepDt, err := cmd.Flags().GetInt("sleep")
		if err != nil {
			log.Fatalf("Error in flag: %v", err)
			return nil
		}
		step := 0
		for {
			if retry >= maxRetry {
				log.Printf("[*] Max retry reached, shutdown")
				return nil
			}
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
			f := string(body)
			if f == "{}\n" {
				retry += 1
				fmt.Printf("[*] No data available. Retry %d out of %d\n", retry, maxRetry)
				time.Sleep(time.Duration(dt) * time.Second)
			} else {
				retry = 0
				fmt.Printf("[*] Step %d \n\tData: %v", step, string(body))
				step += 1
				time.Sleep(time.Duration(sleepDt) * time.Second)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	checkCmd.Flags().IntP(
		"max-retry",
		"r",
		3,
		"maximum number of retries if no service(s) found.",
	)
	checkCmd.Flags().IntP(
		"interval",
		"t",
		5,
		"sleep between retries intervals.",
	)
	checkCmd.Flags().IntP(
		"sleep",
		"s",
		5,
		"sleep between polling intervals.",
	)
}
