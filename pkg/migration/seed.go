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

	"xorm.io/xorm"
)

// seedSystemData ensures the canonical system records (Everyone team, chicken
// bot user, egg project) exist. It is called both from initSchema (fresh DB)
// and from the individual migration Migrate funcs (upgraded DB) so the data
// always ends up in the database regardless of which code path runs first.
// Every operation is idempotent — safe to call more than once.
func seedSystemData(x *xorm.Engine) error {
	everyoneTeam, err := ensureEveryoneTeam(x)
	if err != nil {
		return err
	}

	chicken, err := ensureChickenUser(x)
	if err != nil {
		return err
	}

	return ensureEggProject(x, chicken.ID, everyoneTeam.ID)
}

// --- local structs (independent of model changes) ---

type seedTeam struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	Name        string    `xorm:"varchar(250) not null"`
	Description string    `xorm:"longtext null"`
	CreatedByID int64     `xorm:"bigint not null INDEX"`
	IsPublic    bool      `xorm:"not null default false"`
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
}

func (seedTeam) TableName() string { return "teams" }

type seedUser struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	Username   string    `xorm:"varchar(250) not null unique"`
	Email      string    `xorm:"varchar(250) null"`
	Password   string    `xorm:"varchar(250) null"`
	BotOwnerID int64     `xorm:"bigint null index"`
	Created    time.Time `xorm:"created"`
	Updated    time.Time `xorm:"updated"`
}

func (seedUser) TableName() string { return "users" }

type seedProject struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	Title      string    `xorm:"varchar(250) not null"`
	OwnerID    int64     `xorm:"bigint INDEX not null"`
	IsArchived bool      `xorm:"not null default false"`
	Created    time.Time `xorm:"created not null"`
	Updated    time.Time `xorm:"updated not null"`
}

func (seedProject) TableName() string { return "projects" }

type seedViewFilter struct {
	Filter string `json:"filter"`
}

type seedProjectView struct {
	ID                      int64           `xorm:"autoincr not null unique pk"`
	Title                   string          `xorm:"varchar(255) not null"`
	ProjectID               int64           `xorm:"not null index"`
	ViewKind                int             `xorm:"not null"`
	Filter                  *seedViewFilter `xorm:"json null default null"`
	Position                float64         `xorm:"double null"`
	BucketConfigurationMode int             `xorm:"default 0"`
	Created                 time.Time       `xorm:"created not null"`
	Updated                 time.Time       `xorm:"updated not null"`
}

func (seedProjectView) TableName() string { return "project_views" }

type seedTeamProject struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	TeamID     int64     `xorm:"bigint not null INDEX"`
	ProjectID  int64     `xorm:"bigint not null INDEX"`
	Permission int       `xorm:"bigint INDEX not null default 0"`
	Created    time.Time `xorm:"created not null"`
	Updated    time.Time `xorm:"updated not null"`
}

func (seedTeamProject) TableName() string { return "team_projects" }

// --- idempotent helpers ---

func ensureEveryoneTeam(x *xorm.Engine) (*seedTeam, error) {
	team := &seedTeam{}
	has, err := x.Where("name = ?", "Everyone").Get(team)
	if err != nil {
		return nil, err
	}
	if has {
		return team, nil
	}

	team = &seedTeam{
		Name:        "Everyone",
		Description: "A team that represents all users.",
		IsPublic:    true,
		CreatedByID: 0,
	}
	// MustCols forces created_by_id=0 into the INSERT despite xorm's zero-value omission.
	if _, err := x.NewSession().MustCols("created_by_id").Insert(team); err != nil {
		return nil, err
	}
	return team, nil
}

func ensureChickenUser(x *xorm.Engine) (*seedUser, error) {
	chicken := &seedUser{}
	has, err := x.Where("username = ?", "chicken").Get(chicken)
	if err != nil {
		return nil, err
	}
	if has {
		return chicken, nil
	}

	chicken = &seedUser{Username: "chicken"}
	if _, err := x.Insert(chicken); err != nil {
		return nil, err
	}
	// Self-referential BotOwnerID makes IsBot() return true, blocking interactive login.
	if _, err := x.ID(chicken.ID).Cols("bot_owner_id").Update(&seedUser{BotOwnerID: chicken.ID}); err != nil {
		return nil, err
	}
	return chicken, nil
}

func ensureEggProject(x *xorm.Engine, ownerID, everyoneTeamID int64) error {
	project := &seedProject{}
	has, err := x.Where("title = ?", "egg").Get(project)
	if err != nil {
		return err
	}
	if !has {
		project = &seedProject{Title: "egg", OwnerID: ownerID}
		if _, err := x.Insert(project); err != nil {
			return err
		}

		views := []*seedProjectView{
			{Title: "List", ProjectID: project.ID, ViewKind: 0, Filter: &seedViewFilter{Filter: "done = false"}, Position: 100},
			{Title: "Gantt", ProjectID: project.ID, ViewKind: 1, Position: 200},
			{Title: "Table", ProjectID: project.ID, ViewKind: 2, Position: 300},
			{Title: "Kanban", ProjectID: project.ID, ViewKind: 3, Position: 400, BucketConfigurationMode: 1},
		}
		if _, err := x.Insert(&views); err != nil {
			return err
		}
	}

	// Ensure the Everyone team has access.
	count, err := x.Where("team_id = ? AND project_id = ?", everyoneTeamID, project.ID).Count(&seedTeamProject{})
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = x.Insert(&seedTeamProject{
			TeamID:     everyoneTeamID,
			ProjectID:  project.ID,
			Permission: 2, // Admin
		})
		return err
	}
	return nil
}
