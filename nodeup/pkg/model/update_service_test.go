/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	"testing"

	"k8s.io/kops/upup/pkg/fi"
)

func TestUpdateServiceBuilderAutomaticUpgrade(t *testing.T) {
	RunGoldenTest(t, "tests/updateservicebuilder/automatic", "updateservice", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		builder := UpdateServiceBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestUpdateServiceBuilderExternal(t *testing.T) {
	RunGoldenTest(t, "tests/updateservicebuilder/external", "updateservice", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		builder := UpdateServiceBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}
