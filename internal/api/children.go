package api

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"neprodlenka/internal/clock"
	"neprodlenka/internal/store"
)

const defaultPrice = 13000

type childJSON struct {
	ID         int64   `json:"id"`
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName string  `json:"middle_name"`
	Price      int64   `json:"price"`
	Active     bool    `json:"active"`
	ParentIDs  []int64 `json:"parent_ids"`
}

type memberJSON struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}

func toChildJSON(c store.Child) childJSON {
	return childJSON{
		ID:         c.ID,
		LastName:   c.LastName,
		FirstName:  c.FirstName,
		MiddleName: c.MiddleName,
		Price:      c.Price,
		Active:     c.Active,
		ParentIDs:  c.ParentIDs,
	}
}

func (h *handler) listChildren(w http.ResponseWriter, r *http.Request) {
	children, err := h.store.Children(r.Context())
	if err != nil {
		h.internalError(w, err)
		return
	}
	members, err := h.max.Members(r.Context())
	if err != nil {
		h.log.Error("get chat members", zap.Error(err))
		writeError(w, http.StatusBadGateway, "max_unavailable", "MAX недоступен, попробуйте позже")
		return
	}

	res := struct {
		Children []childJSON  `json:"children"`
		Members  []memberJSON `json:"members"`
	}{Children: []childJSON{}, Members: []memberJSON{}}
	for _, c := range children {
		res.Children = append(res.Children, toChildJSON(c))
	}
	for _, m := range members {
		res.Members = append(res.Members, memberJSON{UserID: m.UserID, Name: m.Name})
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *handler) addChild(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LastName   string `json:"last_name"`
		FirstName  string `json:"first_name"`
		MiddleName string `json:"middle_name"`
		Price      *int64 `json:"price"`
	}
	if !readJSON(w, r, &req) {
		return
	}

	c := store.Child{
		LastName:   strings.TrimSpace(req.LastName),
		FirstName:  strings.TrimSpace(req.FirstName),
		MiddleName: strings.TrimSpace(req.MiddleName),
		Price:      defaultPrice,
	}
	if req.Price != nil {
		c.Price = *req.Price
	}
	if err := validateChild(c); err != nil {
		writeError(w, http.StatusBadRequest, "validation", err.Error())
		return
	}

	c, err := h.store.AddChild(r.Context(), c)
	if err != nil {
		h.internalError(w, err)
		return
	}
	h.redrawToday(r)
	writeJSON(w, http.StatusCreated, toChildJSON(c))
}

func (h *handler) updateChild(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Ребёнок не найден")
		return
	}

	var req struct {
		LastName   *string  `json:"last_name"`
		FirstName  *string  `json:"first_name"`
		MiddleName *string  `json:"middle_name"`
		Price      *int64   `json:"price"`
		Active     *bool    `json:"active"`
		ParentIDs  *[]int64 `json:"parent_ids"`
	}
	if !readJSON(w, r, &req) {
		return
	}

	var validationErr error
	c, err := h.store.UpdateChild(r.Context(), id, func(c *store.Child) error {
		if req.LastName != nil {
			c.LastName = strings.TrimSpace(*req.LastName)
		}
		if req.FirstName != nil {
			c.FirstName = strings.TrimSpace(*req.FirstName)
		}
		if req.MiddleName != nil {
			c.MiddleName = strings.TrimSpace(*req.MiddleName)
		}
		if req.Price != nil {
			c.Price = *req.Price
		}
		if req.Active != nil {
			c.Active = *req.Active
		}
		if req.ParentIDs != nil {
			c.ParentIDs = slices.Compact(slices.Sorted(slices.Values(*req.ParentIDs)))
		}
		validationErr = validateChild(*c)
		return validationErr
	})
	switch {
	case validationErr != nil:
		writeError(w, http.StatusBadRequest, "validation", validationErr.Error())
		return
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Ребёнок не найден")
		return
	case err != nil:
		h.internalError(w, err)
		return
	}

	h.redrawToday(r)
	writeJSON(w, http.StatusOK, toChildJSON(c))
}

func validateChild(c store.Child) error {
	switch {
	case c.LastName == "":
		return errors.New("Укажите фамилию")
	case c.FirstName == "":
		return errors.New("Укажите имя")
	case c.Price < 0:
		return errors.New("Цена не может быть отрицательной")
	}
	return nil
}

func (h *handler) redrawToday(r *http.Request) {
	date := clock.Today(h.now())
	post, ok, err := h.store.Post(r.Context(), date)
	if err != nil {
		h.log.Error("get today post", zap.Error(err))
		return
	}
	if ok && post.Status == store.PostOpen {
		h.redrawer.Redraw(date)
	}
}
