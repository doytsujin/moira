package handler

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/moira-alert/moira/api"
	"github.com/moira-alert/moira/api/controller"
	"github.com/moira-alert/moira/api/dto"
	"github.com/moira-alert/moira/api/middleware"
)

func teams(router chi.Router) {
	router.Get("/", getAllTeams)
	router.Post("/", createTeam)
	router.Route("/{teamId}", func(router chi.Router) {
		router.Use(middleware.TeamContext)
		router.Get("/", getTeam)
		router.Patch("/", updateTeam)
		router.Route("/users", func(router chi.Router) {
			router.Post("/", getTeamUsers)
			router.Post("/", addTeamUsers)
			router.With(middleware.TeamUserIDContext).Delete("/{teamUserId}", deleteTeamUser)
		})
	})
}

func createTeam(writer http.ResponseWriter, request *http.Request) {
	team := dto.TeamModel{}
	err := render.Bind(request, team)
	if err != nil {
		render.Render(writer, request, api.ErrorInvalidRequest(err))
		return
	}
	response, apiErr := controller.CreateTeam(database, team)
	if apiErr != nil {
		render.Render(writer, request, apiErr)
		return
	}
	if err := render.Render(writer, request, response); err != nil {
		render.Render(writer, request, api.ErrorRender(err))
		return
	}
}

func getAllTeams(writer http.ResponseWriter, request *http.Request) {
	user := middleware.GetLogin(request)
	response, err := controller.GetUserTeams(database, user)
	if err != nil {
		render.Render(writer, request, err)
		return
	}

	if err := render.Render(writer, request, response); err != nil {
		render.Render(writer, request, api.ErrorRender(err))
		return
	}

}

func getTeam(writer http.ResponseWriter, request *http.Request) {
	teamID := middleware.GetTeamID(request)

	response, err := controller.GetTeam(database, teamID)
	if err != nil {
		render.Render(writer, request, err)
		return
	}

	if err := render.Render(writer, request, response); err != nil {
		render.Render(writer, request, api.ErrorRender(err))
		return
	}
}

func updateTeam(writer http.ResponseWriter, request *http.Request) {
	team := dto.TeamModel{}
	err := render.Bind(request, team)
	if err != nil {
		render.Render(writer, request, api.ErrorInvalidRequest(err))
		return
	}

	teamID := middleware.GetTeamID(request)

	response, apiErr := controller.UpdateTeam(database, teamID, team)
	if apiErr != nil {
		render.Render(writer, request, apiErr)
		return
	}
	if err := render.Render(writer, request, response); err != nil {
		render.Render(writer, request, api.ErrorRender(err))
		return
	}
}

func getTeamUsers(writer http.ResponseWriter, request *http.Request) {
	teamID := middleware.GetTeamID(request)

	response, err := controller.GetTeamUsers(database, teamID)
	if err != nil {
		render.Render(writer, request, err)
		return
	}

	if err := render.Render(writer, request, response); err != nil {
		render.Render(writer, request, api.ErrorRender(err))
		return
	}
}

func addTeamUsers(writer http.ResponseWriter, request *http.Request) {
	members := dto.TeamMembers{}
	render.Bind(request, members)

	teamID := middleware.GetTeamID(request)

	response, err := controller.AddTeamUsers(database, teamID, members.Usernames)
	if err != nil {
		render.Render(writer, request, err)
		return
	}

	if err := render.Render(writer, request, response); err != nil {
		render.Render(writer, request, api.ErrorRender(err))
		return
	}
}

func deleteTeamUser(writer http.ResponseWriter, request *http.Request) {
	teamID := middleware.GetTeamID(request)
	userID := middleware.GetTeamUserID(request)

	response, err := controller.DeleteTeamUser(database, teamID, userID)
	if err != nil {
		render.Render(writer, request, err)
		return
	}

	if err := render.Render(writer, request, response); err != nil {
		render.Render(writer, request, api.ErrorRender(err))
		return
	}
}
