// Copyright © 2023 OpenIM open source community. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chat

import (
	"io"
	"time"

	"github.com/openimsdk/chat/internal/api/util"

	"github.com/gin-gonic/gin"
	"github.com/openimsdk/chat/pkg/common/apistruct"
	"github.com/openimsdk/chat/pkg/common/constant"
	"github.com/openimsdk/chat/pkg/common/imapi"
	"github.com/openimsdk/chat/pkg/common/mctx"
	"github.com/openimsdk/chat/pkg/protocol/admin"
	chatpb "github.com/openimsdk/chat/pkg/protocol/chat"
	constantpb "github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/a2r"
	"github.com/openimsdk/tools/apiresp"
	"github.com/openimsdk/tools/errs"
)

func New(chatClient chatpb.ChatClient, adminClient admin.AdminClient, imApiCaller imapi.CallerInterface, api *util.Api) *Api {
	return &Api{
		Api:         api,
		chatClient:  chatClient,
		adminClient: adminClient,
		imApiCaller: imApiCaller,
	}
}

type Api struct {
	*util.Api
	chatClient  chatpb.ChatClient
	adminClient admin.AdminClient
	imApiCaller imapi.CallerInterface
}

// ################## ACCOUNT ##################

func (o *Api) SendVerifyCode(c *gin.Context) {
	req, err := a2r.ParseRequest[chatpb.SendVerifyCodeReq](c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	ip, err := o.GetClientIP(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req.Ip = ip
	resp, err := o.chatClient.SendVerifyCode(c, req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, resp)
}

func (o *Api) VerifyCode(c *gin.Context) {
	a2r.Call(chatpb.ChatClient.VerifyCode, o.chatClient, c)
}

func (o *Api) RegisterUser(c *gin.Context) {
	req, err := a2r.ParseRequest[chatpb.RegisterUserReq](c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	ip, err := o.GetClientIP(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req.Ip = ip

	imToken, err := o.imApiCaller.ImAdminTokenWithDefaultAdmin(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiCtx := mctx.WithApiToken(c, imToken)
	rpcCtx := o.WithAdminUser(c)

	// if phone exist, don't return err, just a condition.
	checkResp, err := o.chatClient.CheckPhoneNumberExist(rpcCtx, &chatpb.CheckPhoneNumberExistReq{PhoneNumber: req.User.PhoneNumber})
	if err != nil {
		isUserNotExist, err := o.imApiCaller.AccountCheckSingle(apiCtx, checkResp.Userid)
		if err != nil {
			apiresp.GinError(c, err)
			return
		}
		// if User is  not exist in SDK server. You need delete this user and register new user again.
		if isUserNotExist {
			_, err := o.adminClient.DelAdminAccount(rpcCtx, &admin.DelAdminAccountReq{UserIDs: []string{checkResp.Userid}})
			if err != nil {
				apiresp.GinError(c, err)
				return
			}
		}
	}

	respRegisterUser, err := o.chatClient.RegisterUser(c, req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	userInfo := &sdkws.UserInfo{
		UserID:     respRegisterUser.UserID,
		Nickname:   req.User.Nickname,
		FaceURL:    req.User.FaceURL,
		CreateTime: time.Now().UnixMilli(),
	}
	err = o.imApiCaller.RegisterUser(c, []*sdkws.UserInfo{userInfo})
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	if resp, err := o.adminClient.FindDefaultFriend(rpcCtx, &admin.FindDefaultFriendReq{}); err == nil {
		_ = o.imApiCaller.ImportFriend(apiCtx, respRegisterUser.UserID, resp.UserIDs)
	}
	if resp, err := o.adminClient.FindDefaultGroup(rpcCtx, &admin.FindDefaultGroupReq{}); err == nil {
		_ = o.imApiCaller.InviteToGroup(apiCtx, respRegisterUser.UserID, resp.GroupIDs)
	}
	var resp apistruct.UserRegisterResp
	if req.AutoLogin {
		resp.ImToken, err = o.imApiCaller.UserToken(c, respRegisterUser.UserID, req.Platform)
		if err != nil {
			apiresp.GinError(c, err)
			return
		}
	}
	resp.ChatToken = respRegisterUser.ChatToken
	resp.UserID = respRegisterUser.UserID
	apiresp.GinSuccess(c, &resp)
}

func (o *Api) Login(c *gin.Context) {
	req, err := a2r.ParseRequest[chatpb.LoginReq](c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	ip, err := o.GetClientIP(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req.Ip = ip
	resp, err := o.chatClient.Login(c, req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	imToken, err := o.imApiCaller.UserToken(c, resp.UserID, req.Platform)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &apistruct.LoginResp{
		ImToken:   imToken,
		UserID:    resp.UserID,
		ChatToken: resp.ChatToken,
	})
}

func (o *Api) ResetPassword(c *gin.Context) {
	a2r.Call(chatpb.ChatClient.ResetPassword, o.chatClient, c)
}

func (o *Api) ChangePassword(c *gin.Context) {
	req, err := a2r.ParseRequest[chatpb.ChangePasswordReq](c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	resp, err := o.chatClient.ChangePassword(c, req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	imToken, err := o.imApiCaller.ImAdminTokenWithDefaultAdmin(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	err = o.imApiCaller.ForceOffLine(mctx.WithApiToken(c, imToken), req.UserID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, resp)
}

// ################## USER ##################

func (o *Api) UpdateUserInfo(c *gin.Context) {
	req, err := a2r.ParseRequest[chatpb.UpdateUserInfoReq](c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	respUpdate, err := o.chatClient.UpdateUserInfo(c, req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	opUserType, err := mctx.GetUserType(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	var imToken string
	if opUserType == constant.NormalUser {
		imToken, err = o.imApiCaller.ImAdminTokenWithDefaultAdmin(c)
	} else if opUserType == constant.AdminUser {
		imToken, err = o.imApiCaller.UserToken(c, o.GetDefaultIMAdminUserID(), constantpb.AdminPlatformID)
	} else {
		apiresp.GinError(c, errs.ErrArgs.WrapMsg("opUserType unknown"))
		return
	}
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	var (
		nickName string
		faceURL  string
	)
	if req.Nickname != nil {
		nickName = req.Nickname.Value
	} else {
		nickName = respUpdate.NickName
	}
	if req.FaceURL != nil {
		faceURL = req.FaceURL.Value
	} else {
		faceURL = respUpdate.FaceUrl
	}
	err = o.imApiCaller.UpdateUserInfo(mctx.WithApiToken(c, imToken), req.UserID, nickName, faceURL)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, apistruct.UpdateUserInfoResp{})
}

func (o *Api) FindUserPublicInfo(c *gin.Context) {
	a2r.Call(chatpb.ChatClient.FindUserPublicInfo, o.chatClient, c)
}

func (o *Api) FindUserFullInfo(c *gin.Context) {
	a2r.Call(chatpb.ChatClient.FindUserFullInfo, o.chatClient, c)
}

func (o *Api) SearchUserFullInfo(c *gin.Context) {
	a2r.Call(chatpb.ChatClient.SearchUserFullInfo, o.chatClient, c)
}

func (o *Api) SearchUserPublicInfo(c *gin.Context) {
	a2r.Call(chatpb.ChatClient.SearchUserPublicInfo, o.chatClient, c)
}

func (o *Api) GetTokenForVideoMeeting(c *gin.Context) {
	a2r.Call(chatpb.ChatClient.GetTokenForVideoMeeting, o.chatClient, c)
}

// ################## APPLET ##################

func (o *Api) FindApplet(c *gin.Context) {
	a2r.Call(admin.AdminClient.FindApplet, o.adminClient, c)
}

// ################## CONFIG ##################

func (o *Api) GetClientConfig(c *gin.Context) {
	a2r.Call(admin.AdminClient.GetClientConfig, o.adminClient, c)
}

// ################## CALLBACK ##################

func (o *Api) OpenIMCallback(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req := &chatpb.OpenIMCallbackReq{
		Command: c.Query(constantpb.CallbackCommand),
		Body:    string(body),
	}
	if _, err := o.chatClient.OpenIMCallback(c, req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

func (o *Api) SearchFriend(c *gin.Context) {
	req, err := a2r.ParseRequest[struct {
		UserID string `json:"userID"`
		chatpb.SearchUserInfoReq
	}](c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(c)
	}
	imToken, err := o.imApiCaller.ImAdminTokenWithDefaultAdmin(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	userIDs, err := o.imApiCaller.FriendUserIDs(mctx.WithApiToken(c, imToken), req.UserID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	if len(userIDs) == 0 {
		apiresp.GinSuccess(c, &chatpb.SearchUserInfoResp{})
		return
	}
	req.SearchUserInfoReq.UserIDs = userIDs
	resp, err := o.chatClient.SearchUserInfo(c, &req.SearchUserInfoReq)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, resp)
}
