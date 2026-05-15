// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type teams20260514140000 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	Name        string    `xorm:"varchar(250) not null"`
	Description string    `xorm:"longtext null"`
	CreatedByID int64     `xorm:"bigint not null INDEX"`
	ExternalID  string    `xorm:"varchar(250) null"`
	Issuer      string    `xorm:"text null"`
	IsPublic    bool      `xorm:"not null default false"`
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
}

func (teams20260514140000) TableName() string {
	return "teams"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260514140000",
		Description: "create everyone team",
		Migrate: func(tx *xorm.Engine) error {
			_, err := tx.Insert(&teams20260514140000{
				Name:        "Everyone",
				Description: "A team that represents all users.",
				IsPublic:    true,
			})
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			_, err := tx.Where("name = ?", "Everyone").Delete(&teams20260514140000{})
			return err
		},
	})
}
