package handler

import (
	"errors"
	"net/http"
	"strconv"
)

func parseID(r *http.Request) (int, error) {
	idStr := r.PathValue("id")

	if idStr == "" {
		return 0, errors.New("missing id")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
