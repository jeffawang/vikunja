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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260514170000",
		Description: "create chicken bot user and egg project shared with everyone team",
		Migrate: func(tx *xorm.Engine) error {
			return seedSystemData(tx)
		},
		Rollback: func(tx *xorm.Engine) error {
			project := &models.Project{}
			has, err := tx.Where("title = ?", "egg").Get(project)
			if err != nil {
				return err
			}
			if has {
				if _, err := tx.Where("project_id = ?", project.ID).Delete(&models.TeamProject{}); err != nil {
					return err
				}
				viewIDs := []int64{}
				if err := tx.Table("project_views").Where("project_id = ?", project.ID).Cols("id").Find(&viewIDs); err != nil {
					return err
				}
				if len(viewIDs) > 0 {
					if _, err := tx.In("project_view_id", viewIDs).Delete(&models.Bucket{}); err != nil {
						return err
					}
					if _, err := tx.In("id", viewIDs).Delete(&models.ProjectView{}); err != nil {
						return err
					}
				}
				if _, err := tx.Where("title = ?", "egg").Delete(&models.Project{}); err != nil {
					return err
				}
			}
			_, err = tx.Where("username = ?", "chicken").Delete(&user.User{})
			return err
		},
	})
}
