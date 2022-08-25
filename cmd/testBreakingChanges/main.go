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
	"log"
	"os"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func main() {
	fmt.Println("----Target successful,this file inside cmd/testBreakingChanges runs successfully----")

	oldfile, err := os.ReadFile("old.yaml")
	if err != nil {
		log.Fatal(err)
	}

	newfile, err := os.ReadFile("new.yaml")
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

	list := PrintFields(old.Spec.Versions[0].Schema.OpenAPIV3Schema, "", new.Spec.Versions[0].Schema.OpenAPIV3Schema)

	for i := 0; i < len(list); i++ {
		fmt.Println(list[i])
		fmt.Printf("%T", list[i])
	}
}

// PrintFields function recursively traverses through the keys.
func PrintFields(sch *v1.JSONSchemaProps, prefix string, newSchema *v1.JSONSchemaProps) []string {

	a := make([]string, 10, 20)

	if len(sch.Properties) == 0 {
		return nil
	}

	for key := range sch.Properties {
		val := sch.Properties[key]
		var temp string

		if prefix == "" {
			temp = key
		} else {
			temp = prefix + "." + key
		}

		prop, ok := newSchema.Properties[key]

		if !ok {
			a = append(a, temp)
			continue
		}
		a = append(a, PrintFields(&val, temp, &prop)...)
	}
	return a
}
