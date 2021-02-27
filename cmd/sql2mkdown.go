package cmd

import (
	"errors"

	"github.com/gookit/gcli/v3"
)

var SQL2MkDown = &gcli.Command{
	Name:    "sql2md",
	Aliases: []string{"sql-tomd"},
	Desc:  "convert create table SQL to markdown table",
	Config: func(c *gcli.Command) {

	},
	Func: func(c *gcli.Command, _ []string) error {
		return errors.New("TODO")
	},
}
