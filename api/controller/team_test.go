package controller

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gomodule/redigo/redis"
	"github.com/moira-alert/moira"
	"github.com/moira-alert/moira/api"
	"github.com/moira-alert/moira/api/dto"
	mock_moira_alert "github.com/moira-alert/moira/mock/moira-alert"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateTeam(t *testing.T) {
	Convey("CreateTeam", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		dataBase := mock_moira_alert.NewMockDatabase(mockCtrl)

		team := dto.TeamModel{Name: "testTeam", Description: "test team description"}

		Convey("create successfully", func() {
			ID := ""
			dataBase.EXPECT().SaveTeam(gomock.Any(), team.ToMoiraTeam()).DoAndReturn(func(teamID string, moiraTeam moira.Team) error {
				ID = teamID
				return nil
			})
			response, err := CreateTeam(dataBase, team)
			So(response.ID, ShouldResemble, ID)
			So(err, ShouldBeNil)
		})

		Convey("save error", func() {
			returnErr := fmt.Errorf("unexpected error")
			dataBase.EXPECT().SaveTeam(gomock.Any(), team.ToMoiraTeam()).Return(returnErr)
			response, err := CreateTeam(dataBase, team)
			So(response, ShouldResemble, dto.SaveTeamResponse{})
			So(err, ShouldResemble, api.ErrorInternalServer(fmt.Errorf("cannot save team: %w", returnErr)))
		})
	})
}

func TestGetTeam(t *testing.T) {
	Convey("GetTeam", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		dataBase := mock_moira_alert.NewMockDatabase(mockCtrl)

		teamID := "testTeam"
		team := moira.Team{Name: "testTeam", Description: "test team description"}

		Convey("get successfully", func() {
			dataBase.EXPECT().GetTeam(teamID).Return(team, nil)
			response, err := GetTeam(dataBase, teamID)
			So(response, ShouldResemble, dto.NewTeamModel(team))
			So(err, ShouldBeNil)
		})

		Convey("team not found", func() {
			dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, redis.ErrNil)
			response, err := GetTeam(dataBase, teamID)
			So(response, ShouldResemble, dto.TeamModel{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find team: testTeam"))
		})

		Convey("database error", func() {
			returnErr := fmt.Errorf("unexpected error")
			dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, returnErr)
			response, err := GetTeam(dataBase, teamID)
			So(response, ShouldResemble, dto.TeamModel{})
			So(err, ShouldResemble, api.ErrorInternalServer(fmt.Errorf("cannot get team from database: %w", returnErr)))
		})
	})
}

func TestGetUserTeams(t *testing.T) {
	Convey("GetUserTeams", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		dataBase := mock_moira_alert.NewMockDatabase(mockCtrl)

		userID := "userID"
		teams := []string{"team1", "team2"}

		Convey("get successfully", func() {
			dataBase.EXPECT().GetUserTeams(userID).Return(teams, nil)
			response, err := GetUserTeams(dataBase, userID)
			So(response, ShouldResemble, dto.UserTeams{Teams: teams})
			So(err, ShouldBeNil)
		})

		Convey("teams not found", func() {
			dataBase.EXPECT().GetUserTeams(userID).Return([]string{}, redis.ErrNil)
			response, err := GetUserTeams(dataBase, userID)
			So(response, ShouldResemble, dto.UserTeams{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find user teams: userID"))
		})

		Convey("database error", func() {
			returnErr := fmt.Errorf("unexpected error")
			dataBase.EXPECT().GetUserTeams(userID).Return([]string{}, returnErr)
			response, err := GetUserTeams(dataBase, userID)
			So(response, ShouldResemble, dto.UserTeams{})
			So(err, ShouldResemble, api.ErrorInternalServer(fmt.Errorf("cannot get user teams from database: %w", returnErr)))
		})
	})
}

func TestGetTeamUsers(t *testing.T) {
	Convey("GetTeamUsers", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		dataBase := mock_moira_alert.NewMockDatabase(mockCtrl)

		teamID := "testTeam"
		users := []string{"userID1", "userID2"}

		Convey("get successfully", func() {
			dataBase.EXPECT().GetTeamUsers(teamID).Return(users, nil)
			response, err := GetTeamUsers(dataBase, teamID)
			So(response, ShouldResemble, dto.TeamMembers{Usernames: users})
			So(err, ShouldBeNil)
		})

		Convey("users not found", func() {
			dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{}, redis.ErrNil)
			response, err := GetTeamUsers(dataBase, teamID)
			So(response, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find team users: testTeam"))
		})

		Convey("database error", func() {
			returnErr := fmt.Errorf("unexpected error")
			dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{}, returnErr)
			response, err := GetTeamUsers(dataBase, teamID)
			So(response, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorInternalServer(fmt.Errorf("cannot get team users from database: %w", returnErr)))
		})
	})
}

func TestAddTeamUsers(t *testing.T) {
	Convey("AddTeamUsers", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		dataBase := mock_moira_alert.NewMockDatabase(mockCtrl)

		teamID := "testTeam"
		userID := "userID"
		userID2 := "userID2"
		userID3 := "userID3"

		Convey("add successfully", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, nil),
				dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{userID, userID2}, nil),
				dataBase.EXPECT().GetUserTeams(userID).Return([]string{teamID}, nil),
				dataBase.EXPECT().GetUserTeams(userID2).Return([]string{teamID}, nil),
				dataBase.EXPECT().GetUserTeams(userID3).Return([]string{}, nil),
				dataBase.EXPECT().SaveTeamsAndUsers(teamID,
					[]string{userID, userID2, userID3},
					map[string][]string{
						userID:  {teamID},
						userID2: {teamID},
						userID3: {teamID},
					},
				).Return(nil),
			)
			response, err := AddTeamUsers(dataBase, teamID, []string{userID3})
			So(response, ShouldResemble, dto.TeamMembers{Usernames: []string{userID, userID2, userID3}})
			So(err, ShouldBeNil)
		})

		Convey("team not found", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, redis.ErrNil),
			)
			response, err := AddTeamUsers(dataBase, teamID, []string{userID3})
			So(response, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find team: testTeam"))
		})

		Convey("team users not found", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, nil),
				dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{}, redis.ErrNil),
			)
			response, err := AddTeamUsers(dataBase, teamID, []string{userID3})
			So(response, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find team users: testTeam"))
		})

		Convey("user teams not found", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, nil),
				dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{userID, userID2}, nil),
				dataBase.EXPECT().GetUserTeams(userID).Return([]string{}, redis.ErrNil),
			)
			response, err := AddTeamUsers(dataBase, teamID, []string{userID3})
			So(response, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find user teams: userID"))
		})

		Convey("user already exists", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, nil),
				dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{userID, userID2, userID3}, nil),
				dataBase.EXPECT().GetUserTeams(userID).Return([]string{teamID}, nil),
				dataBase.EXPECT().GetUserTeams(userID2).Return([]string{teamID}, nil),
				dataBase.EXPECT().GetUserTeams(userID3).Return([]string{teamID}, nil),
			)
			response, err := AddTeamUsers(dataBase, teamID, []string{userID3})
			So(response, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorInvalidRequest(fmt.Errorf("one ore many users you specified are already exist in this team: userID3")))
		})

	})
}

func Test_addUserTeam(t *testing.T) {
	Convey("addUserTeam", t, func() {
		Convey("add successfully", func() {
			actual, err := addUserTeam("testTeam3", []string{"testTeam", "testTeam2"})
			So(actual, ShouldResemble, []string{"testTeam", "testTeam2", "testTeam3"})
			So(err, ShouldBeNil)
		})

		Convey("team already exists", func() {
			actual, err := addUserTeam("testTeam", []string{"testTeam", "testTeam2"})
			So(actual, ShouldResemble, []string{})
			So(err, ShouldResemble, fmt.Errorf("team already exist in user teams: testTeam"))
		})
	})
}

func TestUpdateTeam(t *testing.T) {
	Convey("UpdateTeam", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		dataBase := mock_moira_alert.NewMockDatabase(mockCtrl)

		teamID := "testTeam"
		team := dto.TeamModel{Name: "testTeam", Description: "test team description"}

		Convey("update successfully", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(team.ToMoiraTeam(), nil),
				dataBase.EXPECT().SaveTeam(teamID, team.ToMoiraTeam()).Return(nil),
			)
			response, err := UpdateTeam(dataBase, teamID, team)
			So(response.ID, ShouldResemble, teamID)
			So(err, ShouldBeNil)
		})

		Convey("team not found", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(dto.TeamModel{}.ToMoiraTeam(), redis.ErrNil),
			)
			response, err := UpdateTeam(dataBase, teamID, team)
			So(response, ShouldResemble, dto.SaveTeamResponse{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find team: testTeam"))
		})
	})
}

func TestDeleteTeamUser(t *testing.T) {
	Convey("DeleteTeamUser", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		dataBase := mock_moira_alert.NewMockDatabase(mockCtrl)

		teamID := "testTeam"
		userID := "userID"
		userID2 := "userID2"
		userID3 := "userID3"

		Convey("user exists", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, nil),
				dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{userID, userID2, userID3}, nil),
				dataBase.EXPECT().GetUserTeams(userID).Return([]string{teamID}, nil),
				dataBase.EXPECT().GetUserTeams(userID2).Return([]string{teamID}, nil),
				dataBase.EXPECT().GetUserTeams(userID3).Return([]string{teamID}, nil),
				dataBase.EXPECT().SaveTeamsAndUsers(teamID, []string{userID2, userID3}, map[string][]string{
					userID:  {},
					userID2: {teamID},
					userID3: {teamID},
				}).Return(nil),
			)
			reply, err := DeleteTeamUser(dataBase, teamID, userID)
			So(reply, ShouldResemble, dto.TeamMembers{Usernames: []string{userID2, userID3}})
			So(err, ShouldBeNil)
		})
		Convey("team does not exists", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, redis.ErrNil),
			)
			reply, err := DeleteTeamUser(dataBase, teamID, userID)
			So(reply, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find team: testTeam"))
		})
		Convey("team does not have any users", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, nil),
				dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{}, redis.ErrNil),
			)
			reply, err := DeleteTeamUser(dataBase, teamID, userID)
			So(reply, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find team users: testTeam"))
		})
		Convey("team does not contain user", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, nil),
				dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{userID2, userID3}, nil),
			)
			reply, err := DeleteTeamUser(dataBase, teamID, userID)
			So(reply, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorNotFound("user that you specified not found in this team: userID"))
		})
		Convey("one user do not have teams", func() {
			gomock.InOrder(
				dataBase.EXPECT().GetTeam(teamID).Return(moira.Team{}, nil),
				dataBase.EXPECT().GetTeamUsers(teamID).Return([]string{userID, userID2, userID3}, nil),
				dataBase.EXPECT().GetUserTeams(userID).Return([]string{}, redis.ErrNil),
			)
			reply, err := DeleteTeamUser(dataBase, teamID, userID)
			So(reply, ShouldResemble, dto.TeamMembers{})
			So(err, ShouldResemble, api.ErrorNotFound("cannot find user teams: userID"))
		})

	})
}

func Test_removeUserTeam(t *testing.T) {
	Convey("removeUserTeam", t, func() {
		Convey("remove successfully", func() {
			actual, err := removeUserTeam("testTeam", []string{"testTeam", "testTeam2"})
			So(actual, ShouldResemble, []string{"testTeam2"})
			So(err, ShouldBeNil)
		})

		Convey("team not found", func() {
			actual, err := removeUserTeam("testTeam", []string{"testTeam1", "testTeam2"})
			So(actual, ShouldResemble, []string{})
			So(err, ShouldResemble, fmt.Errorf("cannot find team in user teams: testTeam"))
		})
	})
}
