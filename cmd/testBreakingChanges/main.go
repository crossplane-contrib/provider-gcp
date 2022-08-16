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

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func main() {
	fmt.Println("----Target successful,this file inside cmd/testBreakingChanges runs successfully----")

	oldfile, err1 := os.ReadFile("old.yaml")
	if err1 != nil {
		log.Fatal(err1)
	}
	fmt.Println(err1)

	newfile, err2 := os.ReadFile("new.yaml")
	if err2 != nil {
		log.Fatal(err2)
	}

	old := &v1.CustomResourceDefinition{}
	err3 := yaml.Unmarshal(oldfile, old)
	if err3 != nil {
		fmt.Println(err3)
	}

	new := &v1.CustomResourceDefinition{}

	err4 := yaml.Unmarshal(newfile, new)
	if err4 != nil {
		fmt.Println(err4)
	}

	list := PrintFields(old.Spec.Versions[0].Schema.OpenAPIV3Schema, "", new.Spec.Versions[0].Schema.OpenAPIV3Schema)

	for i := 0; i < len(list); i++ {
		fmt.Sprintln(list[i])
	}

}

// PrintFields function recursively traverses through the keys.
func PrintFields(sch *v1.JSONSchemaProps, prefix string, newSchema *v1.JSONSchemaProps) []string {

	a := make([]string, 25, 35)

	if len(sch.Properties) == 0 {
		return []string{}
	}

	for key, val := range sch.Properties {

		val := val
		var temp string

		if prefix == "" {
			temp = key
		} else {
			temp = prefix + "." + key
		}

		prop, ok := newSchema.Properties[key]
		if !ok {
			temp += ": does not exist"
			// fmt.Printf("%s%s: does not exist.\n", prefix, key)
			a = append(a, temp)
		}
		// else {
		// 	fmt.Printf("%s\n", temp)
		// }

		PrintFields(&val, temp, &prop)
	}
	// fmt.Println(a)
	return a

}
