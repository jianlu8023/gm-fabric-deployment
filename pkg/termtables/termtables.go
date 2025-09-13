package termtables

import (
	"fmt"
	"github.com/scylladb/termtables"
)

func TermTablesExample() {
	table := termtables.CreateTable()

	table.AddHeaders("Name", "Age")
	table.AddRow("John", "30")
	table.AddRow("Sam", 18)
	table.AddRow("Julie", 20.14)

	fmt.Println(table.Render())
}
