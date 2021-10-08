package cmd

import (
	"github.com/r-mironov/draftsman/pkg/draftsman"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sync"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate dependency graph of gitlab-ci config files",
	Run: func(cmd *cobra.Command, args []string) {

		g := &sync.WaitGroup{}
		g.Add(1)

		app := &draftsman.ConfigUrl
		config := &draftsman.AppConfig

		app.Host = viper.GetString("GITLAB_HOST")
		draftsman.ProjectsListUrl.Host = viper.GetString("GITLAB_HOST")
		config.TmpDir = viper.GetString("TMP_DIR")
		config.Token = viper.GetString("GITLAB_TOKEN")

		c := draftsman.IncludeElement{
			Project: viper.GetString("PROJECT_PATH"), //CI_PROJECT_PATH
			Id:      "",
			Ref:     viper.GetString("REF"),            //CI_COMMIT_REF_NAME
			File:    viper.GetString("GITLAB_CI_FILE"), //DEFAULT - .gitlab-ci.yml
		}
		go draftsman.ProjectInclude(c, g)
		g.Wait()

		draftsman.List.Generate()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
