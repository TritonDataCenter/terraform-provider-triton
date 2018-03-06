//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/joyent/triton-go/client"
	pkgerrors "github.com/pkg/errors"
)

const groupsPath = "/v1/tsg"

type GroupsClient struct {
	client *client.Client
}

type ServiceGroup struct {
	ID                  int64  `json:"id"`
	GroupName           string `json:"group_name"`
	TemplateID          int64  `json:"template_id"`
	AccountID           string `json:"account_id"`
	Capacity            int    `json:"capacity"`
	HealthCheckInterval int    `json:"health_check_interval"`
}

type ListGroupsInput struct{}

func (c *GroupsClient) List(ctx context.Context, _ *ListGroupsInput) ([]*ServiceGroup, error) {
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   groupsPath,
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to list groups")
	}

	var results []*ServiceGroup
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode list groups response")
	}

	return results, nil
}

type GetGroupInput struct {
	Name string
}

func (i *GetGroupInput) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("group name can not be empty")
	}

	return nil
}

func (c *GroupsClient) Get(ctx context.Context, input *GetGroupInput) (*ServiceGroup, error) {
	if err := input.Validate(); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to validate get group input")
	}

	fullPath := path.Join(groupsPath, input.Name)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to get service group")
	}

	var results *ServiceGroup
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode get group response")
	}

	return results, nil
}

type CreateGroupInput struct {
	GroupName           string `json:"group_name"`
	TemplateID          int64  `json:"template_id"`
	Capacity            int    `json:"capacity"`
	HealthCheckInterval int    `json:"health_check_interval"`
}

func (input *CreateGroupInput) toAPI() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if input.GroupName != "" {
		result["group_name"] = input.GroupName
	}

	if input.TemplateID == 0 {
		return nil, fmt.Errorf("unable to create service group without template ID")
	}
	result["template_id"] = input.TemplateID

	result["capacity"] = input.Capacity
	result["health_check_interval"] = input.HealthCheckInterval

	return result, nil
}

func (c *GroupsClient) Create(ctx context.Context, input *CreateGroupInput) error {
	body, err := input.toAPI()
	if err != nil {
		return pkgerrors.Wrap(err, "unable to validate create group input")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   groupsPath,
		Body:   body,
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return pkgerrors.Wrap(err, "unable to create group")
	}

	return nil
}

type DeleteGroupInput struct {
	Name string
}

func (i *DeleteGroupInput) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("group name can not be empty")
	}

	return nil
}

func (c *GroupsClient) Delete(ctx context.Context, input *DeleteGroupInput) error {
	if err := input.Validate(); err != nil {
		return pkgerrors.Wrap(err, "unable to validate delete group input")
	}

	fullPath := path.Join(groupsPath, input.Name)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return pkgerrors.Wrap(err, "unable to delete group")
	}

	return nil
}