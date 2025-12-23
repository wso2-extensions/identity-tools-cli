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

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)


// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   internal.ROOT_COMMAND,
	Short: "A CLI tool to manage Identity and Access Management tasks",
	Long:  "A CLI tool to manage Identity and Access Management tasks for WSO2 Identity Server and Asgardeo.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {

	if err := RootCmd.Execute(); err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}
