/*
Copyright 2016 The Rook Authors. All rights reserved.

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
package client

import (
	"encoding/json"

	"github.com/rook/rook/pkg/model"
)

const (
	statusQueryName = "status"
)

func (a *RookNetworkRestClient) GetStatusDetails() (model.StatusDetails, error) {
	body, err := a.DoGet(statusQueryName)
	if err != nil {
		return model.StatusDetails{}, err
	}

	var statusDetails model.StatusDetails
	err = json.Unmarshal(body, &statusDetails)
	if err != nil {
		return model.StatusDetails{}, err
	}

	return statusDetails, nil
}
