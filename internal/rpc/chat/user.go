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
	"context"
	"errors"
	"github.com/openimsdk/chat/pkg/eerrs"
	"github.com/openimsdk/protocol/wrapperspb"
	"github.com/openimsdk/tools/utils/datautil"
	"github.com/openimsdk/tools/utils/stringutil"
	"strconv"
	"strings"
	"time"

	"github.com/openimsdk/chat/pkg/common/db/dbutil"
	chatdb "github.com/openimsdk/chat/pkg/common/db/table/chat"
	constantpb "github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/mcontext"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/openimsdk/chat/pkg/common/constant"
	"github.com/openimsdk/chat/pkg/common/mctx"
	"github.com/openimsdk/chat/pkg/protocol/chat"
	"github.com/openimsdk/tools/errs"
)

func (o *chatSvr) checkUpdateInfo(ctx context.Context, req *chat.UpdateUserInfoReq) error {
	if req.AreaCode != nil || req.PhoneNumber != nil {
		if !(req.AreaCode != nil && req.PhoneNumber != nil) {
			return errs.ErrArgs.WrapMsg("areaCode and phoneNumber must be set together")
		}
		if req.AreaCode.Value == "" || req.PhoneNumber.Value == "" {
			if req.AreaCode.Value != req.PhoneNumber.Value {
				return errs.ErrArgs.WrapMsg("areaCode and phoneNumber must be set together")
			}
		}
	}
	if req.UserID == "" {
		return errs.ErrArgs.WrapMsg("user id is empty")
	}

	credentials, err := o.Database.TakeCredentialsByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	var (
		credNum, delNum, addNum = len(credentials), 0, 0
	)

	addFunc := func(s *wrapperspb.StringValue) {
		if s != nil {
			if s.Value != "" {
				addNum++
			}
		}
	}

	for _, s := range []*wrapperspb.StringValue{req.Account, req.PhoneNumber, req.Email} {
		addFunc(s)
	}

	for _, credential := range credentials {
		switch credential.Type {
		case constant.CredentialAccount:
			if req.Account != nil {
				if req.Account.Value == credential.Account {
					req.Account = nil
				} else if req.Account.Value == "" {
					delNum += 1
				}
				return nil
			}
		case constant.CredentialPhone:
			if req.PhoneNumber != nil {
				phoneAccount := BuildCredentialPhone(req.AreaCode.Value, req.PhoneNumber.Value)
				if phoneAccount == credential.Account {
					req.AreaCode = nil
					req.PhoneNumber = nil
				} else if req.PhoneNumber.Value == "" {
					delNum += 1
				}
				return nil
			}
		case constant.CredentialEmail:
			if req.Email != nil {
				if req.Email.Value == credential.Account {
					req.Email = nil
				} else if req.Email.Value == "" {
					delNum += 1
				}
				return nil
			}
		}
	}

	if addNum+credNum-delNum <= 0 {
		return errs.ErrArgs.WrapMsg("a login method must exist")
	}

	if req.PhoneNumber.GetValue() != "" {
		if !strings.HasPrefix(req.AreaCode.GetValue(), "+") {
			req.AreaCode.Value = "+" + req.AreaCode.Value
		}
		if _, err := strconv.ParseUint(req.AreaCode.Value[1:], 10, 64); err != nil {
			return errs.ErrArgs.WrapMsg("area code must be number")
		}
		if _, err := strconv.ParseUint(req.PhoneNumber.GetValue(), 10, 64); err != nil {
			return errs.ErrArgs.WrapMsg("phone number must be number")
		}
		_, err := o.Database.TakeAttributeByPhone(ctx, req.AreaCode.GetValue(), req.PhoneNumber.GetValue())
		if err == nil {
			return eerrs.ErrPhoneAlreadyRegister.Wrap()
		} else if !dbutil.IsDBNotFound(err) {
			return err
		}
	}
	if req.Account.GetValue() != "" {
		if !stringutil.IsAlphanumeric(req.Account.GetValue()) {
			return errs.ErrArgs.WrapMsg("account must be alphanumeric")
		}
		_, err := o.Database.TakeAttributeByAccount(ctx, req.Account.GetValue())
		if err == nil {
			return eerrs.ErrAccountAlreadyRegister.Wrap()
		} else if !dbutil.IsDBNotFound(err) {
			return err
		}
	}
	if req.Email.GetValue() != "" {
		if !stringutil.IsValidEmail(req.Email.GetValue()) {
			return errs.ErrArgs.WrapMsg("invalid email")
		}
		_, err := o.Database.TakeAttributeByAccount(ctx, req.Email.GetValue())
		if err == nil {
			return eerrs.ErrAccountAlreadyRegister.Wrap()
		} else if !dbutil.IsDBNotFound(err) {
			return err
		}
	}
	return nil
}

func (o *chatSvr) UpdateUserInfo(ctx context.Context, req *chat.UpdateUserInfoReq) (*chat.UpdateUserInfoResp, error) {

	opUserID, userType, err := mctx.Check(ctx)
	if err != nil {
		return nil, err
	}

	if err = o.checkUpdateInfo(ctx, req); err != nil {
		return nil, err
	}

	isOrgUser, err := o.Database.IsOrgUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	switch userType {
	case constant.NormalUser:
		if isOrgUser {
			if req.UserID != opUserID {
				return nil, errs.ErrNoPermission.WrapMsg("only admin can update other user info")
			}
			if req.AreaCode != nil {
				return nil, errs.ErrNoPermission.WrapMsg("areaCode can not be updated")
			}
			if req.PhoneNumber != nil {
				return nil, errs.ErrNoPermission.WrapMsg("phoneNumber can not be updated")
			}
			if req.Account != nil {
				return nil, errs.ErrNoPermission.WrapMsg("account can not be updated")
			}
			if req.Email != nil {
				return nil, errs.ErrNoPermission.WrapMsg("email can not be updated")
			}
			if req.Level != nil {
				return nil, errs.ErrNoPermission.WrapMsg("level can not be updated")
			}

			if req.Nickname != nil {
				return nil, errs.ErrNoPermission.WrapMsg("nickname can not be updated")
			}
			if req.FaceURL != nil {
				return nil, errs.ErrNoPermission.WrapMsg("faceURL can not be updated")
			}
			if req.Gender != nil {
				return nil, errs.ErrNoPermission.WrapMsg("gender can not be updated")
			}
			if req.Birth != nil {
				return nil, errs.ErrNoPermission.WrapMsg("birth can not be updated")
			}
			if req.EnglishName != nil {
				return nil, errs.ErrNoPermission.WrapMsg("englishName can not be updated")
			}
			if req.Station != nil {
				return nil, errs.ErrNoPermission.WrapMsg("station can not be updated")
			}
			if req.Telephone != nil {
				return nil, errs.ErrNoPermission.WrapMsg("telephone can not be updated")
			}
		}
		if req.RegisterType != nil {
			return nil, errs.ErrNoPermission.WrapMsg("registerType can not be updated")
		}

	case constant.AdminUser:
	default:
		return nil, errs.ErrNoPermission.WrapMsg("user type error")
	}

	update, err := ToDBAttributeUpdate(req, isOrgUser)
	if err != nil {
		return nil, err
	}
	credUpdate, credDel, err := ToDBCredentialUpdate(req, !isOrgUser)
	if err != nil {
		return nil, err
	}
	if len(update) > 0 {
		if err := o.Database.UpdateUseInfo(ctx, req.UserID, update, credUpdate, credDel); err != nil {
			return nil, err
		}
	}
	return &chat.UpdateUserInfoResp{}, nil
}

func (o *chatSvr) FindUserPublicInfo(ctx context.Context, req *chat.FindUserPublicInfoReq) (*chat.FindUserPublicInfoResp, error) {
	if len(req.UserIDs) == 0 {
		return nil, errs.ErrArgs.WrapMsg("UserIDs is empty")
	}
	attributes, err := o.Database.FindAttribute(ctx, req.UserIDs)
	if err != nil {
		return nil, err
	}
	return &chat.FindUserPublicInfoResp{
		Users: DbToPbAttributes(attributes),
	}, nil
}

func (o *chatSvr) AddUserAccount(ctx context.Context, req *chat.AddUserAccountReq) (*chat.AddUserAccountResp, error) {
	if _, _, err := mctx.Check(ctx); err != nil {
		return nil, err
	}

	if err := o.checkRegisterInfo(ctx, req.User, true); err != nil {
		return nil, err
	}

	if req.User.UserID == "" {
		for i := 0; i < 20; i++ {
			userID := o.genUserID()
			_, err := o.Database.GetUser(ctx, userID)
			if err == nil {
				continue
			} else if dbutil.IsDBNotFound(err) {
				req.User.UserID = userID
				break
			} else {
				return nil, err
			}
		}
		if req.User.UserID == "" {
			return nil, errs.ErrInternalServer.WrapMsg("gen user id failed")
		}
	} else {
		_, err := o.Database.GetUser(ctx, req.User.UserID)
		if err == nil {
			return nil, errs.ErrArgs.WrapMsg("appoint user id already register")
		} else if !dbutil.IsDBNotFound(err) {
			return nil, err
		}
	}

	var (
		credentials     []*chatdb.Credential
		allowChangeRule = datautil.If(req.User.UserType == constant.CommonUser, true, false)
	)

	if req.User.PhoneNumber != "" {
		credentials = append(credentials, &chatdb.Credential{
			UserID:      req.User.UserID,
			Account:     BuildCredentialPhone(req.User.AreaCode, req.User.PhoneNumber),
			Type:        constant.CredentialPhone,
			AllowChange: allowChangeRule,
		})
	}

	if req.User.Account != "" {
		credentials = append(credentials, &chatdb.Credential{
			UserID:      req.User.UserID,
			Account:     req.User.Account,
			Type:        constant.CredentialAccount,
			AllowChange: allowChangeRule,
		})
	}

	if req.User.Email != "" {
		credentials = append(credentials, &chatdb.Credential{
			UserID:      req.User.UserID,
			Account:     req.User.Email,
			Type:        constant.CredentialEmail,
			AllowChange: allowChangeRule,
		})
	}

	register := &chatdb.Register{
		UserID:      req.User.UserID,
		DeviceID:    req.DeviceID,
		IP:          req.Ip,
		Platform:    constantpb.PlatformID2Name[int(req.Platform)],
		AccountType: "",
		Mode:        constant.UserMode,
		CreateTime:  time.Now(),
	}
	account := &chatdb.Account{
		UserID:         req.User.UserID,
		Password:       req.User.Password,
		OperatorUserID: mcontext.GetOpUserID(ctx),
		ChangeTime:     register.CreateTime,
		CreateTime:     register.CreateTime,
	}
	attribute := &chatdb.Attribute{
		UserID:         req.User.UserID,
		Account:        req.User.Account,
		PhoneNumber:    req.User.PhoneNumber,
		AreaCode:       req.User.AreaCode,
		Email:          req.User.Email,
		Nickname:       req.User.Nickname,
		FaceURL:        req.User.FaceURL,
		Gender:         req.User.Gender,
		BirthTime:      time.UnixMilli(req.User.Birth),
		ChangeTime:     register.CreateTime,
		CreateTime:     register.CreateTime,
		AllowVibration: constant.DefaultAllowVibration,
		AllowBeep:      constant.DefaultAllowBeep,
		AllowAddFriend: constant.DefaultAllowAddFriend,
	}

	if req.User.UserType == constant.OrgUser {
		attribute.EnglishName = datautil.ToPtr(req.User.EnglishName.GetValue())
		attribute.Station = datautil.ToPtr(req.User.Station.GetValue())
		attribute.Telephone = datautil.ToPtr(req.User.Telephone.GetValue())
	}
	if err := o.Database.RegisterUser(ctx, register, account, attribute, credentials); err != nil {
		return nil, err
	}
	return &chat.AddUserAccountResp{}, nil
}

func (o *chatSvr) SearchUserPublicInfo(ctx context.Context, req *chat.SearchUserPublicInfoReq) (*chat.SearchUserPublicInfoResp, error) {
	if _, _, err := mctx.Check(ctx); err != nil {
		return nil, err
	}
	total, list, err := o.Database.Search(ctx, constant.FinDAllUser, req.Keyword, req.Genders, req.Pagination)
	if err != nil {
		return nil, err
	}
	return &chat.SearchUserPublicInfoResp{
		Total: uint32(total),
		Users: DbToPbAttributes(list),
	}, nil
}

func (o *chatSvr) SearchUserID(ctx context.Context, req *chat.SearchUserIDReq) (*chat.SearchUserIDResp, error) {
	if req.Pagination == nil {
		return nil, errs.ErrArgs.WrapMsg("pagination is nil")
	}
	if _, _, err := mctx.Check(ctx); err != nil {
		return nil, err
	}
	total, userIDs, err := o.Database.SearchID(ctx, req.Keyword, req.OrUserIDs, req.Pagination)
	if err != nil {
		return nil, err
	}
	return &chat.SearchUserIDResp{
		Total:   uint32(total),
		UserIDs: userIDs,
	}, nil
}

func (o *chatSvr) FindUserFullInfo(ctx context.Context, req *chat.FindUserFullInfoReq) (*chat.FindUserFullInfoResp, error) {
	if _, _, err := mctx.Check(ctx); err != nil {
		return nil, err
	}
	if len(req.UserIDs) == 0 {
		return nil, errs.ErrArgs.WrapMsg("UserIDs is empty")
	}
	attributes, err := o.Database.FindAttribute(ctx, req.UserIDs)
	if err != nil {
		return nil, err
	}
	return &chat.FindUserFullInfoResp{Users: DbToPbUserFullInfos(attributes)}, nil
}

func (o *chatSvr) SearchUserFullInfo(ctx context.Context, req *chat.SearchUserFullInfoReq) (*chat.SearchUserFullInfoResp, error) {
	if _, _, err := mctx.Check(ctx); err != nil {
		return nil, err
	}
	total, list, err := o.Database.Search(ctx, req.Normal, req.Keyword, req.Genders, req.Pagination)
	if err != nil {
		return nil, err
	}
	return &chat.SearchUserFullInfoResp{
		Total: uint32(total),
		Users: DbToPbUserFullInfos(list),
	}, nil
}

func (o *chatSvr) FindUserAccount(ctx context.Context, req *chat.FindUserAccountReq) (*chat.FindUserAccountResp, error) {
	if len(req.UserIDs) == 0 {
		return nil, errs.ErrArgs.WrapMsg("user id list must be set")
	}
	if _, _, err := mctx.CheckAdminOrUser(ctx); err != nil {
		return nil, err
	}
	attributes, err := o.Database.FindAttribute(ctx, req.UserIDs)
	if err != nil {
		return nil, err
	}
	userAccountMap := make(map[string]string)
	for _, attribute := range attributes {
		userAccountMap[attribute.UserID] = attribute.Account
	}
	return &chat.FindUserAccountResp{UserAccountMap: userAccountMap}, nil
}

func (o *chatSvr) FindAccountUser(ctx context.Context, req *chat.FindAccountUserReq) (*chat.FindAccountUserResp, error) {
	if len(req.Accounts) == 0 {
		return nil, errs.ErrArgs.WrapMsg("account list must be set")
	}
	if _, _, err := mctx.CheckAdminOrUser(ctx); err != nil {
		return nil, err
	}
	attributes, err := o.Database.FindAttribute(ctx, req.Accounts)
	if err != nil {
		return nil, err
	}
	accountUserMap := make(map[string]string)
	for _, attribute := range attributes {
		accountUserMap[attribute.Account] = attribute.UserID
	}
	return &chat.FindAccountUserResp{AccountUserMap: accountUserMap}, nil
}

func (o *chatSvr) SearchUserInfo(ctx context.Context, req *chat.SearchUserInfoReq) (*chat.SearchUserInfoResp, error) {
	if _, _, err := mctx.Check(ctx); err != nil {
		return nil, err
	}
	total, list, err := o.Database.SearchUser(ctx, req.Keyword, req.UserIDs, req.Genders, req.Pagination)
	if err != nil {
		return nil, err
	}
	return &chat.SearchUserInfoResp{
		Total: uint32(total),
		Users: DbToPbUserFullInfos(list),
	}, nil
}

func (o *chatSvr) CheckUserExist(ctx context.Context, req *chat.CheckUserExistReq) (resp *chat.CheckUserExistResp, err error) {
	if req.User == nil {
		return nil, errs.ErrArgs.WrapMsg("user is nil")
	}
	if req.User.PhoneNumber != "" {
		account, err := o.Database.TakeCredentialByAccount(ctx, BuildCredentialPhone(req.User.AreaCode, req.User.PhoneNumber))
		// err != nil is not found User
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		if account != nil {
			log.ZDebug(ctx, "Check Number is ", account.Account)
			log.ZDebug(ctx, "Check userID is ", account.UserID)
			return &chat.CheckUserExistResp{Userid: account.UserID, IsRegistered: true}, nil
		}
	}
	if req.User.Email != "" {
		account, err := o.Database.TakeCredentialByAccount(ctx, req.User.AreaCode)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		if account != nil {
			log.ZDebug(ctx, "Check email is ", account.Account)
			log.ZDebug(ctx, "Check userID is ", account.UserID)
			return &chat.CheckUserExistResp{Userid: account.UserID, IsRegistered: true}, nil
		}
	}
	if req.User.Account != "" {
		account, err := o.Database.TakeCredentialByAccount(ctx, req.User.Account)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		if account != nil {
			log.ZDebug(ctx, "Check account is ", account.Account)
			log.ZDebug(ctx, "Check userID is ", account.UserID)
			return &chat.CheckUserExistResp{Userid: account.UserID, IsRegistered: true}, nil
		}
	}
	return nil, nil
}

func (o *chatSvr) DelUserAccount(ctx context.Context, req *chat.DelUserAccountReq) (resp *chat.DelUserAccountResp, err error) {
	if err := o.Database.DelUserAccount(ctx, req.UserIDs); err != nil && errs.Unwrap(err) != mongo.ErrNoDocuments {
		return nil, err
	}
	return nil, nil
}
