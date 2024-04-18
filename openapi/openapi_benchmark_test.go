// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"path/filepath"
	"testing"

	v1212 "github.com/emirot/kyaml/openapi/kubernetesapi/v1212"
	openapi_v2 "github.com/google/gnostic/openapiv2"
	"google.golang.org/protobuf/proto"
)

func BenchmarkProtoUnmarshal(t *testing.B) {
	assetName := filepath.Join(
		"kubernetesapi",
		"v1212",
		"swagger.pb")

	b := v1212.MustAsset(assetName)

	for i := 0; i < t.N; i++ {
		// We parse protobuf and get an openapiv2.Document here.
		if err := proto.Unmarshal(b, &openapi_v2.Document{}); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
		}
	}
}
