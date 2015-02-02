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

		{
			Name:      "find",
			ShortName: "f",
			Usage:     "searches work items that match",
			Action: func(c *cli.Context) {
				find(c.Args()[0])
			},
		},
	}

	app.Run(os.Args)
}

func login() (*rtc.RTC, error) {
	res := rtc.NewRTC("fcoury@br.ibm.com", "tempra14")
	err := res.Login()

	if err != nil {
		return nil, err
	}

	return res, nil
}

func renderTable(wis []*rtc.WorkItem) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Type", "Summary"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColWidth(200)

	for _, wi := range wis {
		table.Append([]string{wi.Id, wi.Type, wi.Summary})
	}
	table.Render()
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

	renderTable(wis)
}

func find(query string) {
	rtc, err := login()
	if err != nil {
		panic(err)
	}

	wis, err := rtc.Search(query)
	if err != nil {
		panic(err)
	}

	renderTable(wis)
}
