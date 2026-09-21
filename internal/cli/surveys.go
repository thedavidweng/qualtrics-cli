package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/config"
	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/output"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
	"github.com/thedavidweng/qualtrics-cli/internal/safety"
)

var (
	surveyName     string
	surveyLanguage string
	surveyCategory string
	listOffset     int
	listLimit      int
	listAll        bool
)

var surveysCmd = &cobra.Command{
	Use:   "surveys",
	Short: "List, inspect, create, and delete surveys",
}

var surveysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List surveys",
	Run: func(cmd *cobra.Command, args []string) {
		runList(cmd.Context(), "surveys.list", "failed to list surveys",
			func(ctx context.Context, client *qualtrics.Client) ([]qualtrics.Survey, *output.PaginationMeta, error) {
				if listAll {
					if cmd.Flags().Changed("offset") {
						return nil, nil, errors.New(errors.InvalidArguments, "--all cannot be combined with --offset", errors.CatValidation, false, nil)
					}
					surveys, err := qualtrics.Paginate(ctx, client.ListSurveys, listLimit)
					return surveys, &output.PaginationMeta{Limit: len(surveys), Total: len(surveys)}, err
				}
				page, err := client.ListSurveys(ctx, qualtrics.ListOptions{Offset: listOffset, Limit: listLimit})
				if err != nil {
					return nil, nil, err
				}
				return page.Items, &output.PaginationMeta{
					Limit:   listLimit,
					Offset:  listOffset,
					Total:   page.Total,
					HasMore: page.HasMore,
				}, nil
			},
			func(surveys []qualtrics.Survey) {
				if fullOutput {
					printJSON(surveys)
					return
				}
				fmt.Printf("%-20s %-40s %s\n", "ID", "NAME", "CREATED")
				for _, s := range surveys {
					fmt.Printf("%-20s %-40s %s\n", s.SurveyID, truncate(s.SurveyName, 40), s.CreationDate)
				}
				fmt.Printf("\n%d surveys\n", len(surveys))
			})
	},
}

var surveysShowCmd = &cobra.Command{
	Use:   "show <surveyId>",
	Short: "Show one survey",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "surveys.show", "failed to get survey",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetSurvey(ctx, args[0])
			},
			func(data any) {
				if fullOutput {
					printJSON(data)
					return
				}
				s, _ := data.(qualtrics.Survey)
				fmt.Printf("id:       %s\n", s.SurveyID)
				fmt.Printf("name:     %s\n", s.SurveyName)
				fmt.Printf("status:   %s\n", s.SurveyStatus)
				fmt.Printf("owner:    %s\n", s.OwnerID)
				fmt.Printf("created:  %s\n", s.CreationDate)
				fmt.Printf("modified: %s\n", s.LastModified)
				fmt.Printf("edit:     https://%s.qualtrics.com/survey-builder/%s/edit\n", currentDatacenter(), s.SurveyID)
				fmt.Printf("preview:  https://%s.qualtrics.com/jfe/preview/%s\n", currentDatacenter(), s.SurveyID)
			})
	},
}

var surveysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new survey",
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "surveys.create", "failed to create survey", safety.TierMutation,
			func() (mutation, *errors.Error) {
				if surveyName == "" {
					return mutation{}, errors.New(errors.InvalidArguments, "--name is required", errors.CatValidation, false, nil)
				}
				req := qualtrics.CreateSurveyRequest{
					SurveyName:      surveyName,
					Language:        surveyLanguage,
					ProjectCategory: surveyCategory,
				}
				return mutation{
					resourceID: "",
					planAfter:  req,
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateSurvey(ctx, req)
					},
					human: func(data any) {
						if fullOutput {
							printJSON(data)
							return
						}
						fmt.Println("survey created")
					},
				}, nil
			})
	},
}

var surveysDeleteCmd = &cobra.Command{
	Use:   "delete <surveyId>",
	Short: "Delete a survey and all of its responses",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "surveys.delete", "failed to delete survey", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID:   args[0],
					typedConfirm: true,
					planAfter:    nil,
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteSurvey(ctx, args[0])
					},
					human: func(data any) { fmt.Printf("deleted %s\n", args[0]) },
				}, nil
			})
	},
}

func currentDatacenter() string {
	cfg, err := config.Load(cfgFile)
	if err != nil || cfg.Active == nil || cfg.Active.Datacenter == "" {
		return "<datacenter>"
	}
	return cfg.Active.Datacenter
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func init() {
	surveysCmd.AddCommand(surveysListCmd)
	surveysCmd.AddCommand(surveysShowCmd)
	surveysCmd.AddCommand(surveysCreateCmd)
	surveysCmd.AddCommand(surveysDeleteCmd)

	surveysListCmd.Flags().IntVar(&listOffset, "offset", 0, "result offset")
	surveysListCmd.Flags().IntVar(&listLimit, "limit", 100, "page size")
	surveysListCmd.Flags().BoolVar(&listAll, "all", false, "walk every page")

	surveysCreateCmd.Flags().StringVar(&surveyName, "name", "", "survey name")
	surveysCreateCmd.Flags().StringVar(&surveyLanguage, "language", "EN", "survey language")
	surveysCreateCmd.Flags().StringVar(&surveyCategory, "project-category", "CORE", "project category")
}
