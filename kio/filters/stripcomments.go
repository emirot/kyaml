// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"github.com/emirot/kyaml/kio"
	"github.com/emirot/kyaml/yaml"
)

type StripCommentsFilter struct{}

var _ kio.Filter = StripCommentsFilter{}

func (f StripCommentsFilter) Filter(slice []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range slice {
		stripComments(slice[i].YNode())
	}
	return slice, nil
}

func stripComments(node *yaml.Node) {
	if node == nil {
		return
	}
	node.HeadComment = ""
	node.LineComment = ""
	node.FootComment = ""
	for i := range node.Content {
		stripComments(node.Content[i])
	}
}
