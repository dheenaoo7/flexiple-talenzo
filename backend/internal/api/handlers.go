package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"talent-search-rubric-arm/backend/internal/domain"
	"talent-search-rubric-arm/backend/internal/session"
)

type API struct {
	service *session.Service
}

func New(service *session.Service) *API {
	return &API{service: service}
}

type createSessionRequest struct {
	Query string `json:"query"`
}

type draftPayload struct {
	SessionID string         `json:"session_id"`
	Filters   domain.Filters `json:"filters"`
	Rubric    domain.Rubric  `json:"rubric"`
}

func (a *API) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		writeJSONError(w, http.StatusBadRequest, "query must not be empty")
		return
	}

	sse, ok := newSSEWriter(w)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	filters, rubric, err := a.service.Draft(r.Context(), req.Query, sse.stage)
	if err != nil {
		sse.error(err)
		return
	}

	sess, err := a.service.StartSession(r.Context(), req.Query)
	if err != nil {
		log.Printf("create session: %v", err)
		sse.error(err)
		return
	}

	sse.result("draft", draftPayload{SessionID: sess.ID, Filters: filters, Rubric: rubric})
}

type runRequest struct {
	Filters domain.Filters `json:"filters"`
	Rubric  domain.Rubric  `json:"rubric"`
}

func (a *API) HandleRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req runRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	sse, ok := newSSEWriter(w)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	version, err := a.service.Run(r.Context(), id, req.Filters, req.Rubric, sse.stage)
	if err != nil {
		log.Printf("run session %s: %v", id, err)
		sse.error(err)
		return
	}
	sse.result("version", version)
}

type evolveRequest struct {
	Message   string            `json:"message"`
	Reactions []domain.Reaction `json:"reactions"`
}

func (a *API) HandleEvolve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req evolveRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Message) == "" && len(req.Reactions) == 0 {
		writeJSONError(w, http.StatusBadRequest, "provide a message and/or reactions")
		return
	}

	sse, ok := newSSEWriter(w)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	version, err := a.service.Evolve(r.Context(), id, req.Message, req.Reactions, sse.stage)
	if err != nil {
		log.Printf("evolve session %s: %v", id, err)
		sse.error(err)
		return
	}
	sse.result("version", version)
}

func (a *API) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detail, err := a.service.GetSessionDetail(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (a *API) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := a.service.ListSessions(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not list sessions")
		return
	}
	if sessions == nil {
		sessions = []domain.SessionSummary{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (a *API) HandleFreeze(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess, err := a.service.Freeze(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "could not freeze session: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

type likeSignalRequest struct {
	VersionID      string `json:"version_id"`
	ProfileID      string `json:"profile_id"`
	CriterionID    string `json:"criterion_id"`
	CriterionLabel string `json:"criterion_label"`
	EvidenceText   string `json:"evidence_text"`
}

func (a *API) HandleLikeSignal(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	var req likeSignalRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	sig, err := a.service.LikeSignal(r.Context(), domain.LikedSignal{
		SessionID:      sessionID,
		VersionID:      req.VersionID,
		ProfileID:      req.ProfileID,
		CriterionID:    req.CriterionID,
		CriterionLabel: req.CriterionLabel,
		EvidenceText:   req.EvidenceText,
	})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "could not save liked signal: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sig)
}

func (a *API) HandleUnlikeSignal(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	signalID := r.PathValue("signalId")
	if err := a.service.UnlikeSignal(r.Context(), sessionID, signalID); err != nil {
		writeJSONError(w, http.StatusNotFound, "liked signal not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	profile, ok := a.service.GetProfile(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "profile not found")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (a *API) HandleListSkills(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.service.ListSkills())
}

func (a *API) HandleListTitles(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.service.ListTitles())
}

func (a *API) HandleListLocations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.service.ListLocations())
}

func (a *API) HandleListCompanyTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.service.ListCompanyTypes())
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
