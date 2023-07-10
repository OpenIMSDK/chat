# Copyright © 2023 OpenIM open source community. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
echo "start building..."
image=openim/openim_chat:v1.1.0
echo "current image version ====> ${image}"

chmod +x ./*.sh
echo "starting running bash shell ./build_all_service.sh"
./build_all_service.sh
cd ../
docker build -t $image . -f ./deploy.Dockerfile
docker push $image
echo "build ok"
