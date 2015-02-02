package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/fcoury/rtc-go/rtc"
	"github.com/olekukonko/tablewriter"
)

func main() {
	app := cli.NewApp()
	app.Name = "rtc-cli"
	app.Usage = "interact with RTC from the command line"
	app.Commands = []cli.Command{
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "list current user work items",
			Action: func(c *cli.Context) {
				list()
			},
		},
	}

	app.Run(os.Args)
}

func list() {
	rtc := rtc.NewRTC("fcoury@br.ibm.com", "tempra14")
	err := rtc.Login()

	if err != nil {
		panic(err)
	}

	wis, err := rtc.CurrentWorkItems()
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Summary"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColWidth(200)

	for _, wi := range wis {
		table.Append([]string{wi.Id, wi.Summary})
	}
	table.Render()
}
