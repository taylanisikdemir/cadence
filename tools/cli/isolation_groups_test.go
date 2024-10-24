// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cli

import (
	"errors"
	"flag"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/common/types"
)

func TestValidateIsolationGroupArgs(t *testing.T) {

	tests := map[string]struct {
		domainArgs          string
		globalDomainArg     string
		setDrainsArgs       []string
		jsonConfigArgs      string
		removeAllDrainsArgs bool

		requiresDomain bool
		expectedErr    error
	}{
		"valid inputs for doing a drain": {
			domainArgs:      "some-domain",
			globalDomainArg: "",
			setDrainsArgs:   []string{"zone-1", "zone-2"},
			jsonConfigArgs:  "",

			expectedErr: nil,
		},
		"valid json input": {
			domainArgs:      "some-domain",
			globalDomainArg: "",
			setDrainsArgs:   nil,
			jsonConfigArgs:  "{}",

			expectedErr: nil,
		},
		"invalid - no domain": {
			domainArgs:      "",
			globalDomainArg: "",
			setDrainsArgs:   nil,
			jsonConfigArgs:  "{}",
			requiresDomain:  true,

			expectedErr: errors.New("the --domain flag is required"),
		},
		"invalid - global domain": {
			domainArgs:      "",
			globalDomainArg: "second domain",
			setDrainsArgs:   nil,
			jsonConfigArgs:  "{}",
			requiresDomain:  true,

			expectedErr: errors.New("the flag '--domain' has to go at the end"),
		},
		"invalid - no config domain": {
			domainArgs:      "domain",
			globalDomainArg: "",
			setDrainsArgs:   nil,
			jsonConfigArgs:  "",
			requiresDomain:  true,

			expectedErr: errors.New("need to specify either \"set-drains\", \"json\" or \"remove-all-drains\" flags"),
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expectedErr, validateIsolationGroupUpdateArgs(
				td.domainArgs,
				td.globalDomainArg,
				td.setDrainsArgs,
				td.jsonConfigArgs,
				td.removeAllDrainsArgs,
				td.requiresDomain))
		})
	}
}

func TestParseCliInput(t *testing.T) {

	tests := map[string]struct {
		setDrainsArgs  []string
		jsonConfigArgs string

		expected    *types.IsolationGroupConfiguration
		expectedErr error
	}{
		"valid inputs for doing a drain": {
			setDrainsArgs:  []string{"zone-1", "zone-2"},
			jsonConfigArgs: "",

			expected: &types.IsolationGroupConfiguration{
				"zone-1": {Name: "zone-1", State: types.IsolationGroupStateDrained},
				"zone-2": {Name: "zone-2", State: types.IsolationGroupStateDrained},
			},
		},
		"valid json input": {
			setDrainsArgs:  nil,
			jsonConfigArgs: "[{\"Name\": \"zone-1\", \"State\": 2}, {\"Name\": \"zone-2\", \"State\": 1}]",
			expected: &types.IsolationGroupConfiguration{
				"zone-1": {Name: "zone-1", State: types.IsolationGroupStateDrained},
				"zone-2": {Name: "zone-2", State: types.IsolationGroupStateHealthy},
			},

			expectedErr: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := parseIsolationGroupCliInputCfg(
				td.setDrainsArgs,
				td.jsonConfigArgs,
				false,
			)
			assert.Equal(t, td.expected, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}

func TestRenderIsolationGroupNormalOutput(t *testing.T) {

	tests := map[string]struct {
		input          types.IsolationGroupConfiguration
		expectedOutput string
	}{
		"valid inputs for doing a drain": {
			input: types.IsolationGroupConfiguration{
				"zone-1": {
					Name:  "zone-1",
					State: types.IsolationGroupStateHealthy,
				},
				"zone-2": {
					Name:  "zone-2",
					State: types.IsolationGroupStateDrained,
				},
				"zone-3-a-very-long-name": {
					Name:  "zone-3-a-very-long-name",
					State: types.IsolationGroupStateDrained,
				},
				"zone-4": {
					Name:  "zone-4",
					State: 5,
				},
			},
			expectedOutput: `Isolation Groups        State
zone-1                  Healthy
zone-2                  Drained
zone-3-a-very-long-name Drained
zone-4                  Unknown state: 5
`,
		},
		"nothing": {
			input: types.IsolationGroupConfiguration{},
			expectedOutput: `-- No groups found --
`,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expectedOutput, string(renderIsolationGroups(td.input)))
		})
	}
}

func TestAdminGetGlobalIsolationGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Table of test cases
	tests := []struct {
		name             string
		setupMocks       func(*admin.MockClient)
		expectedError    string
		flagFormat       string
		mockDepsError    error
		mockContextError error
	}{
		{
			name: "Success with JSON format",
			setupMocks: func(client *admin.MockClient) {
				expectedResponse := &types.GetGlobalIsolationGroupsResponse{
					IsolationGroups: types.IsolationGroupConfiguration{
						"zone-1": {
							Name:  "zone-1",
							State: types.IsolationGroupStateHealthy,
						},
						"zone-2": {
							Name:  "zone-2",
							State: types.IsolationGroupStateDrained,
						},
					},
				}
				client.EXPECT().
					GetGlobalIsolationGroups(gomock.Any(), gomock.Any()).
					Return(expectedResponse, nil).
					Times(1)
			},
			expectedError: "",
			flagFormat:    "json",
		},
		{
			name: "Failed to get global isolation groups",
			setupMocks: func(client *admin.MockClient) {
				client.EXPECT().
					GetGlobalIsolationGroups(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("failed to get isolation-groups")).
					Times(1)
			},
			expectedError: "failed to get isolation-groups",
			flagFormat:    "json",
		},
	}

	// Loop through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the admin client
			adminClient := admin.NewMockClient(mockCtrl)

			// Set up mocks for the current test case
			tt.setupMocks(adminClient)

			// Create mock app with clientFactoryMock, including any deps errors
			app := NewCliApp(&clientFactoryMock{
				serverAdminClient: adminClient,
			})

			// Create CLI context with flags
			set := flag.NewFlagSet("test", 0)
			set.String(FlagFormat, tt.flagFormat, "Format flag")
			c := cli.NewContext(app, set, nil)

			// Call the function under test
			err := AdminGetGlobalIsolationGroups(c)

			// Check the expected outcome
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdminUpdateGlobalIsolationGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Define table-driven tests
	tests := []struct {
		name             string
		setupMocks       func(*admin.MockClient)
		expectedError    string
		flagDomain       string
		removeAllDrains  bool
		mockDepsError    error
		mockContextError error
		validationError  error
		parseConfigError error
	}{
		{
			name: "Success",
			setupMocks: func(client *admin.MockClient) {
				client.EXPECT().
					UpdateGlobalIsolationGroups(gomock.Any(), gomock.Any()).
					Return(&types.UpdateGlobalIsolationGroupsResponse{}, nil).
					Times(1)
			},
			expectedError:   "",
			flagDomain:      "test-domain",
			removeAllDrains: true,
		},
		{
			name: "parse failure",
			setupMocks: func(client *admin.MockClient) {
			},
			expectedError:   "invalid args:",
			flagDomain:      "test-domain",
			removeAllDrains: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the admin client
			adminClient := admin.NewMockClient(mockCtrl)

			// Set up mocks for the current test case
			tt.setupMocks(adminClient)

			// Create mock app with clientFactoryMock, including any deps errors
			app := NewCliApp(&clientFactoryMock{
				serverAdminClient: adminClient,
			})

			// Set up CLI context with flags
			set := flag.NewFlagSet("test", 0)
			set.String(FlagDomain, tt.flagDomain, "Domain flag")
			set.Bool(FlagIsolationGroupsRemoveAllDrains, tt.removeAllDrains, "RemoveAllDrains flag")
			c := cli.NewContext(app, set, nil)

			// Call the function under test
			err := AdminUpdateGlobalIsolationGroups(c)

			// Check the expected outcome
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
