package request

import (
	"encoding/json"
	"net/http"

	"github.com/taiwoajasa245/torah_ai_backend/pkg/response"
)

func DecodeJson(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body", err.Error())
		return false
	}
	return true
}
