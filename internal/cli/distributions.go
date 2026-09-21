package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/output"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
	"github.com/thedavidweng/qualtrics-cli/internal/safety"
)

var (
	distSurveyID    string
	distPayloadFile string
	distLimit       int
	distOffset      int
	distAll         bool
)

var distributionsCmd = &cobra.Command{
	Use:   "distributions",
	Short: "Manage survey distributions and links",
}

var distributionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List distributions for a survey",
	Run: func(cmd *cobra.Command, args []string) {
		runList(cmd.Context(), "distributions.list", "failed to list distributions",
			func(ctx context.Context, client *qualtrics.Client) ([]qualtrics.Distribution, *output.PaginationMeta, error) {
				if distAll {
					if cmd.Flags().Changed("offset") {
						return nil, nil, errors.New(errors.InvalidArguments, "--all cannot be combined with --offset", errors.CatValidation, false, nil)
					}
					items, err := qualtrics.Paginate(ctx, func(ctx context.Context, opts qualtrics.ListOptions) (qualtrics.Page[qualtrics.Distribution], error) {
						return client.ListDistributions(ctx, distSurveyID, opts)
					}, distLimit)
					return items, &output.PaginationMeta{Limit: len(items), Total: len(items)}, err
				}
				page, err := client.ListDistributions(ctx, distSurveyID, qualtrics.ListOptions{Offset: distOffset, Limit: distLimit})
				if err != nil {
					return nil, nil, err
				}
				return page.Items, &output.PaginationMeta{
					Limit:   distLimit,
					Offset:  distOffset,
					Total:   page.Total,
					HasMore: page.HasMore,
				}, nil
			},
			func(items []qualtrics.Distribution) {
				if fullOutput {
					printJSON(items)
					return
				}
				fmt.Printf("%-20s %-15s %-12s %-20s %s\n", "ID", "SURVEY", "TYPE", "STATUS", "SEND_DATE")
				for _, d := range items {
					fmt.Printf("%-20s %-15s %-12s %-20s %s\n", d.ID, d.SurveyID, d.Type, d.RequestStatus, d.SendDate)
				}
				fmt.Printf("\n%d distributions\n", len(items))
			})
	},
}

var distributionsShowCmd = &cobra.Command{
	Use:   "show <distributionId>",
	Short: "Show a distribution",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "distributions.show", "failed to get distribution",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetDistribution(ctx, args[0], distSurveyID)
			},
			func(data any) { printJSON(data) })
	},
}

var distributionsCreateCmd = &cobra.Command{
	Use:   "create -f <file>",
	Short: "Create a distribution",
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "distributions.create", "failed to create distribution", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(distPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: "",
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateDistribution(ctx, payload)
					},
					human: func(data any) {
						if fullOutput {
							printJSON(data)
							return
						}
						fmt.Println("distribution created")
					},
				}, nil
			})
	},
}

var distributionsDeleteCmd = &cobra.Command{
	Use:   "delete <distributionId>",
	Short: "Delete a distribution",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "distributions.delete", "failed to delete distribution", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[0],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteDistribution(ctx, args[0])
					},
					human: func(data any) { fmt.Printf("deleted %s\n", args[0]) },
				}, nil
			})
	},
}

var distLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "Manage distribution links",
}

var distLinksListCmd = &cobra.Command{
	Use:   "list <distributionId>",
	Short: "List links for a distribution",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runList(cmd.Context(), "distributions.links.list", "failed to list distribution links",
			func(ctx context.Context, client *qualtrics.Client) ([]qualtrics.DistributionLink, *output.PaginationMeta, error) {
				if distAll {
					if cmd.Flags().Changed("offset") {
						return nil, nil, errors.New(errors.InvalidArguments, "--all cannot be combined with --offset", errors.CatValidation, false, nil)
					}
					items, err := qualtrics.Paginate(ctx, func(ctx context.Context, opts qualtrics.ListOptions) (qualtrics.Page[qualtrics.DistributionLink], error) {
						return client.ListDistributionLinks(ctx, args[0], distSurveyID, opts)
					}, distLimit)
					return items, &output.PaginationMeta{Limit: len(items), Total: len(items)}, err
				}
				page, err := client.ListDistributionLinks(ctx, args[0], distSurveyID, qualtrics.ListOptions{Offset: distOffset, Limit: distLimit})
				if err != nil {
					return nil, nil, err
				}
				return page.Items, &output.PaginationMeta{
					Limit:   distLimit,
					Offset:  distOffset,
					Total:   page.Total,
					HasMore: page.HasMore,
				}, nil
			},
			func(items []qualtrics.DistributionLink) {
				if fullOutput {
					printJSON(items)
					return
				}
				fmt.Printf("%-20s %-25s %-12s %s\n", "LINK_ID", "EMAIL", "STATUS", "LINK")
				for _, l := range items {
					fmt.Printf("%-20s %-25s %-12s %s\n", l.LinkID, l.Email, l.Status, l.Link)
				}
				fmt.Printf("\n%d links\n", len(items))
			})
	},
}

var distLinksShowCmd = &cobra.Command{
	Use:   "show <distributionId> <linkId>",
	Short: "Show a distribution link",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "distributions.links.show", "failed to get distribution link",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetDistributionLink(ctx, args[0], args[1])
			},
			func(data any) { printJSON(data) })
	},
}

var distLinksCreateCmd = &cobra.Command{
	Use:   "create <distributionId> -f <file>",
	Short: "Create links for a distribution",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "distributions.links.create", "failed to create distribution links", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(distPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[0],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateDistributionLinks(ctx, args[0], payload)
					},
					human: func(data any) {
						if fullOutput {
							printJSON(data)
							return
						}
						fmt.Println("links created")
					},
				}, nil
			})
	},
}

var distLinksUpdateCmd = &cobra.Command{
	Use:   "update <distributionId> <linkId> -f <file>",
	Short: "Update a distribution link",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "distributions.links.update", "failed to update distribution link", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(distPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[1],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UpdateDistributionLink(ctx, args[0], args[1], payload)
					},
					human: func(data any) { fmt.Printf("updated link %s\n", args[1]) },
				}, nil
			})
	},
}

var distLinksDeleteCmd = &cobra.Command{
	Use:   "delete <distributionId> <linkId>",
	Short: "Delete a distribution link",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "distributions.links.delete", "failed to delete distribution link", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[1],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteDistributionLink(ctx, args[0], args[1])
					},
					human: func(data any) { fmt.Printf("deleted link %s\n", args[1]) },
				}, nil
			})
	},
}

func requireFile(path string) ([]byte, *errors.Error) {
	if path == "" {
		return nil, errors.New(errors.InvalidArguments, "-f <file> is required (use - for stdin)", errors.CatValidation, false, nil)
	}
	payload, err := readPayload(path)
	if err != nil {
		return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
	}
	if !json.Valid(payload) {
		return nil, errors.New(errors.InvalidArguments, "payload is not valid JSON", errors.CatValidation, false, nil)
	}
	return payload, nil
}

func init() {
	distributionsCmd.AddCommand(distributionsListCmd)
	distributionsCmd.AddCommand(distributionsShowCmd)
	distributionsCmd.AddCommand(distributionsCreateCmd)
	distributionsCmd.AddCommand(distributionsDeleteCmd)
	distributionsCmd.AddCommand(distLinksCmd)

	distLinksCmd.AddCommand(distLinksListCmd)
	distLinksCmd.AddCommand(distLinksShowCmd)
	distLinksCmd.AddCommand(distLinksCreateCmd)
	distLinksCmd.AddCommand(distLinksUpdateCmd)
	distLinksCmd.AddCommand(distLinksDeleteCmd)

	distributionsListCmd.Flags().StringVar(&distSurveyID, "survey-id", "", "survey ID filter")
	distributionsListCmd.Flags().IntVar(&distOffset, "offset", 0, "result offset")
	distributionsListCmd.Flags().IntVar(&distLimit, "limit", 100, "page size")
	distributionsListCmd.Flags().BoolVar(&distAll, "all", false, "walk every page")

	distributionsShowCmd.Flags().StringVar(&distSurveyID, "survey-id", "", "survey ID")

	distributionsCreateCmd.Flags().StringVarP(&distPayloadFile, "file", "f", "", "distribution JSON file (- for stdin)")

	distLinksListCmd.Flags().StringVar(&distSurveyID, "survey-id", "", "survey ID")
	distLinksListCmd.Flags().IntVar(&distOffset, "offset", 0, "result offset")
	distLinksListCmd.Flags().IntVar(&distLimit, "limit", 100, "page size")
	distLinksListCmd.Flags().BoolVar(&distAll, "all", false, "walk every page")

	distLinksCreateCmd.Flags().StringVarP(&distPayloadFile, "file", "f", "", "links JSON file (- for stdin)")
	distLinksUpdateCmd.Flags().StringVarP(&distPayloadFile, "file", "f", "", "link JSON file (- for stdin)")
}
