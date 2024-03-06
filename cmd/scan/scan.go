package scan

import (
	"fmt"
	"os"

	"github.com/koalalab-inc/bolt-cli/embeds"
	"github.com/koalalab-inc/bolt-cli/pkg"
	"github.com/mbndr/figlet4go"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var scanHelpTemplate = `
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

var ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan your workflows for bolt",
	Long:  `Scan your workflows for bolt`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowClient := pkg.NewWorkflowClient(workflowDir)
		workflows, err := workflowClient.GetWorkflows()
		cobra.CheckErr(err)
		rows := [][]string{}
		for _, workflow := range workflows {
			for _, job := range workflow.Jobs {
				fastened := "✔"
				if !job.Fastened {
					fastened = "✘"
				}
				workflow := fmt.Sprintf("[%s] %s", workflow.Name, workflow.FileName)
				rows = append(rows, []string{workflow, job.Name, fastened})
			}
		}
		cobra.CheckErr(err)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Workflow", "Job", "Fastened Using Bolt"})
		table.SetAutoMergeCellsByColumnIndex([]int{0})
		table.SetRowLine(true)
		for _, row := range rows {
			if row[2] == "✔" {
				table.Rich(row, []tablewriter.Colors{{}, {}, {tablewriter.FgHiGreenColor}})
			} else {
				table.Rich(row, []tablewriter.Colors{{}, {}, {tablewriter.FgHiRedColor}})
			}
		}
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER})

		ascii := figlet4go.NewAsciiRender()

		options := figlet4go.NewRenderOptions()
		options.FontName = "colossal"

		ascii.LoadBindataFont(embeds.ColossalFigletFont, "colossal")

		renderStr, _ := ascii.RenderOpts("Bolt-CLI", options)
		fmt.Printf("\n\n\033[96m%s\033[0m", renderStr)

		fmt.Printf("\033[96m⚡⚡Current status of Bolt instrumentation of workflows in this repository⚡⚡\033[0m\n")
		table.Render()
	},
}

func init() {
	commands := []*cobra.Command{}
	for _, cmd := range commands {
		cmd.SetHelpTemplate(scanHelpTemplate)
		ScanCmd.AddCommand(cmd)
	}
	ScanCmd.SetHelpTemplate(scanHelpTemplate)
}
