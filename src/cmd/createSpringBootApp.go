/*
Copyright (c) 2020, WSO2 Inc. (http://www.wso2.com). All Rights Reserved.

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
package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Install Spring-boot artifacts by decoding the response from artifact service.
func installSpringBootArtifacts(respBody []byte) {

	response := HttpResponse{}
	_ = json.Unmarshal(respBody, &response)
	if len(response.ResponseData.Data) == 0 {
		fmt.Println("No any response from server...!")
		return
	}
	for i := 0; i < len(response.ResponseData.Data); i++ {
		artifactData, err := base64.StdEncoding.DecodeString(response.ResponseData.Data[i].Data)
		artifactMetadataPath := response.ResponseData.Data[i].Metadata.Path
		artifactMetadataOperation := response.ResponseData.Data[i].Metadata.Operation

		var MetadataCopyOperation = "copy"
		if artifactMetadataOperation == MetadataCopyOperation {
			copyArtifacts(artifactData, artifactMetadataPath)
		}
		if err != nil {
			fmt.Println("Error while decoding OIDC registration artifact")
			return
		}
	}
	fmt.Println("Successfully installed the artifacts and secured with OIDC using WSO2 IS..!")
}

func copyArtifacts(artifactData []byte, artifactMetadataPath string) {

	createFileIfNotExist(artifactMetadataPath)
	writeFile(artifactMetadataPath, string(artifactData))
}
