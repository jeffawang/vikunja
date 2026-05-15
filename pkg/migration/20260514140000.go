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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260514140000",
		Description: "create everyone team",
		Migrate: func(tx *xorm.Engine) error {
			s := tx.NewSession()
			defer s.Close()
			if _, err := ensureEveryoneTeam(s); err != nil {
				_ = s.Rollback()
				return err
			}
			return s.Commit()
		},
		Rollback: func(tx *xorm.Engine) error {
			_, err := tx.Where("name = ?", "Everyone").Delete(&models.Team{})
			return err
		},
	})
}
