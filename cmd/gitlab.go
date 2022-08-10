package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

var gitlabOpenCmd = &cobra.Command{
	Use:   "glo",
	Short: "Gitlab Open Project Command",
	Long:  `This command allows the user to search for projects and open a project in the browser.`,
	Args:  cobra.ExactArgs(1),
	Run:   projectSearchAndOpen,
}

func projectSearchAndOpen(cmd *cobra.Command, args []string) {
	gitlab_token := viper.GetString("gitlab_token")
	gitlab_url := viper.GetString("gitlab_url")
	git, err := gitlab.NewClient(gitlab_token, gitlab.WithBaseURL(gitlab_url))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	opt := &gitlab.ListProjectsOptions{Search: gitlab.String(args[0])}
	projects, _, err := git.Projects.ListProjects(opt)
	if err != nil {
		log.Fatal(err)
	}

	var project_names []string
	project_map := make(map[string]*gitlab.Project)
	for _, project := range projects {
		project_names = append(project_names, project.Name)
		project_map[project.Name] = project
	}

	prompt := promptui.Select{
		Label: "Select project",
		Items: project_names,
	}

	_, selectedProject, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Println("Opening project", selectedProject)
	bash_cmd := exec.Command("xdg-open", project_map[selectedProject].HTTPURLToRepo)
	stdout, err := bash_cmd.Output()

	fmt.Println(string(stdout))
}

func init() {
	rootCmd.AddCommand(gitlabOpenCmd)
}
