package fasten

import (
	_ "embed"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koalalab-inc/bolt-cli/embeds"
	"github.com/koalalab-inc/bolt-cli/pkg"
	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
)

var fastenHelpTemplate = `
{{.Name}} - {{.Short}}

Usage:
	{{.UseLine}}


Options:
	{{.LocalFlags.FlagUsages | trimRightSpace}}
{{if gt (len .Commands) 0}}
Available Commands:
{{range .Commands}}{{if .IsAvailableCommand}}
	{{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}

Description:
	{{.Long}}
`

var workflowDir = ".github/workflows"

var FastenCmd = &cobra.Command{
	Use:   "fasten",
	Short: "Fasten your workflows with bolt",

	Run: func(cmd *cobra.Command, args []string) {
		choices, err := getWorkflowWithJobs()
		cobra.CheckErr(err)
		if len(choices) == 0 {
			fmt.Println("All jobs are already fastened with bolt.\nHere is the result of the scan:")
			root := cmd.Root()
			root.SetArgs([]string{"scan"})
			root.Execute()
			return
		}
		initialModel := pkg.Model{
			Choices:  choices,
			Selected: make(map[int]struct{}),
		}
		p := tea.NewProgram(initialModel)

		ascii := figlet4go.NewAsciiRender()

		options := figlet4go.NewRenderOptions()
		options.FontName = "colossal"

		ascii.LoadBindataFont(embeds.ColossalFigletFont, "colossal")

		renderStr, _ := ascii.RenderOpts("Bolt-CLI", options)
		fmt.Printf("\n\n\033[96m%s\033[0m", renderStr)

		m, err := p.Run()
		if err != nil {
			cobra.CheckErr(err)
		}
		selectedIndices := m.(pkg.Model).Selected
		selectedJobsByWorkflow := map[string][]string{}
		for i := range selectedIndices {
			choice := choices[i]
			value := choice.Value.(map[string]string)
			workflow := value["workflow"]
			if _, ok := selectedJobsByWorkflow[workflow]; !ok {
				selectedJobsByWorkflow[workflow] = []string{}
			}
			var job string
			if _, ok := value["job"]; ok {
				job = value["job"]
				selectedJobsByWorkflow[workflow] = append(selectedJobsByWorkflow[workflow], job)
			}
		}
		workflowClient := pkg.NewWorkflowClient(workflowDir)

		for workflow, jobs := range selectedJobsByWorkflow {
			err := workflowClient.AddBoltToWorkflow(workflow, jobs)
			cobra.CheckErr(err)
			fmt.Printf(" ✔ Bolted %s with bolt\n", workflow)
		}
	},
}

func init() {
	commands := []*cobra.Command{}
	for _, cmd := range commands {
		cmd.SetHelpTemplate(fastenHelpTemplate)
		FastenCmd.AddCommand(cmd)
	}
	FastenCmd.SetHelpTemplate(fastenHelpTemplate)
}

func getWorkflowWithJobs() ([]*pkg.Choice, error) {
	workflowClient := pkg.NewWorkflowClient(workflowDir)
	workflows, err := workflowClient.GetWorkflows()

	choices := []*pkg.Choice{}
	if err != nil {
		return nil, err
	}

	for _, workflow := range workflows {
		workflowFileName := workflow.FileName
		choice := &pkg.Choice{
			Label: workflowFileName,
			Value: map[string]string{
				"workflow": workflowFileName,
			},
			Parent: nil,
		}
		jobs := workflow.Jobs
		unfastenedJobs := []*pkg.Job{}
		for _, job := range jobs {
			if !job.Fastened {
				unfastenedJobs = append(unfastenedJobs, job)
			}
		}
		unfastenedJobsCount := len(unfastenedJobs)

		if unfastenedJobsCount > 1 {
			choices = append(choices, choice)
			jobChoices := []*pkg.Choice{}
			for _, unfastenedJob := range unfastenedJobs {
				jobChoice := &pkg.Choice{
					Label: unfastenedJob.Name,
					Value: map[string]string{
						"workflow": workflowFileName,
						"job":      unfastenedJob.Name,
					},
					Parent: choice,
				}
				jobChoices = append(jobChoices, jobChoice)
			}
			choices = append(choices, jobChoices...)
		} else if unfastenedJobsCount == 1 {
			choice := &pkg.Choice{
				Label: fmt.Sprintf("%s/%s", workflowFileName, unfastenedJobs[0].Name),
				Value: map[string]string{
					"workflow": workflowFileName,
					"job":      unfastenedJobs[0].Name,
				},
				Parent: nil,
			}
			choices = append(choices, choice)
		}
	}
	return choices, nil
}
