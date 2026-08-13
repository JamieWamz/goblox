package cli

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/JamieWamz/goblox/internal/storage"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

type options struct {
	dbPath string
	askOne func(survey.Prompt, interface{}, ...survey.AskOpt) error
	ask    func([]*survey.Question, interface{}, ...survey.AskOpt) error
}

// NewRootCommand constructs an isolated command tree. Keeping construction free
// of package globals makes the CLI safe to exercise repeatedly in tests.
func NewRootCommand() *cobra.Command {
	return newRootCommand(&options{
		askOne: survey.AskOne,
		ask:    survey.Ask,
	})
}

func newRootCommand(opts *options) *cobra.Command {
	root := &cobra.Command{
		Use:           "goblox",
		Short:         "A focused task tracker for your terminal",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		Long: `goblox keeps a local task list with priorities, due dates, and status tracking.

Task IDs may be supplied in full or as the unique eight-character prefixes shown
in table output. Use --db to select a different SQLite database.`,
	}

	root.PersistentFlags().StringVar(&opts.dbPath, "db", "./goblox.db", "path to the SQLite database")
	root.AddCommand(
		newAddCommand(opts),
		newArchiveCommand(opts),
		newDeleteCommand(opts),
		newExportCommand(opts),
		newListCommand(opts),
		newShowCommand(opts),
		newUpdateCommand(opts),
	)

	return root
}

func Execute() error {
	return NewRootCommand().Execute()
}

func (o *options) database() (*storage.Database, error) {
	return storage.NewDatabase(o.dbPath)
}
