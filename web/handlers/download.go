package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	twitchURL := r.FormValue("twitch_url")

	id, _, err := h.twitch.ID(twitchURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := fmt.Sprintf("Received Twitch URL (this is ID: %s): <a href=\"%s\">%s</a>", id, twitchURL, twitchURL)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, res)
}
