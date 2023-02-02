/*
 * Copyright (c) 2020, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package cmd

import "testing"

func Test_start(t *testing.T) {
	type args struct {
		serverUrl string
		userName  string
		password  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Start_iamctl",
			args: args{"https://localhost:9443", "admin", "admin"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := start(tt.args.serverUrl, tt.args.userName, tt.args.password); got != tt.want {
				t.Errorf("start() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sendOAuthRequest(t *testing.T) {
	type args struct {
		userName string
		password string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := sendOAuthRequest(tt.args.userName, tt.args.password)
			if got != tt.want {
				t.Errorf("sendOAuthRequest() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("sendOAuthRequest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
