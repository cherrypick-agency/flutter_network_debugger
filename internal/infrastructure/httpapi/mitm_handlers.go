package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// handleV1MITMStatus: GET returns current MITM config/status
func (d *Deps) handleV1MITMStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use GET", nil)
		return
	}
	type resp struct {
		Enabled bool     `json:"enabled"`
		HasCA   bool     `json:"hasCA"`
		Allow   []string `json:"allow,omitempty"`
		Deny    []string `json:"deny,omitempty"`
	}
	out := resp{Enabled: d.Cfg.MITMEnabled, HasCA: d.MITM != nil && d.MITM.CA != nil, Allow: d.Cfg.MITMDomainsAllow, Deny: d.Cfg.MITMDomainsDeny}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// handleV1MITMGetCA: GET returns current CA certificate (PEM). If CA absent, 404.
func (d *Deps) handleV1MITMGetCA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use GET", nil)
		return
	}
	if d.MITM == nil || d.MITM.CA == nil || len(d.MITM.CA.tlsCert.Certificate) == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no CA configured", nil)
		return
	}
	// First cert in chain is leaf (root CA here)
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=network-debugger-dev-ca.crt")
	for _, b := range d.MITM.CA.tlsCert.Certificate {
		// write all certs as PEM blocks
		if len(b) == 0 {
			continue
		}
		_ = pemEncodeCert(w, b)
	}
}

// handleV1MITMGenerate: POST generates a new dev CA in-memory; returns PEMs.
// This is for convenience; in real conditions CA should be generated beforehand and stored securely.
func (d *Deps) handleV1MITMGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use POST", nil)
		return
	}
	var in struct {
		CN string `json:"cn"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.CN == "" {
		in.CN = "network-debugger dev CA"
	}
	certPEM, keyPEM, err := GenerateDevCA(in.CN, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "CA_GENERATE_FAILED", err.Error(), nil)
		return
	}
	// swap runtime CA (in-memory only)
	if ca, err2 := LoadCertAuthorityFromPEM(certPEM, keyPEM); err2 == nil {
		if d.MITM == nil {
			d.MITM = &MITM{}
		}
		d.MITM.CA = ca
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"certPEM": string(certPEM), "keyPEM": string(keyPEM)})
}

// helper: write DER to PEM
func pemEncodeCert(w http.ResponseWriter, der []byte) error {
	_, err := w.Write([]byte("-----BEGIN CERTIFICATE-----\n"))
	if err != nil {
		return err
	}
	b := make([]byte, base64.StdEncoding.EncodedLen(len(der)))
	base64.StdEncoding.Encode(b, der)
	// wrap at 64 columns
	for i := 0; i < len(b); i += 64 {
		j := i + 64
		if j > len(b) {
			j = len(b)
		}
		if _, err := w.Write(b[i:j]); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	_, err = w.Write([]byte("-----END CERTIFICATE-----\n"))
	return err
}
