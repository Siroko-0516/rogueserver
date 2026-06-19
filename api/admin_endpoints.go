package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func checkAdminSecret(r *http.Request) bool {
	secret := os.Getenv("ADMIN_SECRET_KEY")
	if secret == "" {
		return false
	}
	return r.Header.Get("X-Admin-Key") == secret
}

// POST /admin/givemoney?username=xxx&amount=999999
func handleAdminGiveMoney(w http.ResponseWriter, r *http.Request) {
	if !checkAdminSecret(r) {
		httpError(w, r, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		httpError(w, r, fmt.Errorf("missing username"), http.StatusBadRequest)
		return
	}

	amountStr := r.URL.Query().Get("amount")
	if amountStr == "" {
		amountStr = "9999999"
	}

	var amount int
	fmt.Sscanf(amountStr, "%d", &amount)

	uuid, err := db.Store.FetchUUIDFromUsername(username)
	if err != nil {
		httpError(w, r, fmt.Errorf("user not found: %s", err), http.StatusNotFound)
		return
	}

	system, err := db.Store.ReadSystemSaveData(uuid)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to read save data: %s", err), http.StatusInternalServerError)
		return
	}

	system.GameStats = nil

	err = db.Store.StoreSystemSaveData(uuid, system)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to store save data: %s", err), http.StatusInternalServerError)
		return
	}

	// Also update session save data money
	for slot := 0; slot < 5; slot++ {
		session, err := db.Store.ReadSessionSaveData(uuid, slot)
		if err != nil {
			continue
		}
		session.Money = amount
		db.Store.StoreSessionSaveData(uuid, session, slot)
	}

	writeJSON(w, r, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("gave %d money to %s", amount, username),
	})
}

// POST /admin/unlockalldex?username=xxx
func handleAdminUnlockAllDex(w http.ResponseWriter, r *http.Request) {
	if !checkAdminSecret(r) {
		httpError(w, r, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		httpError(w, r, fmt.Errorf("missing username"), http.StatusBadRequest)
		return
	}

	uuid, err := db.Store.FetchUUIDFromUsername(username)
	if err != nil {
		httpError(w, r, fmt.Errorf("user not found: %s", err), http.StatusNotFound)
		return
	}

	system, err := db.Store.ReadSystemSaveData(uuid)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to read save data: %s", err), http.StatusInternalServerError)
		return
	}

	if system.DexData == nil {
		system.DexData = make(defs.DexData)
	}

	// Unlock all pokemon 1-1025
	for i := 1; i <= 1025; i++ {
		system.DexData[i] = defs.DexEntry{
			SeenAttr:     "264191",
			CaughtAttr:   "264191",
			NatureAttr:   67108863,
			SeenCount:    999,
			CaughtCount:  999,
			HatchedCount: 0,
			Ivs:          []int{31, 31, 31, 31, 31, 31},
			Ribbons:      "",
		}
	}

	err = db.Store.StoreSystemSaveData(uuid, system)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to store save data: %s", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, r, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("unlocked all dex for %s", username),
	})
}

// GET /admin/getsave?username=xxx  (for debugging)
func handleAdminGetSave(w http.ResponseWriter, r *http.Request) {
	if !checkAdminSecret(r) {
		httpError(w, r, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		httpError(w, r, fmt.Errorf("missing username"), http.StatusBadRequest)
		return
	}

	uuid, err := db.Store.FetchUUIDFromUsername(username)
	if err != nil {
		httpError(w, r, fmt.Errorf("user not found: %s", err), http.StatusNotFound)
		return
	}

	system, err := db.Store.ReadSystemSaveData(uuid)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to read save data: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(system)
}
