package httpapi

import (
	"crypto/sha1"
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
	if root := authorityRootCertificate(caOrNil(d)); root != nil {
		out.CASubject = root.Subject.String()
		if root.SerialNumber != nil {
			out.CASerial = root.SerialNumber.Text(16)
		}
		sum := sha1.Sum(root.Raw)
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
	tlsCert := authorityTLSCertificate(caOrNil(d))
	if len(tlsCert.Certificate) == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no CA configured", nil)
		return
	}
	// First cert in chain is leaf (root CA here)
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=network-debugger-dev-ca.crt")
	for _, b := range tlsCert.Certificate {
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
	if root := authorityRootCertificate(caOrNil(d)); root != nil {
		sum := sha1.Sum(root.Raw)
		fp = hex.EncodeToString(sum[:])
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"certPath":    certPath,
		"keyPath":     keyPath,
		"fingerprint": fp,
	})
}

func caOrNil(d *Deps) *CertAuthority {
	if d == nil || d.MITM == nil {
		return nil
	}
	return d.MITM.CA
}
