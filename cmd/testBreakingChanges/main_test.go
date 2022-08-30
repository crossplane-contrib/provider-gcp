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
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestBreakingChanges(t *testing.T) {

	type args struct {
		oldyaml *v1.JSONSchemaProps
		newyaml *v1.JSONSchemaProps
	}
	type want struct {
		result []string
	}
	cases := map[string]struct {
		args
		want
	}{
		"No changes": {
			args: args{
				oldyaml: &v1.JSONSchemaProps{
					Properties: map[string]v1.JSONSchemaProps{
						"apiVersion": {},
						"kind":       {},
						"metadata":   {},
						"status":     { 
							Properties: map[string]v1.JSONSchemaProps{
								conditions:{},	
								atProvider: {
									Properties: map[string]v1.JSONSchemaProps{
										"id": {},
									},
					            },
							},
						},
						"spec": {
							Properties: map[string]v1.JSONSchemaProps{
								"deletionPolicy": {},
								"forProvider": {
									Properties: map[string]v1.JSONSchemaProps{
                                       "description": {},
									   "enableInboundForwarding": {},
									   "enableLogging": {},
									   "networks": {},
									   "alternativeNameServerConfig": {
										Properties: map[string]v1.JSONSchemaProps{
											"targetNameServers": {},
									   },
									},
								},
								"providerRef": {
									Properties: map[string]v1.JSONSchemaProps{
                                        "name": {},
									},
								},
								"providerConfigRef": {
									Properties: map[string]v1.JSONSchemaProps{
										"name": {},
									},
								},	
								"publishConnectionDetailsTo": {	
									Properties: map[string]v1.JSONSchemaProps{
										"name": {},
										"configRef": {
											Properties: map[string]v1.JSONSchemaProps{
												"name": {},
											},
										},
										"metadata": {
											Properties: map[string]v1.JSONSchemaProps{
												"type": {},
												"annotations":{},
												"labels": {},
										},	
									},
								},
								"writeConnectionSecretToRef": {
									Properties: map[string]v1.JSONSchemaProps{
                                     "name": {},
									 "namespace":{},
									},
								},
							},
						},
					},				
				},
				newyaml: &v1.JSONSchemaProps{
					Properties: map[string]v1.JSONSchemaProps{
						"apiVersion": {},
						"kind":       {},
						"metadata":   {},
						"status":     { 
							Properties: map[string]v1.JSONSchemaProps{
								conditions:{},	
								atProvider: {
									Properties: map[string]v1.JSONSchemaProps{
										"id": {},
									},
					            },
							},
						},		
						"spec": {
							Properties: map[string]v1.JSONSchemaProps{
								"deletionPolicy":             {},
								"forProvider":                {
									Properties: map[string]v1.JSONSchemaProps{
                                       "description": {},
									   "enableInboundForwarding": {},
									   "enableLogging": {},
									   "networks": {},
									   "alternativeNameServerConfig": {
										Properties: map[string]v1.JSONSchemaProps{
											"targetNameServers": {},
									   },

									},
								},
								"providerRef":          {
									Properties: map[string]v1.JSONSchemaProps{
                                        "name": {},
									},
								},
								"providerConfigRef": {
									Properties: map[string]v1.JSONSchemaProps{
										"name": {},
									},
								},
								"publishConnectionDetailsTo": {
									Properties: map[string]v1.JSONSchemaProps{
										"name": {},
										"configRef": {
											Properties: map[string]v1.JSONSchemaProps{
												"name": {},
											},
										},
										"metadata": {
											Properties: map[string]v1.JSONSchemaProps{
												"type": {},
												"annotations":{},
												"labels": {},
										},	
									},
								},
								"writeConnectionSecretToRef": {
									Properties: map[string]v1.JSONSchemaProps{
                                     "name": {},
									 "namespace":{},
									},
								},
							},
						},
					},
				},	
			},	
			want: want{
					result: []string{ },
			},
		},
			
		"spec.forProvider.description": {
			args: args{
				oldyaml: &v1.JSONSchemaProps{
					Properties: map[string]v1.JSONSchemaProps{
						"spec": {
							Properties: map[string]v1.JSONSchemaProps{
								"deletionPolicy":{},
								 "forProvider": {
									Properties: map[string]v1.JSONSchemaProps{
                                       "description": {},
									   "enableInboundForwarding": {},
									   "enableLogging": {},
									   "networks": {},
									   "alternativeNameServerConfig": {
										Properties: map[string]v1.JSONSchemaProps{
											"targetNameServers": {},
									    },
									   },
									},
								},
							},
						},
					},
				},	
				newyaml: &v1.JSONSchemaProps{
					Properties: map[string]v1.JSONSchemaProps{
						"spec": {
							Properties: map[string]v1.JSONSchemaProps{
								"deletionPolicy":             {},
								 "forProvider":                {
									Properties: map[string]v1.JSONSchemaProps{
                                       // "description": {},
									   "enableInboundForwarding": {},
									   "enableLogging": {},
									   "networks": {},
									   "alternativeNameServerConfig": {
										Properties: map[string]v1.JSONSchemaProps{
											"targetNameServers": {},
									   },
									   },
									},
								},
							},
						},
					},
				},
			},
			want: want{
				result: []string{"spec.forProvider.description"},
		    },
		},

		"kind": {
			args: args{
				oldyaml: &v1.JSONSchemaProps{
					Properties: map[string]v1.JSONSchemaProps{
						"apiVersion": {},
						"kind":       {},
						"metadata":   {},
						"status":     { 
							Properties: map[string]v1.JSONSchemaProps{
								conditions:{},	
								atProvider: {
									Properties: map[string]v1.JSONSchemaProps{
										"id": {},
									},
					            },
							},
						},
						"spec": { 
							Properties: map[string]v1.JSONSchemaProps{
								"deletionPolicy":             {},
						},
					},
				},
				newyaml: &v1.JSONSchemaProps{
					Properties: map[string]v1.JSONSchemaProps{
						"apiVersion": {},
						// "kind":    {},
						"metadata":   {},
						"status":     { 
							Properties: map[string]v1.JSONSchemaProps{
								conditions:{},	
								atProvider: {
									Properties: map[string]v1.JSONSchemaProps{
										"id": {},
									},
					            },
							},
						},
					    "spec": { 
							Properties: map[string]v1.JSONSchemaProps{
							"deletionPolicy":             {},
						},
					},
				},
			},
			want: want{
					result: []string{"kind"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := cmp.Diff(tc.oldyaml, tc.newyaml)
			if len(tc.want.result) == 0 && len(got) == 0 {
				t.Log("Both are same yaml file\n")
			} else !reflect.DeepEqual(tc.want.result, got) {
				t.Log("PrintFields(...): want \n:  got\n:", tc.want.result, got)
			} 

		})
	}
}


