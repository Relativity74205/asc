package cmd

import (
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

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
	Run:   projectSearchAndOpen,
}

func projectSearchAndOpen(cmd *cobra.Command, args []string) {
	gitlabClient := getGitlabClient(cmd)
	searchString := strings.Join(args, " ")
	membership, _ := cmd.Flags().GetBool("membership")
	namespace, _ := cmd.Flags().GetBool("namespace")

	listProjectOptions := &gitlab.ListProjectsOptions{
		Search:           gitlab.String(searchString),
		SearchNamespaces: gitlab.Bool(namespace),
		Membership:       gitlab.Bool(membership),
	}

	projects, _, err := gitlabClient.Projects.ListProjects(listProjectOptions)
	if err != nil {
		cmd.PrintErr(err)
		os.Exit(1)
	}

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
		cmd.PrintErrf("Prompt for project selection failed %v\n", err)
		return
	}

	cmd.Println("Opening project", selectedProjectNameWithNamespace)
	if err := exists("xdg-open", "--manual"); err != nil {
		cmd.Printf("Cannot open project in the browser (%v). The url to the project is: %v \n", err, selectedProjectUrl)
	} else {
		bash_cmd := exec.Command("xdg-open", selectedProjectUrl)
		err := bash_cmd.Run()

		if err != nil {
			cmd.PrintErr(err)
		} else {
			cmd.Printf("Opened project %v in the browser.\n", selectedProjectName)
		}
	}
}

// This function checks if a given application exists on $PATH. The variadic arg argument can
// be used to pass arguments like '--version' to the command.
func exists(app string, arg ...string) error {
	bash_cmd := exec.Command(app, arg...)
	return bash_cmd.Run()
}

func getGitlabClient(cmd *cobra.Command) *gitlab.Client {
	gitlab_token := viper.GetString("gitlab_token")
	gitlab_url := viper.GetString("gitlab_url")
	gitlab, err := gitlab.NewClient(gitlab_token, gitlab.WithBaseURL(gitlab_url))
	if err != nil {
		cmd.PrintErrf("Failed to create GitLab client (%v).\n", err)
		os.Exit(1)
	}
	return gitlab
}

func init() {
	rootCmd.AddCommand(gitlabCmd)
	gitlabCmd.AddCommand(searchCmd)

	gitlabCmd.Flags().BoolP("membership", "m", false, "Limit by projects that the current user is a member of.")
	gitlabCmd.Flags().BoolP("namespace", "n", true, "Include ancestor namespaces when matching search criteria. ")
}
