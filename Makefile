# --------------------------------------------------------------------
# Copyright (c) 2020, WSO2 Inc. (http://wso2.com) All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# -----------------------------------------------------------------------

VERSION?=1.0.8-SNAPSHOT

.PHONY: install-cli
install-cli:
	 cd iamctl && ./build.sh -t main.go -v ${VERSION} -f

.PHONY: install-cli-skip-test
install-cli-skip-test:
	cd iamctl && ./build.sh -t main.go -v ${VERSION} -f test.skip
