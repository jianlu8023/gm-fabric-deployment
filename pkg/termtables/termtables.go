package termtables

import (
	"fmt"
	"github.com/jianlu8023/go-tools/v2/pkg/colour"
	"github.com/scylladb/termtables"
)

func TermTablesExample() {
	table := termtables.CreateTable()

	table.AddHeaders(colour.Blue("Name"), colour.Blue("Age"))
	table.AddRow("John", "30")
	table.AddRow("Sam", 18)
	table.AddRow("Julie", 20.14)

	fmt.Println(table.Render())
}
