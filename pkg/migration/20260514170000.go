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

type users20260514170000 struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	Username   string    `xorm:"varchar(250) not null unique"`
	Email      string    `xorm:"varchar(250) null"`
	Password   string    `xorm:"varchar(250) null"`
	BotOwnerID int64     `xorm:"bigint null index"`
	Created    time.Time `xorm:"created"`
	Updated    time.Time `xorm:"updated"`
}

func (users20260514170000) TableName() string {
	return "users"
}

type teams20260514170000 struct {
	ID   int64  `xorm:"bigint autoincr not null unique pk"`
	Name string `xorm:"varchar(250) not null"`
}

func (teams20260514170000) TableName() string {
	return "teams"
}

type projects20260514170000 struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	Title      string    `xorm:"varchar(250) not null"`
	OwnerID    int64     `xorm:"bigint INDEX not null"`
	IsArchived bool      `xorm:"not null default false"`
	Created    time.Time `xorm:"created not null"`
	Updated    time.Time `xorm:"updated not null"`
}

func (projects20260514170000) TableName() string {
	return "projects"
}

type teamProjects20260514170000 struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	TeamID     int64     `xorm:"bigint not null INDEX"`
	ProjectID  int64     `xorm:"bigint not null INDEX"`
	Permission int       `xorm:"bigint INDEX not null default 0"`
	Created    time.Time `xorm:"created not null"`
	Updated    time.Time `xorm:"updated not null"`
}

func (teamProjects20260514170000) TableName() string {
	return "team_projects"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260514170000",
		Description: "create chicken bot user and egg project shared with everyone team",
		Migrate: func(tx *xorm.Engine) error {
			// Create the chicken system bot user with no password or email.
			// BotOwnerID starts as zero (omitted from INSERT → NULL), then is
			// updated to the user's own ID so that IsBot() returns true and
			// interactive login is permanently blocked.
			chicken := &users20260514170000{Username: "chicken"}
			if _, err := tx.Insert(chicken); err != nil {
				return err
			}
			if _, err := tx.ID(chicken.ID).Cols("bot_owner_id").Update(&users20260514170000{BotOwnerID: chicken.ID}); err != nil {
				return err
			}

			// Find the Everyone team (created by migration 20260514140000).
			team := &teams20260514170000{}
			has, err := tx.Where("name = ?", "Everyone").Get(team)
			if err != nil {
				return err
			}
			if !has {
				return nil
			}

			// Create the egg project owned by the chicken bot.
			project := &projects20260514170000{
				Title:   "egg",
				OwnerID: chicken.ID,
			}
			if _, err := tx.Insert(project); err != nil {
				return err
			}

			// Share the egg project with the Everyone team at read/write permission (1).
			_, err = tx.Insert(&teamProjects20260514170000{
				TeamID:     team.ID,
				ProjectID:  project.ID,
				Permission: 1,
			})
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			project := &projects20260514170000{}
			has, err := tx.Where("title = ?", "egg").Get(project)
			if err != nil {
				return err
			}
			if has {
				if _, err := tx.Where("project_id = ?", project.ID).Delete(&teamProjects20260514170000{}); err != nil {
					return err
				}
				if _, err := tx.Where("title = ?", "egg").Delete(&projects20260514170000{}); err != nil {
					return err
				}
			}

			_, err = tx.Where("username = ?", "chicken").Delete(&users20260514170000{})
			return err
		},
	})
}
