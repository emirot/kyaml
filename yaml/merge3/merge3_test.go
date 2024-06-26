// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

import (
	"strings"
	"testing"

	. "github.com/emirot/kyaml/yaml/merge3"
	"github.com/stretchr/testify/assert"
)

var testCases = [][]testCase{scalarTestCases, listTestCases, mapTestCases, elementTestCases, kustomizationTestCases}

func TestMerge(t *testing.T) {
	for i := range testCases {
		for j := range testCases[i] {
			tc := testCases[i][j]
			t.Run(tc.description, func(t *testing.T) {
				actual, err := MergeStrings(tc.local, tc.origin, tc.update, tc.infer)
				if tc.err == nil {
					if !assert.NoError(t, err, tc.description) {
						t.FailNow()
					}
					if !assert.Equal(t,
						strings.TrimSpace(tc.expected), strings.TrimSpace(actual), tc.description) {
						t.FailNow()
					}
				} else {
					if !assert.Errorf(t, err, tc.description) {
						t.FailNow()
					}
					if !assert.Contains(t, tc.err.Error(), err.Error()) {
						t.FailNow()
					}
				}
			})
		}
	}
}

type testCase struct {
	description string
	origin      string
	update      string
	local       string
	expected    string
	err         error
	infer       bool
}
