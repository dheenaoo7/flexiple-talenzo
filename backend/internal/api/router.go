package api

import "net/http"

func (a *API) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/sessions", a.HandleCreateSession)
	mux.HandleFunc("GET /api/sessions", a.HandleListSessions)
	mux.HandleFunc("GET /api/sessions/{id}", a.HandleGetSession)
	mux.HandleFunc("POST /api/sessions/{id}/run", a.HandleRun)
	mux.HandleFunc("POST /api/sessions/{id}/evolve", a.HandleEvolve)
	mux.HandleFunc("POST /api/sessions/{id}/freeze", a.HandleFreeze)
	mux.HandleFunc("POST /api/sessions/{id}/liked-signals", a.HandleLikeSignal)
	mux.HandleFunc("DELETE /api/sessions/{id}/liked-signals/{signalId}", a.HandleUnlikeSignal)
	mux.HandleFunc("GET /api/profiles/{id}", a.HandleGetProfile)
	mux.HandleFunc("GET /api/skills", a.HandleListSkills)
	mux.HandleFunc("GET /api/titles", a.HandleListTitles)
	mux.HandleFunc("GET /api/locations", a.HandleListLocations)
	mux.HandleFunc("GET /api/company-types", a.HandleListCompanyTypes)

	return logging(mux)
}
