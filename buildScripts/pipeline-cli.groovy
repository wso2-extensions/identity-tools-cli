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
pipeline {
    agent any
    stages {
        stage('Build') {
            steps {

                git url: 'https://github.com/piraveena/identity-tools-cli.git',
                        branch: 'test-build'

                sh '''#!/bin/bash +x
                    echo "***********************************************************"
                    echo "Building the cli tool"
                    echo "***********************************************************"  
                    echo "***********************************************************"                                  
                   make install-cli
                '''
            }
            post {
                // archive the installers.
                success {
                    archiveArtifacts 'src/build/*.tar.gz'
                    archiveArtifacts 'src/build/*.zip'
                }
            }

        }
    }
}
