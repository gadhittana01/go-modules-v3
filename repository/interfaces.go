package repository

// This file contains marker interfaces that services can optionally use
// The actual Repository interface should be defined in each service's repository package

// BaseQuerier is a marker interface
// Each service defines their own Querier interface with specific methods
type BaseQuerier interface {
	// Services will define their specific query methods
}
