package api

import (
	"cmp"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"

	"neprodlenka/internal/clock"
	"neprodlenka/internal/store"
)

const (
	defaultPrice    = 13000
	maxClubNameLen  = 100
	clubStartLayout = "15:04"
)

type childJSON struct {
	ID         int64      `json:"id"`
	LastName   string     `json:"last_name"`
	FirstName  string     `json:"first_name"`
	MiddleName string     `json:"middle_name"`
	Price      int64      `json:"price"`
	Active     bool       `json:"active"`
	Weekdays   []int      `json:"weekdays"`
	ParentIDs  []int64    `json:"parent_ids"`
	Clubs      []clubJSON `json:"clubs"`
}

type clubJSON struct {
	Weekday int    `json:"weekday"`
	Start   string `json:"start"`
	Name    string `json:"name"`
}

type memberJSON struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}

func toChildJSON(c store.Child) childJSON {
	clubs := []clubJSON{}
	for _, club := range c.Clubs {
		clubs = append(clubs, clubJSON(club))
	}
	return childJSON{
		ID:         c.ID,
		LastName:   c.LastName,
		FirstName:  c.FirstName,
		MiddleName: c.MiddleName,
		Price:      c.Price,
		Active:     c.Active,
		Weekdays:   c.Weekdays,
		ParentIDs:  c.ParentIDs,
		Clubs:      clubs,
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
		LastName   *string     `json:"last_name"`
		FirstName  *string     `json:"first_name"`
		MiddleName *string     `json:"middle_name"`
		Price      *int64      `json:"price"`
		Active     *bool       `json:"active"`
		Weekdays   *[]int      `json:"weekdays"`
		ParentIDs  *[]int64    `json:"parent_ids"`
		Clubs      *[]clubJSON `json:"clubs"`
	}
	if !readJSON(w, r, &req) {
		return
	}

	var weekdays []int
	if req.Weekdays != nil {
		weekdays, err = normalizeWeekdays(*req.Weekdays)
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation", err.Error())
			return
		}
	}

	var clubs []store.Club
	if req.Clubs != nil {
		clubs, err = normalizeClubs(*req.Clubs)
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation", err.Error())
			return
		}
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
		if req.Weekdays != nil {
			c.Weekdays = weekdays
		}
		if req.ParentIDs != nil {
			ids := slices.Clone(*req.ParentIDs)
			slices.Sort(ids)
			c.ParentIDs = slices.Compact(ids)
		}
		if req.Clubs != nil {
			c.Clubs = clubs
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

func normalizeWeekdays(in []int) ([]int, error) {
	weekdays := slices.Clone(in)
	slices.Sort(weekdays)
	weekdays = slices.Compact(weekdays)
	switch {
	case len(weekdays) == 0:
		return nil, errors.New("выберите хотя бы один день посещения")
	case weekdays[0] < 1 || weekdays[len(weekdays)-1] > 7:
		return nil, errors.New("неверный день недели")
	}
	return weekdays, nil
}

func normalizeClubs(in []clubJSON) ([]store.Club, error) {
	clubs := []store.Club{}
	for _, c := range in {
		if c.Weekday < 1 || c.Weekday > 7 {
			return nil, errors.New("укажите день недели кружка")
		}
		start, name, err := normalizeClub(c.Start, c.Name)
		if err != nil {
			return nil, err
		}
		clubs = append(clubs, store.Club{Weekday: c.Weekday, Start: start, Name: name})
	}
	slices.SortFunc(clubs, func(a, b store.Club) int {
		return cmp.Or(cmp.Compare(a.Weekday, b.Weekday), strings.Compare(a.Start, b.Start), strings.Compare(a.Name, b.Name))
	})
	return slices.Compact(clubs), nil
}

func normalizeClub(start, name string) (string, string, error) {
	name, err := normalizeClubName(name)
	if err != nil {
		return "", "", err
	}
	t, err := time.Parse(clubStartLayout, strings.TrimSpace(start))
	if err != nil {
		return "", "", errors.New("укажите время начала кружка в формате ЧЧ:ММ")
	}
	return t.Format(clubStartLayout), name, nil
}

func normalizeClubName(name string) (string, error) {
	name = strings.TrimSpace(name)
	switch {
	case name == "":
		return "", errors.New("укажите название кружка")
	case utf8.RuneCountInString(name) > maxClubNameLen:
		return "", errors.New("слишком длинное название кружка")
	}
	return name, nil
}

func validateChild(c store.Child) error {
	switch {
	case c.LastName == "":
		return errors.New("укажите фамилию")
	case c.FirstName == "":
		return errors.New("укажите имя")
	case c.Price < 0:
		return errors.New("цена не может быть отрицательной")
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
