/*
Copyright 2022 The Crossplane Authors.

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

package main

import (
	"log"
	"os"
	"reflect"
	"testing"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func testInput(oldyaml, newyaml string) []string {

	// Reading old yaml
	oldfile, err := os.ReadFile(oldyaml)
	if err != nil {
		log.Fatal(err)
	}

	// Reading new yaml
	newfile, err := os.ReadFile(newyaml)
	if err != nil {
		log.Fatal(err)
	}

	old := &v1.CustomResourceDefinition{}
	err = yaml.Unmarshal(oldfile, old)
	if err != nil {
		log.Fatal(err)
	}

	new := &v1.CustomResourceDefinition{}
	err = yaml.Unmarshal(newfile, new)
	if err != nil {
		log.Fatal(err)
	}
	return PrintFields(old.Spec.Versions[0].Schema.OpenAPIV3Schema, "", new.Spec.Versions[0].Schema.OpenAPIV3Schema)
}

func TestBreakingChanges(t *testing.T) {
	type args struct {
		oldyaml string
		newyaml string
	}
	type want struct {
		result []string
	}
	cases := map[string]struct {
		args
		want
	}{
		"Nochange": {
			args: args{
				oldyaml: "old.yaml",
				newyaml: "new.yaml",
			},
			want: want{
				result: []string{},
			},
		},
		"spec.forProvider.description": {
			args: args{
				oldyaml: "old.yaml",
				newyaml: "new.yaml",
			},
			want: want{
				result: []string{"spec.forProvider.description"},
			},
		},
		"spec.forProvider.enableLogging": {
			args: args{
				oldyaml: "old.yaml",
				newyaml: "new.yaml",
			},
			want: want{
				result: []string{"spec.forProvider.enableLogging"},
			},
		},
		"spec.forProvider.enableInboundForwarding": {
			args: args{
				oldyaml: "old.yaml",
				newyaml: "new.yaml",
			},
			want: want{
				result: []string{"spec.forProvider.enableInboundForwarding"},
			},
		},
		"spec.providerConfigRef": {
			args: args{
				oldyaml: "old.yaml",
				newyaml: "new.yaml",
			},
			want: want{
				result: []string{"spec.providerConfigRef"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := testInput(tc.oldyaml, tc.newyaml)
			if len(tc.want.result) == 0 && len(got) == 0 {
				t.Log("Both are same yaml file\n")
			} else if reflect.DeepEqual(tc.want.result, got) {
				t.Log("PrintFields(...): want: ", tc.want.result, "\ngot: ", got)
			}

		})
	}

}
