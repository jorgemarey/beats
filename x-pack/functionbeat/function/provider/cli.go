// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package provider

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/x-pack/functionbeat/manager/core/bundle"
)

// CLIManager is the interface implemented by each provider to expose a command CLI interface
// to their interface.
type CLIManager interface {
	// Deploy takes a function name and deploy functionbeat and the function configuration to the provider.
	Deploy(string) error

	//Update takes a function name and update the configuration to the remote provider.
	Update(string) error

	// Remove takes a function name and remove the specific function from the remote provider.
	Remove(string) error

	// Export prints exported function template to stdout.
	Export(string) error
}

// ZipResourcesFunc is a function which returns the list of resources for each function type for a provider.
type ZipResourcesFunc func() map[string][]bundle.Resource

// CLIManagerFactory factory method to call to create a new CLI manager
type CLIManagerFactory func(*logp.Logger, *common.Config, Provider) (CLIManager, error)