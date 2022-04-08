/*
Copyright 2019 The Crossplane Authors.

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

package googleapi

import (
	"net/http"
	"testing"

	"google.golang.org/api/googleapi"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
)

func TestIsErrorNotFound(t *testing.T) {
	tests := []struct {
		name string
		args error
		want bool
	}{
		{name: "Nil", args: nil, want: false},
		{name: "Other", args: errors.New("foo"), want: false},
		{name: "404", args: &googleapi.Error{Code: http.StatusNotFound}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsErrorNotFound(tt.args); got != tt.want {
				t.Errorf("IsErrorBucketNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
