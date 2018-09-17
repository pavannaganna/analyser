package views

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

// Print renders the ascii table view to stdout
func Print(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Metric", "Value", "Percentage"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
