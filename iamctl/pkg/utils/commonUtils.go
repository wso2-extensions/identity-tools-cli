/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
*
* WSO2 LLC. licenses this file to you under the Apache License,
* Version 2.0 (the "License"); you may not use this file except
* in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied. See the License for the
* specific language governing permissions and limitations
* under the License.
 */

package utils

import (
	"path/filepath"
	"strings"
)

type FileInfo struct {
	ResourceName  string
	FileName      string
	FileExtension string
}

func GetFileInfo(filePath string) (fileInfo FileInfo) {

	fileInfo.FileName = filepath.Base(filePath)
	fileInfo.FileExtension = filepath.Ext(fileInfo.FileName)
	fileInfo.ResourceName = strings.TrimSuffix(fileInfo.FileName, fileInfo.FileExtension)

	return fileInfo
}

func Contains(slice []string, item string) bool {

	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
