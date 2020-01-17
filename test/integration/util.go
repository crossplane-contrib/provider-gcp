// +build integration

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

package integration

import (
	"context"
	"io/ioutil"
	"time"

	"sigs.k8s.io/yaml"
)

func waitFor(ctx context.Context, interval time.Duration, check func(chan error)) error {
	ch := make(chan error, 1)
	go func() {
		for {
			check(ch)
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

func unmarshalFromFile(path string, obj interface{}) error {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(dat, obj); err != nil {
		return err
	}
	return nil
}
