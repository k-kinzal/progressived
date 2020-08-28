package cmd

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	"os"
)

var (
	// aws credentials
	awsRegion          string
	awsProfile         string
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsSession         *session.Session

	rootCmd = &cobra.Command{
		Use:   "progressived",
		Short: "Daemon for progressive delivery",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// aws credentials
			c := &aws.Config{}
			if awsRegion != "" {
				c.Region = &awsRegion
			}
			if awsProfile != "" {
				c.Credentials = credentials.NewSharedCredentials("", awsProfile)
			}
			if awsAccessKeyId != "" && awsSecretAccessKey != "" {
				c.Credentials = credentials.NewStaticCredentials(awsAccessKeyId, awsSecretAccessKey, "")
			}
			if c.Region != nil || c.Credentials != nil {
				sess, err := session.NewSession(c)
				if err != nil {
					return err
				}
				awsSession = sess
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		Version:       GetVersion(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func init() {
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
	rootCmd.PersistentFlags().StringVar(&awsRegion, "aws-region", "", "Using a specific profile from an AWS credential file")
	rootCmd.PersistentFlags().StringVar(&awsProfile, "aws-profile", "", "The AWS region to use. overrides the configuration in config/env")
	rootCmd.PersistentFlags().StringVar(&awsAccessKeyId, "aws-access-key-id", "", "AWS access key ID. overrides the configuration in config/env.")
	rootCmd.PersistentFlags().StringVar(&awsSecretAccessKey, "aws-secret-access-key", "", "AWS secret access key. overrides the configuration in config/env.")
	rootCmd = setFlags(rootCmd)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("progressived: %v", err))
		os.Exit(1)
	}
	return nil
}
