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
	"fmt"
	"os"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func main() {
	fmt.Println("----Target successful,this file inside cmd/testBreakingChanges runs successfully----")

	oldfile, _ := os.ReadFile("old.yaml")
	newfile, _ := os.ReadFile("new.yaml")

	old := &v1.CustomResourceDefinition{}
	err := yaml.Unmarshal(oldfile, old)
	if err != nil {
		fmt.Println(err)
	}

	new := &v1.CustomResourceDefinition{}
	err2 := yaml.Unmarshal(newfile, new)
	if err != nil {
		fmt.Println(err2)
	}

	PrintFields(old.Spec.Versions[0].Schema.OpenAPIV3Schema, "", new.Spec.Versions[0].Schema.OpenAPIV3Schema)

}

func PrintFields(sch *v1.JSONSchemaProps, prefix string, newSchema *v1.JSONSchemaProps) {
	if len(sch.Properties) == 0 {
		return
	}

	for key, val := range sch.Properties {

		var temp string

		if prefix == "" {
			temp = key
		} else {
			temp = prefix + "." + key
		}

		prop, ok := newSchema.Properties[key]
		if !ok {
			fmt.Printf("%s: does not exist.\n", temp)

		}
		// to print every other key which exists:
		// else {
		// 	fmt.Printf("%s\n", temp)
		// }

		PrintFields(&val, temp, &prop)

	}
}
