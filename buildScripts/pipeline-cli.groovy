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
def VERSION
def CLI_REPO = "https://github.com/wso2-extensions/identity-tools-cli.git"
def BRANCH = "master"
def repo = "wso2-extensions/identity-tools-cli"

node('PRODUCT_ECS') {
    stage('Preparation') {
        // Get some code from a GitHub repository
        checkout([$class           : 'GitSCM', branches: [[name: BRANCH]],
                  userRemoteConfigs: [[url: CLI_REPO]]])

    }
    stage('Build') {
        if (ReleaseVersion != "") {
            VERSION = ReleaseVersion
            sh """
                VERSION=${VERSION} make install-cli
              """
        } else {
            sh """
            make install-cli
          """
        }

        if (ReleaseVersion != "") {
            withCredentials([usernamePassword(credentialsId: '4ff4a55b-1313-45da-8cbf-b2e100b1accd', usernameVariable: 'GIT_USERNAME', passwordVariable: 'GIT_PASSWORD')]) {
                def patternToFind = "VERSION" + version()
                echo patternToFind
                textToReplace = "VERSION?=" + DevelopmentVersion
                echo textToReplace
                replaceText("Makefile", patternToFind, textToReplace)

                sh 'git config  user.email "email-id"'
                sh 'git status'
                sh 'git add Makefile'
                sh 'git commit -m "Update to next development version"'
                sh 'git push -u https://{GIT_USERNAME}:${GIT_PASSWORD}@github.com/wso2-extensions/identity-tools-cli.git HEAD:master'

                def response = sh returnStdout: true,
                        script: "curl --retry 5 -s -u ${GIT_USERNAME}:${GIT_PASSWORD} " +
                                "-d '{\"tag_name\": \"v${ReleaseVersion}\", \"target_commitish\": \"${BRANCH}\", \"name\":\"Identity CLI-tool Release v${ReleaseVersion}\",\"body\":\"Identity CLI-tool version v${ReleaseVersion} released! \",\"prerelease\": true}' " +
                                "https://api.github.com/repos/${repo}/releases"

                uploadUrl = getUploadUrl(response)

                macfile = sh(returnStdout: true, script: "basename iamctl/build/iamctl-*-macosx-x64.tar.gz").trim()
                ubuntufile = sh(returnStdout: true, script: "basename iamctl/build/iamctl-*-linux-x64.tar.gz").trim()
                windowsfile = sh(returnStdout: true, script: "basename iamctl/build/iamctl-*-windows-x64.zip").trim()


                sh returnStdout: true,
                        script: "curl -s -H \"Content-Type: application/octet-stream\" -u ${GIT_USERNAME}:${GIT_PASSWORD} " +
                                "--data-binary @iamctl/build/${macfile} " +
                                "${uploadUrl}?name=${macfile}\\&label=${macfile}"

                sh returnStdout: true,
                        script: "curl -s -H \"Content-Type: application/octet-stream\" -u ${GIT_USERNAME}:${GIT_PASSWORD} " +
                                "--data-binary @iamctl/build/${ubuntufile} " +
                                "${uploadUrl}?name=${ubuntufile}\\&label=${ubuntufile}"

                sh returnStdout: true,
                        script: "curl -s -H \"Content-Type: application/octet-stream\" -u ${GIT_USERNAME}:${GIT_PASSWORD} " +
                                "--data-binary @iamctl/build/${windowsfile} " +
                                "${uploadUrl}?name=${windowsfile}\\&label=${windowsfile}"
                //end withCredentials
            }
        }

    }
    stage('Results') {
        archive 'iamctl/build/*.tar.gz'
        archive 'iamctl/build/*.zip'
    }

}

def version() {
    def matcher = readFile("Makefile") =~ 'VERSION(.*)'
    return matcher ? matcher[0][1] : null
}

def replaceText(String filepath, String pattern, String replaceText) {
    def text = readFile filepath
    def isExists = fileExists filepath
    if (isExists) {

        def fileText = text.replace(pattern, replaceText)
        writeFile file: filepath, text: fileText
    } else {
        println("WARN: " + filepath + "file not found")
    }
}

@NonCPS
def getUploadUrl(response) {
    JsonSlurper slurper = new JsonSlurper();
    def data = slurper.parseText(response)
    def upload_url = data.upload_url
    return upload_url.substring(0, upload_url.lastIndexOf("{"))
}
