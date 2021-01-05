package main

import (
	"errors"
	"fmt"
	"strconv"
)

// MergeCmd ...
type MergeCmd struct {
	PR int `arg name optional:pr`
}

// Run ...
func (mc *MergeCmd) Run(cmd *CmdContext) error {
	if mc.PR == 0 { // not given
		var err error
		mc.PR, err = strconv.Atoi(Pick(cmd, Lister(cmd.gte), "Which PR to merge?"))
		if err != nil {
			return err
		}
	}
	if mc.PR == 0 {
		return nil
	}
	c := cmd.ctx

	c.PR.ID = strconv.Itoa(mc.PR)
	err := cmd.gte.GetPR(c)
	if err != nil {
		fmt.Printf("Failed getting PR information.")
		return err
	}
	err = cmd.jra.GetIssue(c)
	if err != nil {
		fmt.Printf("Failed getting JIRA issue information.")
		return err
	}

	pol := &Policy{}
	if pol.
		Add(Enforcer(PROwnerChecker{})).
		Add(Enforcer(PRTitleChecker{})).
		Add(Enforcer(&IssueMergable{})).
		Check(c) != true {
		return errors.New("----")
	}

	fmt.Println("")
	Info("Merging PR...")
	err = cmd.gte.MergePR(c)
	if err != nil {
		fmt.Printf("Failed to merge. Please check at %s", c.PR.URL)
		return err
	}
	Info("Updating JIRA issue...")
	cmd.jra.UpdateIssue(c,
		fmt.Sprintf("PR %s signed off by %s and %s", c.PR.ID, c.PR.Owners[0], c.PR.Owners[1]),
		fmt.Sprintf("%d", cmd.conf.YUN.Done))
	return nil
}
