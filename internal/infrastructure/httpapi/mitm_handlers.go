package httpapi

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// handleV1MITMStatus: GET returns current MITM config/status
func (d *Deps) handleV1MITMStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use GET", nil)
		return
	}
	type resp struct {
		Enabled       bool     `json:"enabled"`
		HasCA         bool     `json:"hasCA"`
		Allow         []string `json:"allow,omitempty"`
		Deny          []string `json:"deny,omitempty"`
		CASubject     string   `json:"caSubject,omitempty"`
		CASerial      string   `json:"caSerial,omitempty"`
		CAFingerprint string   `json:"caFingerprint,omitempty"`
		CACertPath    string   `json:"caCertPath,omitempty"`
	}
	out := resp{Enabled: d.Cfg.MITMEnabled, HasCA: d.MITM != nil && d.MITM.CA != nil, Allow: d.Cfg.MITMDomainsAllow, Deny: d.Cfg.MITMDomainsDeny, CACertPath: d.Cfg.MITMCACertFile}
	if d.MITM != nil && d.MITM.CA != nil && d.MITM.CA.caCert != nil {
		out.CASubject = d.MITM.CA.caCert.Subject.String()
		if d.MITM.CA.caCert.SerialNumber != nil {
			out.CASerial = d.MITM.CA.caCert.SerialNumber.Text(16)
		}
		sum := sha1.Sum(d.MITM.CA.caCert.Raw)
		out.CAFingerprint = hex.EncodeToString(sum[:])
	}
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

// handleV1MITMRegeneratePersist: POST generates a new dev CA, persists to files
// (using configured paths or ./data defaults) and swaps runtime CA.
func (d *Deps) handleV1MITMRegeneratePersist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use POST", nil)
		return
	}
	base := "data"
	_ = os.MkdirAll(base, 0o755)
	certPath := d.Cfg.MITMCACertFile
	keyPath := d.Cfg.MITMCAKeyFile
	if certPath == "" {
		certPath = filepath.Join(base, "mitm_dev_ca.crt")
	}
	if keyPath == "" {
		keyPath = filepath.Join(base, "mitm_dev_ca.key")
	}
	certPEM, keyPEM, err := GenerateDevCA("network-debugger dev CA", 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "CA_GENERATE_FAILED", err.Error(), nil)
		return
	}
	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, "CA_WRITE_FAILED", err.Error(), nil)
		return
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		writeError(w, http.StatusInternalServerError, "CA_WRITE_FAILED", err.Error(), nil)
		return
	}
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err == nil {
		if d.MITM == nil {
			d.MITM = &MITM{}
		}
		d.MITM.CA = ca
	}
	// update cfg paths so status shows them
	d.Cfg.MITMCACertFile = certPath
	d.Cfg.MITMCAKeyFile = keyPath
	fp := ""
	if d.MITM != nil && d.MITM.CA != nil && d.MITM.CA.caCert != nil {
		sum := sha1.Sum(d.MITM.CA.caCert.Raw)
		fp = hex.EncodeToString(sum[:])
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"certPath":    certPath,
		"keyPath":     keyPath,
		"fingerprint": fp,
	})
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
