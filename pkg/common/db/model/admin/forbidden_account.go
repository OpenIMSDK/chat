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

package admin

import (
	"context"

	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/db/ormutil"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/errs"
	"github.com/OpenIMSDK/chat/pkg/common/db/table/admin"
	"gorm.io/gorm"
)

func NewForbiddenAccount(db *gorm.DB) admin.ForbiddenAccountInterface {
	return &ForbiddenAccount{db: db}
}

type ForbiddenAccount struct {
	db *gorm.DB
}

func (o *ForbiddenAccount) Create(ctx context.Context, ms []*admin.ForbiddenAccount) error {
	return errs.Wrap(o.db.WithContext(ctx).Create(&ms).Error)
}

func (o *ForbiddenAccount) Take(ctx context.Context, userID string) (*admin.ForbiddenAccount, error) {
	var f admin.ForbiddenAccount
	return &f, errs.Wrap(o.db.WithContext(ctx).Where("user_id = ?", userID).Take(&f).Error)
}

// delete a forbidden account
func (o *ForbiddenAccount) Delete(ctx context.Context, userIDs []string) error {
	return errs.Wrap(o.db.WithContext(ctx).Where("user_id in ?", userIDs).Delete(&admin.ForbiddenAccount{}).Error)
}

// find	a forbidden account
func (o *ForbiddenAccount) Find(ctx context.Context, userIDs []string) ([]*admin.ForbiddenAccount, error) {
	var ms []*admin.ForbiddenAccount
	return ms, errs.Wrap(o.db.WithContext(ctx).Where("user_id in ?", userIDs).Find(&ms).Error)
}

// search account
func (o *ForbiddenAccount) Search(ctx context.Context, keyword string, page int32, size int32) (uint32, []*admin.ForbiddenAccount, error) {
	return ormutil.GormSearch[admin.ForbiddenAccount](o.db.WithContext(ctx), []string{"user_id", "reason", "operator_user_id"}, keyword, page, size)
}
