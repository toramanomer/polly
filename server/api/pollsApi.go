package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/toramanomer/polly/primitives"
	"github.com/toramanomer/polly/repository"
)

type createPollRequest struct {
	Question  primitives.Question `json:"question"`
	Options   []string            `json:"options"`
	ExpiresAt time.Time           `json:"expiresAt"`
}

func (req *createPollRequest) validate() map[string][]string {
	errs := make(map[string][]string)

	if questionErrors := req.Question.Validate(); questionErrors != nil {
		errs["question"] = questionErrors
	}

	if len(req.Options) < 2 {
		errs["options"] = append(errs["options"], "At least 2 options are required")
	}

	if len(req.Options) > 6 {
		errs["options"] = append(errs["options"], "A maximum of 6 options are allowed")
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (api *API) CreatePoll(w http.ResponseWriter, r *http.Request) {
	var request createPollRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"type":  "invalid_request_body",
			"title": "The request body is invalid.",
		})
		return
	}

	if errs := request.validate(); errs != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"type":   "validation_error",
			"errors": errs,
		})
		return
	}

	poll := repository.NewPoll(repository.NewPollParams{
		UserID:    ResolveUserID(r),
		Question:  request.Question,
		ExpiresAt: request.ExpiresAt,
		Options:   request.Options,
	})

	if err := api.repository.CreatePollWithOptions(r.Context(), poll); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"type":  "internal_server_error",
			"title": "An internal server error occurred.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(poll)
}

func (api *API) GetPollByID(w http.ResponseWriter, r *http.Request) {
	pollID, err := uuid.Parse(chi.URLParam(r, "pollID"))
	if err != nil || uuid.Nil == pollID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"type":  "invalid_request",
			"title": "Poll ID is not a valid",
		})
		return
	}

	poll, err := api.repository.GetPollWithOptions(r.Context(), pollID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"type":  "internal_server_error",
			"title": "An internal server error occurred.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(poll)
}

type voteOnPollRequest struct {
	OptionID uuid.UUID `json:"optionID"`
}

func (req *voteOnPollRequest) validate() map[string][]string {
	errs := make(map[string][]string)

	if req.OptionID == uuid.Nil {
		errs["optionID"] = append(errs["optionID"], "Option ID is not a valid")
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (api *API) VoteOnPoll(w http.ResponseWriter, r *http.Request) {
	pollID, err := uuid.Parse(chi.URLParam(r, "pollID"))
	if err != nil || uuid.Nil == pollID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"type":  "invalid_request",
			"title": "Poll ID is not a valid",
		})
		return
	}

	var request voteOnPollRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"type":  "invalid_request",
			"title": "Request body is not valid",
		})
		return
	}

	if errs := request.validate(); errs != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"type":   "validation_error",
			"errors": errs,
		})
		return
	}

	vote := repository.NewVote(pollID, request.OptionID)
	if err := api.repository.RecordVote(r.Context(), vote); err != nil {
		switch {
		case errors.Is(err, repository.ErrPollNotFound):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"type":  "not_found",
				"title": "Poll not found",
			})
		case errors.Is(err, repository.ErrOptionBelongsToPoll):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"type":  "invalid_request",
				"title": "Option does not belong to the poll",
			})

		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"type":  "internal_server_error",
				"title": "An internal server error occurred.",
			})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Vote recorded successfully",
	})
}

func (api *API) GetUserPolls(w http.ResponseWriter, r *http.Request) {
	userID := ResolveUserID(r)

	polls, err := api.repository.GetUserPollsWithStats(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"type":  "internal_server_error",
			"title": "An internal server error occurred.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(polls)
}
