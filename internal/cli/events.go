package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
	"github.com/thedavidweng/qualtrics-cli/internal/safety"
)

var (
	eventPayloadFile string
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage event subscriptions (webhooks)",
}

var eventSubsCmd = &cobra.Command{
	Use:   "subscriptions",
	Short: "Manage webhook subscriptions",
}

var eventSubsGetCmd = &cobra.Command{
	Use:   "get <subscriptionId>",
	Short: "Get an event subscription",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "events.subscriptions.get", "failed to get subscription",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetEventSubscription(ctx, args[0])
			},
			func(data any) { printJSON(data) })
	},
}

var eventSubsCreateCmd = &cobra.Command{
	Use:   "create -f <file>",
	Short: "Create an event subscription",
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "events.subscriptions.create", "failed to create subscription", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(eventPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: "",
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateEventSubscription(ctx, payload)
					},
					human: func(data any) {
						if fullOutput {
							printJSON(data)
							return
						}
						fmt.Println("event subscription created")
					},
				}, nil
			})
	},
}

var eventSubsDeleteCmd = &cobra.Command{
	Use:   "delete <subscriptionId>",
	Short: "Delete an event subscription",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "events.subscriptions.delete", "failed to delete subscription", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[0],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteEventSubscription(ctx, args[0])
					},
					human: func(data any) { fmt.Printf("deleted subscription %s\n", args[0]) },
				}, nil
			})
	},
}

func init() {
	eventsCmd.AddCommand(eventSubsCmd)
	eventSubsCmd.AddCommand(eventSubsGetCmd)
	eventSubsCmd.AddCommand(eventSubsCreateCmd)
	eventSubsCmd.AddCommand(eventSubsDeleteCmd)

	eventSubsCreateCmd.Flags().StringVarP(&eventPayloadFile, "file", "f", "", "subscription JSON file (- for stdin)")
}
