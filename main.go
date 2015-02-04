package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/fcoury/rtc-go/config"
	"github.com/fcoury/rtc-go/rtc"
	"github.com/kennygrant/sanitize"
	"github.com/olekukonko/tablewriter"
)

var appConfig *config.Config

func main() {
	app := cli.NewApp()
	app.Name = "rtc"
	app.Usage = "interact with RTC from the command line"
	app.Commands = []cli.Command{
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "list current user work items",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type",
					Value: "",
					Usage: "Filters items of a given type",
				},
			},
			Action: func(c *cli.Context) {
				list(c.String("type"))
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

		{
			Name:      "info",
			ShortName: "i",
			Usage:     "displays info about a work item",
			Action: func(c *cli.Context) {
				info(c.Args()[0])
			},
		},

		{
			Name:      "create",
			ShortName: "c",
			Usage:     "creates a new work item",
			Action: func(c *cli.Context) {
				create()
			},
		},

		{
			Name:      "close",
			ShortName: "cl",
			Usage:     "closes a work item",
			Action: func(c *cli.Context) {
				close(c.Args()[0])
			},
		},

		{
			Name:      "move",
			ShortName: "mv",
			Usage:     "moves a work item to a given iteration (see iters)",
			Action: func(c *cli.Context) {
				move(c.Args()[0], c.Args()[1])
			},
		},

		{
			Name:      "artifact",
			ShortName: "art",
			Usage:     "creates an artifact task for a story",
			Action: func(c *cli.Context) {
				createArtifact(c.Args()[0])
			},
		},

		{
			Name:      "config",
			ShortName: "cf",
			Usage:     "reconfigures the app settings",
			Action: func(c *cli.Context) {
				reconfig()
			},
		},

		{
			Name:      "open",
			ShortName: "o",
			Usage:     "opens the work item on the RTC web interface",
			Action: func(c *cli.Context) {
				open(c.Args()[0])
			},
		},

		// dev-related commands

		{
			Name:      "request",
			ShortName: "req",
			Usage:     "performs an URL request",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "method",
					Value: "GET",
					Usage: "Request method",
				},
			},
			Action: func(c *cli.Context) {
				request(c.String("method"), c.Args()[0])
			},
		},

		{
			Name:      "releases",
			ShortName: "rel",
			Usage:     "show all releases",
			Action: func(c *cli.Context) {
				releases()
			},
		},

		{
			Name:      "iterations",
			ShortName: "iter",
			Usage:     "show all iterations",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all",
					Usage: "Shows completed iterations",
				},
			},
			Action: func(c *cli.Context) {
				iterations(c.Bool("all"))
			},
		},

		{
			Name:  "test",
			Usage: "test",
			Action: func(c *cli.Context) {
				test(c.Args()[0])
			},
		},
	}

	c, err := config.ReadConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	appConfig = c

	app.Run(os.Args)
}

func login() (*rtc.RTC, error) {
	res := rtc.NewRTC(appConfig.User, appConfig.Pass, appConfig.OwnerId)
	err := res.Login()

	if err != nil {
		return nil, err
	}

	return res, nil
}

func renderTable(wis []*rtc.WorkItem) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Type", "Summary", "Planned For"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColWidth(120)

	for _, wi := range wis {
		table.Append([]string{wi.Id, wi.Type, wi.Summary, wi.PlannedFor})
	}
	table.Render()
}

func list(wiType string) {
	r, err := login()

	if err != nil {
		panic(err)
	}

	wiType = strings.ToLower(wiType)

	wis, err := r.CurrentWorkItems()
	var rwis []*rtc.WorkItem

	if wiType == "" {
		rwis = wis
	} else {
		for _, wi := range wis {
			if strings.ToLower(wi.Type) == wiType {
				rwis = append(rwis, wi)
			}
		}
	}

	if err != nil {
		panic(err)
	}

	renderTable(rwis)
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

func info(id string) {
	r, err := login()
	if err != nil {
		panic(err)
	}

	wi, err := r.GetWorkItem(id)
	if err != nil {
		panic(err)
	}

	// desc := strings.Replace(wi.Description, "<br/>", "\n", -1)
	desc := sanitize.HTML(wi.Description)
	title := fmt.Sprintf(" %s %s - %s ", wi.Type, wi.Id, wi.Summary)

	fmt.Println(strings.Repeat("-", len(title)))
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", len(title)))
	fmt.Println("")

	fmt.Println("         State:", wi.State, "/", wi.Resolution)
	fmt.Println("   Planned For:", wi.PlannedFor)
	fmt.Println("    Created By:", wi.CreatedBy)
	fmt.Println("         Owner:", wi.OwnedBy)
	fmt.Println("")
	fmt.Println(strings.Repeat("-", len(title)))
	fmt.Printf("\nDescription:\n\n%s\n", desc)

	if len(wi.Parents) > 0 {
		fmt.Println("")
		fmt.Println(strings.Repeat("-", len(title)))
		fmt.Println("\nParents:\n")
		for _, p := range wi.Parents {
			fmt.Printf("  - %s %s - %s\n", p.Type, p.Id, p.Summary)
		}
	}

	fmt.Println("")
	// fmt.Printf("Id: %s\nType: %s\nSummary: %s\nPlanned for: %s\nCreated by: %s\nOwner: %s\n\nDescription:\n%s\n",
	// 	wi.Id, wi.Type, wi.Summary, wi.PlannedFor, wi.CreatedBy, wi.OwnedBy, desc)
}

func create() {
	r, err := login()
	if err != nil {
		panic(err)
	}

	// wi := &rtc.WorkItem{Summary: "DebugOptions", Type: "task"}
	wi, err := r.Retrieve("1281671")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", wi)
}

func request(method string, url string) {
	r, err := login()
	if err != nil {
		panic(err)
	}

	body, err := r.Request(method, url)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

func releases() {
	r, err := login()
	if err != nil {
		panic(err)
	}

	rels, err := r.GetReleases()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", rels)
}

func iterations(all bool) {
	r, err := login()
	if err != nil {
		panic(err)
	}

	iters, err := r.GetIterations()
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Iteration"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColWidth(120)

	for i, iter := range iters {
		if !all && iter.Completed == "true" {
			continue
		}
		table.Append([]string{strconv.Itoa(i), iter.Label})
	}
	table.Render()
}

func test(id string) {
	r, err := login()
	if err != nil {
		panic(err)
	}

	r.GetWorkItem(id)
}

func close(id string) {
	r, err := login()
	if err != nil {
		panic(err)
	}

	err = r.Close(id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Successfully closed work item", id)
}

func move(id string, iterId string) {
	r, err := login()
	if err != nil {
		panic(err)
	}

	_, iter, err := r.MoveToIteration(id, iterId)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Successfully moved work item " + id + " to iteration " + iter.Label)
}

func createArtifact(id string) {
	r, err := login()
	if err != nil {
		panic(err)
	}

	// wi, err := r.CreateSubTask(id, "Artifacts")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// fmt.Println("Created " + wi.Title())
	m, err := r.GetAllValues()
	if err != nil {
		panic(err)
	}

	for k, v := range m {
		fmt.Println("***", k)

		for kk, vv := range v {
			fmt.Printf("%s = %s\n", kk, vv)
		}
	}

}

func reconfig() {
	config.CreateConfig()
}

func open(id string) {
	r, err := login()
	if err != nil {
		panic(err)
	}

	err = r.OpenWorkItem(id)
	if err != nil {
		panic(err)
	}
}
