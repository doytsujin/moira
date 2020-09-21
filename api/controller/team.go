package controller

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/moira-alert/moira"
	"github.com/moira-alert/moira/api"
	"github.com/moira-alert/moira/api/dto"
	"github.com/moira-alert/moira/database"
)

// CreateTeam is a controller function that creates a new team in Moira
func CreateTeam(dataBase moira.Database, team dto.TeamModel) (dto.SaveTeamResponse, *api.ErrorResponse) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return dto.SaveTeamResponse{}, api.ErrorInternalServer(fmt.Errorf("cannot generate id for team: %w", err))
	}
	// TODO(litleleprikon): Investigate do we need check for existence of team with this UUID.
	// In theory probability of two exactly the same UUIDs is miserable.
	teamID := uuid.String()
	err = dataBase.SaveTeam(teamID, team.ToMoiraTeam())
	if err != nil {
		return dto.SaveTeamResponse{}, api.ErrorInternalServer(fmt.Errorf("cannot save team: %w", err))
	}

	return dto.SaveTeamResponse{ID: teamID}, nil
}

// GetTeam is a controller function that returns a team by it's ID
func GetTeam(dataBase moira.Database, teamID string) (dto.TeamModel, *api.ErrorResponse) {
	team, err := dataBase.GetTeam(teamID)

	if err != nil {
		if err == database.ErrNil {
			return dto.TeamModel{}, api.ErrorNotFound(fmt.Sprintf("cannot find team: %s", teamID))
		}
		return dto.TeamModel{}, api.ErrorInternalServer(fmt.Errorf("cannot get team from database: %w", err))
	}

	teamModel := dto.NewTeamModel(team)
	return teamModel, nil
}

// GetUserTeams is a controller function that returns a teams in which user is a member bu user ID
func GetUserTeams(dataBase moira.Database, userID string) (dto.UserTeams, *api.ErrorResponse) {
	teams, err := dataBase.GetUserTeams(userID)

	if err != nil {
		if err == database.ErrNil {
			return dto.UserTeams{}, api.ErrorNotFound(fmt.Sprintf("cannot find user teams: %s", userID))
		}
		return dto.UserTeams{}, api.ErrorInternalServer(fmt.Errorf("cannot get user teams from database: %w", err))
	}

	result := dto.UserTeams{
		Teams: teams,
	}
	return result, nil
}

// GetTeamUsers is a controller function that returns a users of team by team ID
func GetTeamUsers(dataBase moira.Database, teamID string) (dto.TeamMembers, *api.ErrorResponse) {
	users, err := dataBase.GetTeamUsers(teamID)

	if err != nil {
		if err == database.ErrNil {
			return dto.TeamMembers{}, api.ErrorNotFound(fmt.Sprintf("cannot find team users: %s", teamID))
		}
		return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot get team users from database: %w", err))
	}

	result := dto.TeamMembers{
		Usernames: users,
	}
	return result, nil
}

// AddTeamUsers is a controller function that adds a users to certain team
func AddTeamUsers(dataBase moira.Database, teamID string, newUsers []string) (dto.TeamMembers, *api.ErrorResponse) {
	_, err := dataBase.GetTeam(teamID)
	if err != nil {
		if err == database.ErrNil {
			return dto.TeamMembers{}, api.ErrorNotFound(fmt.Sprintf("cannot find team: %s", teamID))
		}
		return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot get team from database: %w", err))
	}

	existingUsers, err := dataBase.GetTeamUsers(teamID)
	if err != nil {
		if err == database.ErrNil {
			return dto.TeamMembers{}, api.ErrorNotFound(fmt.Sprintf("cannot find team users: %s", teamID))
		}
		return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot get team users from database: %w", err))
	}

	teamsMap := map[string][]string{}
	finalUsers := []string{}

	for _, userID := range existingUsers {
		userTeams, err := dataBase.GetUserTeams(userID)
		if err != nil {
			if err == database.ErrNil {
				return dto.TeamMembers{}, api.ErrorNotFound(fmt.Sprintf("cannot find user teams: %s", userID))
			}
			return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot get user teams from database: %w", err))
		}
		teamsMap[userID] = userTeams
		finalUsers = append(finalUsers, userID)
	}

	for _, userID := range newUsers {
		if _, ok := teamsMap[userID]; ok {
			return dto.TeamMembers{}, api.ErrorInvalidRequest(fmt.Errorf("one ore many users you specified are already exist in this team: %s", userID))
		}

		userTeams, err := dataBase.GetUserTeams(userID)
		if err != nil && err != redis.ErrNil {
			return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot get user teams from database: %w", err))
		}

		userTeams, err = addUserTeam(teamID, userTeams)
		if err != nil {
			return dto.TeamMembers{}, api.ErrorInvalidRequest(fmt.Errorf("cannot save new team for user: %s, %w", userID, err))
		}

		teamsMap[userID] = userTeams
		finalUsers = append(finalUsers, userID)
	}

	err = dataBase.SaveTeamsAndUsers(teamID, finalUsers, teamsMap)
	if err != nil {
		api.ErrorInternalServer(fmt.Errorf("cannot save users for team: %s %w", teamID, err))
	}

	result := dto.TeamMembers{
		Usernames: finalUsers,
	}
	return result, nil
}

func addUserTeam(teamID string, teams []string) ([]string, error) {
	for _, currentTeamID := range teams {
		if teamID == currentTeamID {
			return []string{}, fmt.Errorf("team already exist in user teams: %s", teamID)
		}
	}
	teams = append(teams, teamID)
	return teams, nil
}

// UpdateTeam is a controller function that updates an existing team in Moira
func UpdateTeam(dataBase moira.Database, teamID string, team dto.TeamModel) (dto.SaveTeamResponse, *api.ErrorResponse) {
	_, err := dataBase.GetTeam(teamID)
	if err != nil {
		if err == database.ErrNil {
			return dto.SaveTeamResponse{}, api.ErrorNotFound(fmt.Sprintf("cannot find team: %s", teamID))
		}
		return dto.SaveTeamResponse{}, api.ErrorInternalServer(fmt.Errorf("cannot get team from database: %w", err))
	}
	err = dataBase.SaveTeam(teamID, team.ToMoiraTeam())
	if err != nil {
		return dto.SaveTeamResponse{}, api.ErrorInternalServer(fmt.Errorf("cannot save team: %w", err))
	}
	return dto.SaveTeamResponse{ID: teamID}, nil
}

// DeleteTeamUser is a controller function that removes a user from certain team
func DeleteTeamUser(dataBase moira.Database, teamID string, removeUserID string) (dto.TeamMembers, *api.ErrorResponse) {
	_, err := dataBase.GetTeam(teamID)
	if err != nil {
		if err == database.ErrNil {
			return dto.TeamMembers{}, api.ErrorNotFound(fmt.Sprintf("cannot find team: %s", teamID))
		}
		return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot get team from database: %w", err))
	}

	existingUsers, err := dataBase.GetTeamUsers(teamID)
	if err != nil {
		if err == database.ErrNil {
			return dto.TeamMembers{}, api.ErrorNotFound(fmt.Sprintf("cannot find team users: %s", teamID))
		}
		return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot get team users from database: %w", err))
	}

	userFound := false
	for _, userID := range existingUsers {
		if userID == removeUserID {
			userFound = true
		}
	}
	if !userFound {
		return dto.TeamMembers{}, api.ErrorNotFound(fmt.Sprintf("user that you specified not found in this team: %s", removeUserID))
	}

	teamsMap := map[string][]string{}
	finalUsers := []string{}

	for _, userID := range existingUsers {
		userTeams, err := dataBase.GetUserTeams(userID)
		if err != nil {
			if err == database.ErrNil {
				return dto.TeamMembers{}, api.ErrorNotFound(fmt.Sprintf("cannot find user teams: %s", userID))
			}
			return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot get user teams from database: %w", err))
		}
		if userID == removeUserID {
			userTeams, err = removeUserTeam(teamID, userTeams)
			if err != nil {
				return dto.TeamMembers{}, api.ErrorInternalServer(fmt.Errorf("cannot remove team from user: %w", err))
			}
		} else {
			finalUsers = append(finalUsers, userID)
		}
		teamsMap[userID] = userTeams
	}

	err = dataBase.SaveTeamsAndUsers(teamID, finalUsers, teamsMap)
	if err != nil {
		api.ErrorInternalServer(fmt.Errorf("cannot save users for team: %s %w", teamID, err))
	}

	result := dto.TeamMembers{
		Usernames: finalUsers,
	}
	return result, nil
}

func removeUserTeam(teamID string, teams []string) ([]string, error) {
	for i, currentTeamID := range teams {
		if teamID == currentTeamID {
			teams[i] = teams[len(teams)-1]   // Copy last element to index i.
			teams[len(teams)-1] = ""         // Erase last element (write zero value).
			return teams[:len(teams)-1], nil // Truncate slice.
		}
	}
	return []string{}, fmt.Errorf("cannot find team in user teams: %s", teamID)
}
