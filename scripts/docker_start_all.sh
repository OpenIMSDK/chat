#!/bin/bash
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

#Include shell font styles and some basic information
SCRIPTS_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
OPENIM_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

${SCRIPTS_ROOT}/start_all.sh

#fixme The 10 second delay to start the project is for the docker-compose one-click to start openIM when the infrastructure dependencies are not started

sleep 10
time=`date +"%Y-%m-%d %H:%M:%S"`
echo "==========================================================">>$OPENIM_ROOT/logs/openIM.log 2>&1 &
echo "==========================================================">>$OPENIM_ROOT/logs/openIM.log 2>&1 &
echo "==========================================================">>$OPENIM_ROOT/logs/openIM.log 2>&1 &
echo "==========server start time:${time}===========">>$OPENIM_ROOT/logs/openIM.log 2>&1 &
echo "==========================================================">>$OPENIM_ROOT/logs/openIM.log 2>&1 &
echo "==========================================================">>$OPENIM_ROOT/logs/openIM.log 2>&1 &
echo "==========================================================">>$OPENIM_ROOT/logs/openIM.log 2>&1 &


i=1
while ((i == 1))
do
  sleep 5
done

sleep 15

#fixme prevents the openIM service exit after execution in the docker container
tail -f /dev/null
