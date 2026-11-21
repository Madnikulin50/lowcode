package types

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
/// the code is regenerated from anomaly/types/types.yaml

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBaselineSetWalk(t *testing.T) {
	var (
		value = make(BaselineSet, 3)
		req   = require.New(t)
	)

	// check walk with no errors
	{
		err := value.Walk(func(*Baseline) error {
			return nil
		})
		req.NoError(err)
	}

	// check walk with error
	req.Error(value.Walk(func(*Baseline) error { return fmt.Errorf("walk error") }))
}

func TestBaselineSetFilter(t *testing.T) {
	var (
		value = make(BaselineSet, 3)
		req   = require.New(t)
	)

	// filter nothing
	{
		set, err := value.Filter(func(*Baseline) (bool, error) {
			return true, nil
		})
		req.NoError(err)
		req.Equal(len(set), len(value))
	}

	// filter one item
	{
		found := false
		set, err := value.Filter(func(*Baseline) (bool, error) {
			if !found {
				found = true
				return found, nil
			}
			return false, nil
		})
		req.NoError(err)
		req.Len(set, 1)
	}

	// filter error
	{
		_, err := value.Filter(func(*Baseline) (bool, error) {
			return false, fmt.Errorf("filter error")
		})
		req.Error(err)
	}
}

func TestBaselineSetIDs(t *testing.T) {
	var (
		value = make(BaselineSet, 3)
		req   = require.New(t)
	)

	// construct objects
	value[0] = new(Baseline)
	value[1] = new(Baseline)
	value[2] = new(Baseline)
	// set ids
	value[0].ID = 1
	value[1].ID = 2
	value[2].ID = 3

	// Find existing
	{
		val := value.FindByID(2)
		req.Equal(uint64(2), val.ID)
	}

	// Find non-existing
	{
		val := value.FindByID(4)
		req.Nil(val)
	}

	// List IDs from set
	{
		val := value.IDs()
		req.Equal(len(val), len(value))
	}
}

func TestFindingSetWalk(t *testing.T) {
	var (
		value = make(FindingSet, 3)
		req   = require.New(t)
	)

	// check walk with no errors
	{
		err := value.Walk(func(*Finding) error {
			return nil
		})
		req.NoError(err)
	}

	// check walk with error
	req.Error(value.Walk(func(*Finding) error { return fmt.Errorf("walk error") }))
}

func TestFindingSetFilter(t *testing.T) {
	var (
		value = make(FindingSet, 3)
		req   = require.New(t)
	)

	// filter nothing
	{
		set, err := value.Filter(func(*Finding) (bool, error) {
			return true, nil
		})
		req.NoError(err)
		req.Equal(len(set), len(value))
	}

	// filter one item
	{
		found := false
		set, err := value.Filter(func(*Finding) (bool, error) {
			if !found {
				found = true
				return found, nil
			}
			return false, nil
		})
		req.NoError(err)
		req.Len(set, 1)
	}

	// filter error
	{
		_, err := value.Filter(func(*Finding) (bool, error) {
			return false, fmt.Errorf("filter error")
		})
		req.Error(err)
	}
}

func TestFindingSetIDs(t *testing.T) {
	var (
		value = make(FindingSet, 3)
		req   = require.New(t)
	)

	// construct objects
	value[0] = new(Finding)
	value[1] = new(Finding)
	value[2] = new(Finding)
	// set ids
	value[0].ID = 1
	value[1].ID = 2
	value[2].ID = 3

	// Find existing
	{
		val := value.FindByID(2)
		req.Equal(uint64(2), val.ID)
	}

	// Find non-existing
	{
		val := value.FindByID(4)
		req.Nil(val)
	}

	// List IDs from set
	{
		val := value.IDs()
		req.Equal(len(val), len(value))
	}
}

func TestRuleSetWalk(t *testing.T) {
	var (
		value = make(RuleSet, 3)
		req   = require.New(t)
	)

	// check walk with no errors
	{
		err := value.Walk(func(*Rule) error {
			return nil
		})
		req.NoError(err)
	}

	// check walk with error
	req.Error(value.Walk(func(*Rule) error { return fmt.Errorf("walk error") }))
}

func TestRuleSetFilter(t *testing.T) {
	var (
		value = make(RuleSet, 3)
		req   = require.New(t)
	)

	// filter nothing
	{
		set, err := value.Filter(func(*Rule) (bool, error) {
			return true, nil
		})
		req.NoError(err)
		req.Equal(len(set), len(value))
	}

	// filter one item
	{
		found := false
		set, err := value.Filter(func(*Rule) (bool, error) {
			if !found {
				found = true
				return found, nil
			}
			return false, nil
		})
		req.NoError(err)
		req.Len(set, 1)
	}

	// filter error
	{
		_, err := value.Filter(func(*Rule) (bool, error) {
			return false, fmt.Errorf("filter error")
		})
		req.Error(err)
	}
}

func TestRuleSetIDs(t *testing.T) {
	var (
		value = make(RuleSet, 3)
		req   = require.New(t)
	)

	// construct objects
	value[0] = new(Rule)
	value[1] = new(Rule)
	value[2] = new(Rule)
	// set ids
	value[0].ID = 1
	value[1].ID = 2
	value[2].ID = 3

	// Find existing
	{
		val := value.FindByID(2)
		req.Equal(uint64(2), val.ID)
	}

	// Find non-existing
	{
		val := value.FindByID(4)
		req.Nil(val)
	}

	// List IDs from set
	{
		val := value.IDs()
		req.Equal(len(val), len(value))
	}
}
