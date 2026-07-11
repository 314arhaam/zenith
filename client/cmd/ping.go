/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func ping() error {
	endpoint := strings.Join([]string{baseURL, "ping"}, "/")
	//
	resp, err := http.Get(
		endpoint,
	)
	if err != nil {
		// log.Fatalf("Error making GET /ping: %v", err)
		return fmt.Errorf("error making GET /ping: %v", err)
	}
	defer resp.Body.Close()
	//
	if resp.StatusCode != http.StatusOK {
		// log.Fatalf("received status code %d", resp.StatusCode)
		return fmt.Errorf("received status code %d", resp.StatusCode)
	}
	if val, err := io.ReadAll(resp.Body); err != nil {
		// log.Fatalf("response error: %s", err)
		return fmt.Errorf("response error: %s", err)
	} else {
		fmt.Print(string(val))
	}
	return nil
}

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Check availablity of zenith server, either single-shot or continuous",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		to, err := cmd.Flags().GetInt("until")
		if err != nil {
			return fmt.Errorf("error in getting value of flag `--until` or `-u`: %w", err)
		}
		startTime := time.Now()
		if to > 0 {
			attempt := 1
			for {
				err := ping()
				if err == nil {
					return nil
				}
				if time.Since(startTime).Seconds() > float64(to) {
					return fmt.Errorf("Timeout reached without any proper response from server")
				}
				fmt.Printf("(%5ds / %5ds) Attempt: %d passed with error: %v\n", int(time.Since(startTime).Seconds()), to, attempt, err)
				attempt += 1
				time.Sleep(5 * time.Second)
			}
		} else {
			return ping()
		}
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
	pingCmd.Flags().IntP(
		"until",
		"u",
		0,
		"Timeout value. If not pass or pass value zero `0`, it acts as a single shot normal ping command.",
	)
}
