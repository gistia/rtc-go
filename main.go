package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/fcoury/rtc-go/models"
	"github.com/fcoury/rtc-go/rtc"
	"github.com/gistia/tablewriter"
	"github.com/kennygrant/sanitize"
)

var appConfig *Config

type UpdateAttrs struct {
	Estimate  string
	TimeSpent string
	Iteration string
	Start     bool
	Resolve   bool
	Close     bool
	Reopen    bool
}

type Query struct {
	Mine       bool
	Resolved   bool
	Unresolved bool
	Current    bool
	Summary    string
	Parent     string
	Owner      string
	Sort       string
	SortAsc    bool
	MaxResults int
}

func (q Query) Check() error {
	if q.Resolved {
		if q.Unresolved {
			return errors.New("Can't use --closed with --open")
		}
	}

	return nil
}

func (u UpdateAttrs) HasAttributes() bool {
	return u.Estimate != "" && u.TimeSpent != "" && u.Iteration != ""
}

func main() {
	app := cli.NewApp()
	app.Name = "rtc"
	app.Usage = "interact with RTC from the command line"
	app.Author = "Felipe Coury"
	app.Email = "fcoury@br.ibm.com"
	app.Version = "1.0"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Display verbose logs",
		},
	}
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
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "summary",
					Usage: "Omits the description of the work item",
				},
			},
			Action: func(c *cli.Context) {
				info(c.Args()[0], c.Bool("summary"))
			},
		},

		{
			Name:      "show",
			ShortName: "s",
			Usage:     "displays info about a work item",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "summary",
					Usage: "Omits the description of the work item",
				},
			},
			Action: func(c *cli.Context) {
				info(c.Args()[0], c.Bool("summary"))
			},
		},

		{
			Name:      "create",
			ShortName: "c",
			Usage:     "creates a new work item",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type",
					Value: "task",
					Usage: "Type of the work item to be created",
				},
				cli.StringFlag{
					Name:  "parent",
					Value: "",
					Usage: "Id of the parent task, if any",
				},
			},
			Action: func(c *cli.Context) {
				create(c.Args()[0], c.String("type"), c.String("parent"))
			},
		},

		{
			Name:      "update",
			ShortName: "u",
			Usage:     "updates a work item",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "estimate",
					Value: "",
					Usage: "Updates the estimated time for the work item",
				},
				cli.StringFlag{
					Name:  "timespent",
					Value: "",
					Usage: "Updates the time spent working on the work item",
				},
				cli.StringFlag{
					Name:  "iteration",
					Value: "",
					Usage: "Updates the iteration the work item (see iterations command)",
				},
				cli.BoolFlag{
					Name:  "start",
					Usage: "Starts the work item",
				},
				cli.BoolFlag{
					Name:  "resolve",
					Usage: "Resolves the work item",
				},
				cli.BoolFlag{
					Name:  "close",
					Usage: "Closes the work item",
				},
				cli.BoolFlag{
					Name:  "reopen",
					Usage: "Reopens the work item",
				},
			},
			Action: func(c *cli.Context) {
				attrs := UpdateAttrs{
					Estimate:  c.String("estimate"),
					TimeSpent: c.String("timespent"),
					Iteration: c.String("iteration"),
					Start:     c.Bool("start"),
					Resolve:   c.Bool("resolve"),
					Close:     c.Bool("close"),
					Reopen:    c.Bool("reopen"),
				}
				update(c.Args()[0], attrs)
			},
		},

		{
			Name:      "query",
			ShortName: "q",
			Usage:     "queries work items",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "summary",
					Value: "",
					Usage: "Work items that the summary contain the words",
				},
				cli.StringFlag{
					Name:  "parent",
					Value: "",
					Usage: "Work items with the given parent id",
				},
				cli.StringFlag{
					Name:  "owner",
					Value: "",
					Usage: "Filters by the owner name",
				},
				cli.StringFlag{
					Name:  "sort",
					Value: "modified",
					Usage: "Field to sort by",
				},
				cli.IntFlag{
					Name:  "maxresults",
					Value: 15,
					Usage: "How many results to display",
				},
				cli.BoolFlag{
					Name:  "mine",
					Usage: "Everything assigned to me",
				},
				cli.BoolFlag{
					Name:  "closed",
					Usage: "All work items that are closed",
				},
				cli.BoolFlag{
					Name:  "open",
					Usage: "All work items that are open or in progress",
				},
				cli.BoolFlag{
					Name:  "current",
					Usage: "Shows only work items for current iteration",
				},
				cli.BoolFlag{
					Name:  "asc",
					Usage: "Sorts by ascending order (descending is default)",
				},
			},
			Action: func(c *cli.Context) {
				q := Query{
					Mine:       c.Bool("mine"),
					Resolved:   c.Bool("closed"),
					Unresolved: c.Bool("open"),
					Current:    c.Bool("current"),
					Summary:    c.String("summary"),
					Parent:     c.String("parent"),
					Owner:      c.String("owner"),
					Sort:       c.String("sort"),
					SortAsc:    c.Bool("asc"),
					MaxResults: c.Int("maxresults"),
				}
				query(q)
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
			Name:      "subtask",
			ShortName: "st",
			Usage:     "creates a subtask for a story",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type",
					Value: "Artifacts",
					Usage: "Creates a subtask of a given type",
				},
			},
			Action: func(c *cli.Context) {
				createSubtask(c.Args()[0], c.String("type"))
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
			Name:      "show-config",
			ShortName: "shcfg",
			Usage:     "shows the config path",
			Action: func(c *cli.Context) {
				s, err := configFile()
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(s)
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

		{
			Name:      "tree",
			ShortName: "t",
			Usage:     "shows workitem parents and children, if any",
			Action: func(c *cli.Context) {
				tree(c.Args()[0])
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
			Usage:     "show releases",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all",
					Usage: "Shows completed releases",
				},
			},
			Action: func(c *cli.Context) {
				releases(c.Bool("all"))
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
				test()
			},
		},
	}

	c, err := ReadConfig()
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
	if len(wis) < 1 {
		fmt.Println("No work items to be displayed.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Type", "Summary", "Planned For", "Owner", "State"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColWidth(appConfig.MaxWidth)

	for _, wi := range wis {
		table.Append([]string{wi.Id, wi.Type, wi.Summary, wi.PlannedFor, wi.Owner(), wi.State})
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
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Listing the %d open items assigned to you\n\n", len(rwis))

	renderTable(rwis)
}

func find(query string) {
	rtc, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Searching for work items containing \"%s\"...\n", query)

	wis, err := rtc.Search(query)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("")
	renderTable(wis)
}

func showTree(wi *rtc.WorkItem) {
	if len(wi.Parents) > 0 {
		fmt.Println("Parents:\n")
		for _, p := range wi.Parents {
			fmt.Printf("  - %s %s - %s\n", p.Type, p.Id, p.Summary)
		}

		if len(wi.Children) > 0 {
			fmt.Println("")
		}
	}

	if len(wi.Children) > 0 {
		fmt.Println("Children:\n")
		for _, p := range wi.Children {
			fmt.Printf("  - %s %s - %s\n", p.Type, p.Id, p.Summary)
		}
	}
}

func info(id string, summary bool) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wi, err := r.GetWorkItem(id)
	if err != nil {
		fmt.Println(err.Error())
		return
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

	if wi.Estimate != "" {
		fmt.Println("      Estimate:", wi.Estimate)
	}

	if wi.TimeSpent != "" {
		fmt.Println("    Time spent:", wi.TimeSpent)
	}

	if !summary {
		fmt.Println("")
		fmt.Println(strings.Repeat("-", len(title)))
		fmt.Printf("\nDescription:\n\n%s\n", desc)
	}

	if len(wi.Parents) > 0 || len(wi.Children) > 0 {
		fmt.Println("")
		fmt.Println(strings.Repeat("-", len(title)))
		fmt.Println("")
		showTree(wi)
	}

	fmt.Println("")
	// fmt.Printf("Id: %s\nType: %s\nSummary: %s\nPlanned for: %s\nCreated by: %s\nOwner: %s\n\nDescription:\n%s\n",
	// 	wi.Id, wi.Type, wi.Summary, wi.PlannedFor, wi.CreatedBy, wi.OwnedBy, desc)
}

func tree(id string) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wi, err := r.GetWorkItem(id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(wi.Title())
	fmt.Println("")
	showTree(wi)
}

func update(id string, attrs UpdateAttrs) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if attrs.Start {
		fmt.Println("Attempting to start work item %s...\n", id)
		err = r.PerformAction("start", id, "startWorking", "Started")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	if attrs.Resolve {
		fmt.Println("Attempting to resolve work item %s...\n", id)
		err = r.PerformAction("resolve", id, "resolve", "Resolved")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	if attrs.Close {
		fmt.Println("Attempting to close work item %s...\n", id)
		err = r.PerformAction("close", id, "close", "Closed")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	if attrs.Reopen {
		fmt.Println("Attempting to reopen work item %s...\n", id)
		err = r.PerformAction("reopen", id, "reopen", "Reopened")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	if attrs.HasAttributes() {
		wi := rtc.WorkItem{
			Id:          id,
			Estimate:    attrs.Estimate,
			TimeSpent:   attrs.TimeSpent,
			IterationId: attrs.Iteration,
		}

		fmt.Printf("Updating work item %s...\n", id)
		err = r.Update(wi)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	fmt.Println("Work item successfully updated.")
}

func create(summary string, taskType string, parentId string) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wi := &rtc.WorkItem{Summary: summary, Type: taskType}
	fmt.Printf("Creating %s %s...\n", wi.Type, wi.Summary)

	rwi, err := r.Create(wi)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if parentId != "" {
		fmt.Printf("Adding %s as parent...\n", parentId)
		err = r.AddParent(rwi.Id, parentId)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	fmt.Printf("Successfully created: %s\n", rwi.Title())
}

func request(method string, url string) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	body, err := r.Request(method, url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(string(body))
}

func releases(all bool) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	rels, err := r.GetReleases()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Release"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColWidth(appConfig.MaxWidth)

	for i, rel := range rels {
		if !all && rel.Completed == "true" {
			continue
		}
		table.Append([]string{strconv.Itoa(i), rel.Label})
	}
	table.Render()
}

func iterations(all bool) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	iters, err := r.GetIterations()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Iteration"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColWidth(appConfig.MaxWidth)

	for i, iter := range iters {
		if !all && iter.Completed == "true" {
			continue
		}
		table.Append([]string{strconv.Itoa(i), iter.Label})
	}
	table.Render()
}

func test() {
	// r, err := login()
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// r.Query()
}

func close(id string) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
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
		fmt.Println(err.Error())
		return
	}

	ids := strings.Split(id, ",")

	var iter models.Iteration

	for _, i := range ids {
		fmt.Println("Moving work item " + i + "...")
		_, iter, err = r.MoveToIteration(i, iterId)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	fmt.Println("\nSuccessfully moved work items " + id + " to iteration " + iter.Label)
}

func createSubtask(id string, taskType string) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Creating a subtask of type %s...\n", taskType)
	wi, err := r.CreateSubTask(id, taskType)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Created " + wi.Title())

}

func reconfig() {
	err := CreateConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func open(id string) {
	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Opening work item %s in your browser...\n", id)
	err = r.OpenWorkItem(id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func query(q Query) {
	fs := []rtc.Filter{}

	r, err := login()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err = q.Check(); err != nil {
		fmt.Println(err.Error())
		return
	}

	if q.Mine {
		f := rtc.Filter{
			Field:  "owner",
			Oper:   "is",
			Values: []string{appConfig.OwnerId},
		}
		fs = append(fs, f)
	}

	if q.Resolved {
		f := rtc.Filter{
			Field:  "internalState",
			Oper:   "is",
			Values: []string{},
			Vars:   []map[string]string{map[string]string{"id": "state", "arguments": "closed"}},
		}
		fs = append(fs, f)
	}

	if q.Unresolved {
		f := rtc.Filter{
			Field:  "internalState",
			Oper:   "is",
			Values: []string{},
			Vars:   []map[string]string{map[string]string{"id": "state", "arguments": "open or in progress"}},
		}
		fs = append(fs, f)
	}

	if q.Current {
		f := rtc.Filter{
			Field:  "target",
			Oper:   "is",
			Values: []string{},
			Vars:   []map[string]string{map[string]string{"id": "current milestone", "arguments": ""}},
		}
		fs = append(fs, f)
	}

	if q.Summary != "" {
		f := rtc.Filter{
			Field:  "summary",
			Oper:   "contains",
			Values: []string{q.Summary},
		}
		fs = append(fs, f)
	}

	if q.Parent != "" {
		f := rtc.Filter{
			Field:  "link:com.ibm.team.workitem.linktype.parentworkitem:target/id",
			Oper:   "is",
			Values: []string{q.Parent},
		}
		fs = append(fs, f)
	}

	if q.Owner != "" {
		f := rtc.Filter{
			Field:  "owner/name",
			Oper:   "contains",
			Values: []string{q.Owner},
		}
		fs = append(fs, f)
	}

	fmt.Println("Querying RTC for work items that match your query...")
	wis, err := r.Query(fs, q.Sort, q.SortAsc, q.MaxResults)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	renderTable(wis)
}
