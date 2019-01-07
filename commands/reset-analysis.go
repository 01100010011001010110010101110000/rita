package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/activecm/rita/resources"
	"github.com/urfave/cli"
)

func init() {
	reset := cli.Command{
		Name:      "reset",
		Aliases:   []string{"reset-analysis"},
		Usage:     "Reset analysis of a database",
		ArgsUsage: "<database>",
		Flags: []cli.Flag{
			forceFlag,
			allFlag,
			configFlag,
		},
		Action: func(c *cli.Context) error {
			res := resources.InitResources(c.String("config"))
			db := c.Args().Get(0)
			if db == "" {
				return cli.NewExitError("Specify a database or use -a flag for all databases", -1)
			}

			return resetAnalysis(db, res, c.Bool("force"))
		},
	}

	bootstrapCommands(reset)
}

// // resetAnalysis cleans out all of the analysis data, leaving behind only the
// // raw data from parsing the logs
// func resetAnalysisAll(database string, res *resources.Resources, forceFlag bool) error {
// 	//clean database
//
// 	conn := res.Config.T.Structure.ConnTable
// 	http := res.Config.T.Structure.HTTPTable
// 	dns := res.Config.T.Structure.DNSTable
//
// 	names, err := res.DB.Session.DB(database).CollectionNames()
// 	if err != nil || len(names) == 0 {
// 		return cli.NewExitError("Failed to find analysis results", -1)
// 	}
//
// }

// resetAnalysis cleans out all of the analysis data, leaving behind only the
// raw data from parsing the logs
func resetAnalysis(database string, res *resources.Resources, forceFlag bool) error {
	//clean database

	// The following collections are created at import and cannot be reset:
	conn := res.Config.T.Structure.ConnTable
	http := res.Config.T.Structure.HTTPTable
	dns := res.Config.T.Structure.DNSTable
	strobe := res.Config.T.Structure.FrequentConnTable

	names, err := res.DB.Session.DB(database).CollectionNames()
	if err != nil || len(names) == 0 {
		return cli.NewExitError("Failed to find analysis results", -1)
	}

	if !forceFlag {
		fmt.Print("Are you sure you want to reset analysis for ", database, " [y/N] ")

		read := bufio.NewReader(os.Stdin)

		response, _ := read.ReadString('\n')

		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			fmt.Println("Resetting database:", database)
		} else {
			return cli.NewExitError("Database "+database+" was not reset.", 0)
		}
	}

	//check if we had an issue dropping a collection
	var err2Flag error
	for _, name := range names {
		switch name {
		case conn, http, dns, strobe:
			continue
		default:
			err2 := res.DB.Session.DB(database).C(name).DropCollection()
			if err2 != nil {
				fmt.Fprintf(os.Stderr, "Failed to drop collection: %s\n", err2.Error())
				err2Flag = err2
			}
		}
	}

	//change metadb
	err3 := res.MetaDB.MarkDBAnalyzed(database, false)

	if err3 != nil {
		return cli.NewExitError("Failed to update metadb", -1)
	}

	if err == nil && err2Flag == nil && err3 == nil {
		fmt.Fprintf(os.Stdout, "\t[-] Successfully reset analysis of %s.\n", database)
	}
	return nil
}
