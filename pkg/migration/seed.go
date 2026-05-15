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

	"xorm.io/xorm"
)

// seedSystemData ensures the canonical system records (Everyone team, chicken
// bot user, egg project + default views + buckets, team_projects share) exist.
// Called from initSchema (fresh DB) and from individual migration Migrate funcs
// (upgraded DB). Each step is idempotent.
func seedSystemData(x *xorm.Engine) error {
	s := x.NewSession()
	defer s.Close()

	everyone, err := ensureEveryoneTeam(s)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	chicken, err := ensureChickenUser(s)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := ensureEggProject(s, chicken, everyone.ID); err != nil {
		_ = s.Rollback()
		return err
	}

	return s.Commit()
}

func ensureEveryoneTeam(s *xorm.Session) (*models.Team, error) {
	team := &models.Team{}
	has, err := s.Where("name = ?", "Everyone").Get(team)
	if err != nil {
		return nil, err
	}
	if has {
		return team, nil
	}

	team = &models.Team{
		Name:        "Everyone",
		Description: "A team that represents all users.",
		IsPublic:    true,
	}
	// Bypass Team.CreateNewTeam (would try to add user 0 as a member).
	// MustCols forces created_by_id=0 into the INSERT despite zero-value omission.
	if _, err := s.MustCols("created_by_id").Insert(team); err != nil {
		return nil, err
	}
	return team, nil
}

func ensureChickenUser(s *xorm.Session) (*user.User, error) {
	chicken := &user.User{}
	has, err := s.Where("username = ?", "chicken").Get(chicken)
	if err != nil {
		return nil, err
	}
	if has {
		return chicken, nil
	}

	chicken = &user.User{Username: "chicken"}
	// Bypass user.CreateUser to skip password/email validation for this bot.
	if _, err := s.Insert(chicken); err != nil {
		return nil, err
	}
	// Self-referential BotOwnerID makes IsBot() true → blocks login.
	chicken.BotOwnerID = chicken.ID
	if _, err := s.ID(chicken.ID).Cols("bot_owner_id").Update(chicken); err != nil {
		return nil, err
	}
	return chicken, nil
}

func ensureEggProject(s *xorm.Session, chicken *user.User, everyoneTeamID int64) error {
	project := &models.Project{}
	has, err := s.Where("title = ?", "egg").Get(project)
	if err != nil {
		return err
	}
	if !has {
		project = &models.Project{Title: "egg", OwnerID: chicken.ID}
		if _, err := s.Insert(project); err != nil {
			return err
		}
		// Canonical view+bucket creation — same path the UI uses for new projects.
		if err := models.CreateDefaultViewsForProject(s, project, chicken, true, true); err != nil {
			return err
		}
	}

	count, err := s.Where("team_id = ? AND project_id = ?", everyoneTeamID, project.ID).Count(&models.TeamProject{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err = s.Insert(&models.TeamProject{
		TeamID:     everyoneTeamID,
		ProjectID:  project.ID,
		Permission: models.PermissionAdmin,
	})
	return err
}
