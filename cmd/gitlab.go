package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

type GitlabConfig struct {
	searchNamespace  bool
	searchMembership bool
}

func projectSearchAndOpen(args []string, gitlabConfig GitlabConfig) {
	gitlabClient := getGitlabClient()
	searchString := strings.Join(args, " ")

	listProjectOptions := &gitlab.ListProjectsOptions{
		Search:           gitlab.String(searchString),
		SearchNamespaces: gitlab.Bool(gitlabConfig.searchNamespace),
		Membership:       gitlab.Bool(gitlabConfig.searchMembership),
	}

	projects, _, err := gitlabClient.Projects.ListProjects(listProjectOptions)
	cobra.CheckErr(err)

	var projectNamesWithNamespace, projectNames []string
	projectMap := make(map[string]*gitlab.Project)
	for _, project := range projects {
		projectNamesWithNamespace = append(projectNamesWithNamespace, project.NameWithNamespace)
		projectNames = append(projectNames, project.Name)
		projectMap[project.Name] = project
	}

	prompt := promptui.Select{
		Label: "Select project",
		Items: projectNamesWithNamespace,
	}

	selectedPosition, selectedProjectNameWithNamespace, err := prompt.Run()
	selectedProjectName := projectNames[selectedPosition]
	selectedProjectUrl := projectMap[selectedProjectName].HTTPURLToRepo

	if err != nil {
		fmt.Printf("Prompt for project selection failed %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Opening project", selectedProjectNameWithNamespace)
	if err := exists("xdg-open", "--manual"); err != nil {
		fmt.Printf("Cannot open project in the browser (%v). The url to the project is: %v \n", err, selectedProjectUrl)
	} else {
		bash_cmd := exec.Command("xdg-open", selectedProjectUrl)
		err := bash_cmd.Run()

		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		} else {
			fmt.Printf("Opened project %v in the browser.\n", selectedProjectName)
		}
	}
}

// This function checks if a given application exists on $PATH. The variadic arg argument can
// be used to pass arguments like '--version' to the command.
func exists(app string, arg ...string) error {
	bash_cmd := exec.Command(app, arg...)
	return bash_cmd.Run()
}

func getGitlabClient() *gitlab.Client {
	gitlabUrl := viper.GetString("gitlab_url")
	gitlabToken := viper.GetString("gitlab_token")
	gitlab, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabUrl))
	if err != nil {
		fmt.Printf("Failed to create GitLab client (%v).\n", err)
		os.Exit(1)
	}
	return gitlab
}

func init() {
	gitlabConfig := GitlabConfig{}

	var gitlabCmd = &cobra.Command{
		Use:   "gitlab",
		Short: "Gitlab Command",
		Long:  `This command allows the user different gitlab operations.`,
	}

	var searchCmd = &cobra.Command{
		Use:   "search",
		Short: "Gitlab Search & Open Command",
		Long:  `This command allows the user to search for a project and open it.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectSearchAndOpen(args, gitlabConfig)
		},
	}
	rootCmd.AddCommand(gitlabCmd)
	gitlabCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVarP(&gitlabConfig.searchMembership, "membership", "m", false, "Limit by projects that the current user is a member of.")
	searchCmd.Flags().BoolVarP(&gitlabConfig.searchNamespace, "namespace", "n", true, "Include ancestor namespaces when matching search criteria.")
}
