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
	dirPayloadFile string
	dirLimit       int
	dirOffset      int
)

var directoriesCmd = &cobra.Command{
	Use:   "directories",
	Short: "Manage XM directories, mailing lists, and contacts",
}

var directoriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List directories",
	Run: func(cmd *cobra.Command, args []string) {
		runList(cmd.Context(), "directories.list", "failed to list directories",
			func(ctx context.Context, client *qualtrics.Client) ([]qualtrics.Directory, *output.PaginationMeta, error) {
				page, err := client.ListDirectories(ctx, qualtrics.ListOptions{Offset: dirOffset, Limit: dirLimit})
				if err != nil {
					return nil, nil, err
				}
				return page.Items, &output.PaginationMeta{
					Limit:   dirLimit,
					Offset:  dirOffset,
					Total:   page.Total,
					HasMore: page.HasMore,
				}, nil
			},
			func(items []qualtrics.Directory) {
				fmt.Printf("%-25s %s\n", "DIRECTORY_ID", "NAME")
				for _, d := range items {
					fmt.Printf("%-25s %s\n", d.DirectoryID, d.Name)
				}
				fmt.Printf("\n%d directories\n", len(items))
			})
	},
}

var directoriesShowCmd = &cobra.Command{
	Use:   "show <directoryId>",
	Short: "Show a directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "directories.show", "failed to get directory",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetDirectory(ctx, args[0])
			},
			func(data any) { printJSON(data) })
	},
}

var directoriesCreateCmd = &cobra.Command{
	Use:   "create -f <file>",
	Short: "Create a directory",
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.create", "failed to create directory", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(dirPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: "",
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateDirectory(ctx, payload)
					},
					human: func() { fmt.Println("directory created") },
				}, nil
			})
	},
}

var directoriesUpdateCmd = &cobra.Command{
	Use:   "update <directoryId> -f <file>",
	Short: "Update a directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.update", "failed to update directory", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(dirPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[0],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UpdateDirectory(ctx, args[0], payload)
					},
					human: func() { fmt.Printf("updated %s\n", args[0]) },
				}, nil
			})
	},
}

var directoriesDeleteCmd = &cobra.Command{
	Use:   "delete <directoryId>",
	Short: "Delete a directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.delete", "failed to delete directory", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[0],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteDirectory(ctx, args[0])
					},
					human: func() { fmt.Printf("deleted %s\n", args[0]) },
				}, nil
			})
	},
}

// MailingLists

var mailinglistsCmd = &cobra.Command{
	Use:   "mailinglists",
	Short: "Manage mailing lists within a directory",
}

var mailinglistsListCmd = &cobra.Command{
	Use:   "list <directoryId>",
	Short: "List mailing lists",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runList(cmd.Context(), "directories.mailinglists.list", "failed to list mailing lists",
			func(ctx context.Context, client *qualtrics.Client) ([]qualtrics.MailingList, *output.PaginationMeta, error) {
				page, err := client.ListMailingLists(ctx, args[0], qualtrics.ListOptions{Offset: dirOffset, Limit: dirLimit})
				if err != nil {
					return nil, nil, err
				}
				return page.Items, &output.PaginationMeta{
					Limit:   dirLimit,
					Offset:  dirOffset,
					Total:   page.Total,
					HasMore: page.HasMore,
				}, nil
			},
			func(items []qualtrics.MailingList) {
				fmt.Printf("%-20s %s\n", "MAILING_LIST_ID", "NAME")
				for _, m := range items {
					fmt.Printf("%-20s %s\n", m.ID, m.Name)
				}
				fmt.Printf("\n%d mailing lists\n", len(items))
			})
	},
}

var mailinglistsShowCmd = &cobra.Command{
	Use:   "show <directoryId> <mailingListId>",
	Short: "Show a mailing list",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "directories.mailinglists.show", "failed to get mailing list",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetMailingList(ctx, args[0], args[1])
			},
			func(data any) { printJSON(data) })
	},
}

var mailinglistsCreateCmd = &cobra.Command{
	Use:   "create <directoryId> -f <file>",
	Short: "Create a mailing list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.mailinglists.create", "failed to create mailing list", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(dirPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[0],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateMailingList(ctx, args[0], payload)
					},
					human: func() { fmt.Println("mailing list created") },
				}, nil
			})
	},
}

var mailinglistsUpdateCmd = &cobra.Command{
	Use:   "update <directoryId> <mailingListId> -f <file>",
	Short: "Update a mailing list",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.mailinglists.update", "failed to update mailing list", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(dirPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[1],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UpdateMailingList(ctx, args[0], args[1], payload)
					},
					human: func() { fmt.Printf("updated %s\n", args[1]) },
				}, nil
			})
	},
}

var mailinglistsDeleteCmd = &cobra.Command{
	Use:   "delete <directoryId> <mailingListId>",
	Short: "Delete a mailing list",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.mailinglists.delete", "failed to delete mailing list", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[1],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteMailingList(ctx, args[0], args[1])
					},
					human: func() { fmt.Printf("deleted %s\n", args[1]) },
				}, nil
			})
	},
}

// Contacts

var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "Manage contacts in a mailing list",
}

var contactsListCmd = &cobra.Command{
	Use:   "list <directoryId> <mailingListId>",
	Short: "List contacts",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runList(cmd.Context(), "directories.mailinglists.contacts.list", "failed to list contacts",
			func(ctx context.Context, client *qualtrics.Client) ([]qualtrics.Contact, *output.PaginationMeta, error) {
				page, err := client.ListContacts(ctx, args[0], args[1], qualtrics.ListOptions{Offset: dirOffset, Limit: dirLimit})
				if err != nil {
					return nil, nil, err
				}
				return page.Items, &output.PaginationMeta{
					Limit:   dirLimit,
					Offset:  dirOffset,
					Total:   page.Total,
					HasMore: page.HasMore,
				}, nil
			},
			func(items []qualtrics.Contact) {
				fmt.Printf("%-20s %-15s %-15s %-30s %s\n", "CONTACT_ID", "FIRST", "LAST", "EMAIL", "UNSUB")
				for _, c := range items {
					fmt.Printf("%-20s %-15s %-15s %-30s %v\n", c.ContactID, c.FirstName, c.LastName, c.Email, c.Unsubscribed)
				}
				fmt.Printf("\n%d contacts\n", len(items))
			})
	},
}

var contactsShowCmd = &cobra.Command{
	Use:   "show <directoryId> <mailingListId> <contactId>",
	Short: "Show a contact",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "directories.mailinglists.contacts.show", "failed to get contact",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetContact(ctx, args[0], args[1], args[2])
			},
			func(data any) { printJSON(data) })
	},
}

var contactsCreateCmd = &cobra.Command{
	Use:   "create <directoryId> <mailingListId> -f <file>",
	Short: "Create a contact",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.mailinglists.contacts.create", "failed to create contact", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(dirPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[1],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateContact(ctx, args[0], args[1], payload)
					},
					human: func() { fmt.Println("contact created") },
				}, nil
			})
	},
}

var contactsUpdateCmd = &cobra.Command{
	Use:   "update <directoryId> <mailingListId> <contactId> -f <file>",
	Short: "Update a contact",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.mailinglists.contacts.update", "failed to update contact", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requireFile(dirPayloadFile)
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[2],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UpdateContact(ctx, args[0], args[1], args[2], payload)
					},
					human: func() { fmt.Printf("updated %s\n", args[2]) },
				}, nil
			})
	},
}

var contactsDeleteCmd = &cobra.Command{
	Use:   "delete <directoryId> <mailingListId> <contactId>",
	Short: "Delete a contact",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "directories.mailinglists.contacts.delete", "failed to delete contact", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[2],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteContact(ctx, args[0], args[1], args[2])
					},
					human: func() { fmt.Printf("deleted %s\n", args[2]) },
				}, nil
			})
	},
}

func init() {
	directoriesCmd.AddCommand(directoriesListCmd)
	directoriesCmd.AddCommand(directoriesShowCmd)
	directoriesCmd.AddCommand(directoriesCreateCmd)
	directoriesCmd.AddCommand(directoriesUpdateCmd)
	directoriesCmd.AddCommand(directoriesDeleteCmd)
	directoriesCmd.AddCommand(mailinglistsCmd)

	mailinglistsCmd.AddCommand(mailinglistsListCmd)
	mailinglistsCmd.AddCommand(mailinglistsShowCmd)
	mailinglistsCmd.AddCommand(mailinglistsCreateCmd)
	mailinglistsCmd.AddCommand(mailinglistsUpdateCmd)
	mailinglistsCmd.AddCommand(mailinglistsDeleteCmd)
	mailinglistsCmd.AddCommand(contactsCmd)

	contactsCmd.AddCommand(contactsListCmd)
	contactsCmd.AddCommand(contactsShowCmd)
	contactsCmd.AddCommand(contactsCreateCmd)
	contactsCmd.AddCommand(contactsUpdateCmd)
	contactsCmd.AddCommand(contactsDeleteCmd)

	directoriesListCmd.Flags().IntVar(&dirOffset, "offset", 0, "result offset")
	directoriesListCmd.Flags().IntVar(&dirLimit, "limit", 100, "page size")
	directoriesCreateCmd.Flags().StringVarP(&dirPayloadFile, "file", "f", "", "directory JSON file (- for stdin)")
	directoriesUpdateCmd.Flags().StringVarP(&dirPayloadFile, "file", "f", "", "directory JSON file (- for stdin)")

	mailinglistsListCmd.Flags().IntVar(&dirOffset, "offset", 0, "result offset")
	mailinglistsListCmd.Flags().IntVar(&dirLimit, "limit", 100, "page size")
	mailinglistsCreateCmd.Flags().StringVarP(&dirPayloadFile, "file", "f", "", "mailing list JSON file (- for stdin)")
	mailinglistsUpdateCmd.Flags().StringVarP(&dirPayloadFile, "file", "f", "", "mailing list JSON file (- for stdin)")

	contactsListCmd.Flags().IntVar(&dirOffset, "offset", 0, "result offset")
	contactsListCmd.Flags().IntVar(&dirLimit, "limit", 100, "page size")
	contactsCreateCmd.Flags().StringVarP(&dirPayloadFile, "file", "f", "", "contact JSON file (- for stdin)")
	contactsUpdateCmd.Flags().StringVarP(&dirPayloadFile, "file", "f", "", "contact JSON file (- for stdin)")
}
