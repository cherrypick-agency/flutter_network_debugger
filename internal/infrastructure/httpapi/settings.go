package httpapi

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
)

type responseDelayDTO struct {
    Enabled bool   `json:"enabled"`
    Value   string `json:"value"` // "1500" или диапазон "1000-3000"
}

type settingsDTO struct {
    ResponseDelay responseDelayDTO `json:"responseDelay"`
}

// handleV1Settings — простой рантайм-эндпоинт для чтения/записи настроек прокси.
// Сейчас поддерживается только Response Delay (фикс или диапазон в мс).
func (d *Deps) handleV1Settings(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        rd := responseDelayDTO{}
        if d.Cfg.ResponseDelayMinMs > 0 && d.Cfg.ResponseDelayMaxMs > 0 {
            rd.Enabled = true
            rd.Value = strconv.Itoa(d.Cfg.ResponseDelayMinMs) + "-" + strconv.Itoa(d.Cfg.ResponseDelayMaxMs)
        } else if d.Cfg.ResponseDelayMs > 0 {
            rd.Enabled = true
            rd.Value = strconv.Itoa(d.Cfg.ResponseDelayMs)
        } else {
            rd.Enabled = false
            rd.Value = ""
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(settingsDTO{ResponseDelay: rd})
        return
    case http.MethodPost:
        var in settingsDTO
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
            return
        }
        rd := in.ResponseDelay
        // Выключено — обнуляем всё. Иначе парсим value (число или диапазон min-max в мс)
        if !rd.Enabled || strings.TrimSpace(rd.Value) == "" || strings.TrimSpace(rd.Value) == "0" {
            d.Cfg.ResponseDelayMs = 0
            d.Cfg.ResponseDelayMinMs = 0
            d.Cfg.ResponseDelayMaxMs = 0
        } else {
            v := strings.TrimSpace(rd.Value)
            if strings.Contains(v, "-") {
                parts := strings.SplitN(v, "-", 2)
                minStr := strings.TrimSpace(parts[0])
                maxStr := strings.TrimSpace(parts[1])
                min, err1 := strconv.Atoi(minStr)
                max, err2 := strconv.Atoi(maxStr)
                if err1 != nil || err2 != nil || min < 0 || max < 0 {
                    writeError(w, http.StatusBadRequest, "BAD_VALUE", "value must be number or range like 1000-3000", nil)
                    return
                }
                if max < min { min, max = max, min }
                d.Cfg.ResponseDelayMs = 0
                d.Cfg.ResponseDelayMinMs = min
                d.Cfg.ResponseDelayMaxMs = max
            } else {
                n, err := strconv.Atoi(v)
                if err != nil || n < 0 {
                    writeError(w, http.StatusBadRequest, "BAD_VALUE", "value must be non-negative integer or range", nil)
                    return
                }
                d.Cfg.ResponseDelayMinMs = 0
                d.Cfg.ResponseDelayMaxMs = 0
                d.Cfg.ResponseDelayMs = n
            }
        }

        // Вернём актуальные значения аналогично GET
        w.Header().Set("Content-Type", "application/json")
        cur := settingsDTO{}
        if d.Cfg.ResponseDelayMinMs > 0 && d.Cfg.ResponseDelayMaxMs > 0 {
            cur.ResponseDelay.Enabled = true
            cur.ResponseDelay.Value = strconv.Itoa(d.Cfg.ResponseDelayMinMs) + "-" + strconv.Itoa(d.Cfg.ResponseDelayMaxMs)
        } else if d.Cfg.ResponseDelayMs > 0 {
            cur.ResponseDelay.Enabled = true
            cur.ResponseDelay.Value = strconv.Itoa(d.Cfg.ResponseDelayMs)
        } else {
            cur.ResponseDelay.Enabled = false
            cur.ResponseDelay.Value = ""
        }
        _ = json.NewEncoder(w).Encode(cur)
        return
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
}


