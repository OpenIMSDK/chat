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
	"github.com/OpenIMSDK/chat/pkg/common/constant"
	"github.com/OpenIMSDK/chat/pkg/common/db/table/admin"
	"gorm.io/gorm"
)

func NewApplet(db *gorm.DB) *Applet {
	return &Applet{
		db: db,
	}
}

type Applet struct {
	db *gorm.DB
}

// create applet
func (o *Applet) Create(ctx context.Context, applets ...*admin.Applet) error {
	return errs.Wrap(o.db.WithContext(ctx).Create(&applets).Error)
}

// delete
func (o *Applet) Del(ctx context.Context, ids []string) error {
	return errs.Wrap(o.db.WithContext(ctx).Where("id in (?)", ids).Delete(&admin.Applet{}).Error)
}

// update applet
func (o *Applet) Update(ctx context.Context, id string, data map[string]any) error {
	return errs.Wrap(o.db.WithContext(ctx).Model(&admin.Applet{}).Where("id = ?", id).Updates(data).Error)
}

func (o *Applet) Take(ctx context.Context, id string) (*admin.Applet, error) {
	var a admin.Applet
	return &a, errs.Wrap(o.db.WithContext(ctx).Where("id = ?", id).Take(&a).Error)
}

// search applet
func (o *Applet) Search(ctx context.Context, keyword string, page int32, size int32) (uint32, []*admin.Applet, error) {
	return ormutil.GormSearch[admin.Applet](o.db.WithContext(ctx), []string{"name", "id", "app_id", "version"}, keyword, page, size)
}

// find on shelf
func (o *Applet) FindOnShelf(ctx context.Context) ([]*admin.Applet, error) {
	var ms []*admin.Applet
	return ms, errs.Wrap(o.sort(o.db).Where("status = ?", constant.StatusOnShelf).Find(&ms).Error)
}

// find ID
func (o *Applet) FindID(ctx context.Context, ids []string) ([]*admin.Applet, error) {
	var ms []*admin.Applet
	return ms, errs.Wrap(o.sort(o.db).Where("id in (?)", ids).Find(&ms).Error)
}

// sort
func (o *Applet) sort(db *gorm.DB) *gorm.DB {
	return db.Order("priority desc, create_time desc")
}
