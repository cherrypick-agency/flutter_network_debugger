package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"network-debugger/internal/features/scripting/domain"
)

type scriptArchiveDownload struct {
	Filename string
	Data     []byte
}

type scriptArchiveService struct {
	service *ScriptHandlers
}

func newScriptArchiveService(h *ScriptHandlers) scriptArchiveService {
	return scriptArchiveService{service: h}
}

func (s scriptArchiveService) export(ctx context.Context, id string) (*scriptArchiveDownload, *scriptAdminError) {
	if id == "" {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "script id required"}
	}
	script, err := s.service.service.GetScript(ctx, id)
	if err != nil {
		return nil, &scriptAdminError{Status: http.StatusNotFound, Message: "script not found: " + err.Error()}
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	metadata := map[string]any{
		"name":          script.Name,
		"description":   script.Description,
		"language":      script.Language,
		"runtime":       script.Runtime,
		"priority":      script.Priority,
		"enabled":       script.Enabled,
		"triggerType":   script.TriggerType,
		"matchRules":    script.MatchRules,
		"config":        script.Config,
		"sourceCode":    script.SourceCode,
		"dependencies":  script.Dependencies,
		"exportVersion": "1.0",
		"exportedAt":    time.Now().Format(time.RFC3339),
	}

	if err := writeJSONZipEntry(zipWriter, "metadata.json", metadata); err != nil {
		return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: "failed to create metadata: " + err.Error()}
	}
	if script.SourceCode != "" {
		if err := writeZipEntry(zipWriter, getMainFilename(script.Language), []byte(script.SourceCode)); err != nil {
			return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: "failed to create main file: " + err.Error()}
		}
	}
	for filename, content := range script.Dependencies {
		if err := writeZipEntry(zipWriter, filename, []byte(content)); err != nil {
			return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: fmt.Sprintf("failed to create %s: %v", filename, err)}
		}
	}
	if len(script.Code) > 0 {
		if err := writeZipEntry(zipWriter, "output.wasm", script.Code); err != nil {
			return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: "failed to create wasm file: " + err.Error()}
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: "failed to finalize zip: " + err.Error()}
	}

	safeFilename := strings.ReplaceAll(script.Name, " ", "-")
	safeFilename = strings.ReplaceAll(safeFilename, "/", "-")
	return &scriptArchiveDownload{Filename: safeFilename + ".zip", Data: buf.Bytes()}, nil
}

func (s scriptArchiveService) importZip(ctx context.Context, r *http.Request) (*scriptDTO, *scriptAdminError) {
	file, _, apiErr := parseMultipartZip(r, "", "no file provided")
	if apiErr != nil {
		return nil, apiErr
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "failed to read file: " + err.Error()}
	}
	if len(fileBytes) < 4 || fileBytes[0] != 'P' || fileBytes[1] != 'K' {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "invalid ZIP file: bad signature"}
	}
	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "invalid ZIP file: " + err.Error()}
	}

	metadata, sourceFiles, wasmCode, apiErr := parseScriptArchive(zipReader)
	if apiErr != nil {
		return nil, apiErr
	}
	newScript, apiErr := buildImportedScript(metadata, sourceFiles, wasmCode)
	if apiErr != nil {
		return nil, apiErr
	}
	if err := s.service.service.CreateScript(ctx, &newScript); err != nil {
		return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: "failed to create script: " + err.Error()}
	}
	dto := toScriptDTO(&newScript)
	return &dto, nil
}

func parseScriptArchive(zipReader *zip.Reader) (map[string]any, map[string]string, string, *scriptAdminError) {
	var metadata map[string]any
	sourceFiles := make(map[string]string)
	var wasmCode string
	fileCount := 0
	seenFiles := make(map[string]bool)

	for _, zipFile := range zipReader.File {
		fileCount++
		if fileCount > maxFilesInZip {
			return nil, nil, "", &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("too many files in ZIP (max %d)", maxFilesInZip)}
		}
		if zipFile.FileInfo().Mode()&os.ModeSymlink != 0 {
			continue
		}
		sanitizedName, apiErr := sanitizeZipEntry(zipFile.Name)
		if apiErr != nil {
			return nil, nil, "", apiErr
		}
		if seenFiles[sanitizedName] {
			return nil, nil, "", &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("duplicate file in ZIP: %s", sanitizedName)}
		}
		seenFiles[sanitizedName] = true
		if apiErr := validateZipFileLimits(zipFile, sanitizedName); apiErr != nil {
			return nil, nil, "", apiErr
		}
		content, apiErr := readLimitedZipContent(zipFile, sanitizedName)
		if apiErr != nil {
			return nil, nil, "", apiErr
		}

		switch sanitizedName {
		case "metadata.json":
			if err := json.Unmarshal(content, &metadata); err != nil {
				return nil, nil, "", &scriptAdminError{Status: http.StatusBadRequest, Message: "invalid metadata: " + err.Error()}
			}
		case "output.wasm":
			wasmCode = base64.StdEncoding.EncodeToString(content)
		default:
			sourceFiles[sanitizedName] = string(content)
		}
	}

	if metadata == nil {
		return nil, nil, "", &scriptAdminError{Status: http.StatusBadRequest, Message: "metadata.json not found in ZIP"}
	}
	return metadata, sourceFiles, wasmCode, nil
}

func buildImportedScript(metadata map[string]any, sourceFiles map[string]string, wasmCode string) (domain.Script, *scriptAdminError) {
	name, _ := metadata["name"].(string)
	description, _ := metadata["description"].(string)
	language, _ := metadata["language"].(string)
	runtime, _ := metadata["runtime"].(string)
	sourceCode, _ := metadata["sourceCode"].(string)

	dependencies := make(map[string]string)
	if deps, ok := metadata["dependencies"].(map[string]any); ok {
		for k, v := range deps {
			if strVal, ok := v.(string); ok {
				dependencies[k] = strVal
			}
		}
	} else {
		mainFilename := getMainFilename(language)
		for filename, content := range sourceFiles {
			if filename != mainFilename {
				dependencies[filename] = content
			} else if sourceCode == "" {
				sourceCode = content
			}
		}
	}

	if name == "" {
		return domain.Script{}, &scriptAdminError{Status: http.StatusBadRequest, Message: "name is required in metadata"}
	}
	if language == "" {
		return domain.Script{}, &scriptAdminError{Status: http.StatusBadRequest, Message: "language is required in metadata"}
	}

	var wasmCodeBytes []byte
	if wasmCode != "" {
		decoded, err := base64.StdEncoding.DecodeString(wasmCode)
		if err != nil {
			log.Printf("Warning: failed to decode WASM code: %v", err)
		} else {
			wasmCodeBytes = decoded
		}
	}

	scriptRuntime := domain.RuntimeExtism
	if runtime == "dart" {
		scriptRuntime = domain.RuntimeDart
	}

	newScript := domain.Script{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		Language:     language,
		Runtime:      scriptRuntime,
		SourceCode:   sourceCode,
		Dependencies: dependencies,
		Code:         wasmCodeBytes,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Enabled:      true,
	}
	if priority, ok := metadata["priority"].(float64); ok {
		newScript.Priority = int(priority)
	}
	if enabled, ok := metadata["enabled"].(bool); ok {
		newScript.Enabled = enabled
	}
	if triggerType, ok := metadata["triggerType"].(string); ok {
		newScript.TriggerType = domain.TriggerType(triggerType)
	}
	if matchRulesMap, ok := metadata["matchRules"].(map[string]any); ok {
		matchRules := domain.MatchRules{}
		if methods, ok := matchRulesMap["methods"].([]any); ok {
			for _, m := range methods {
				if method, ok := m.(string); ok {
					matchRules.Methods = append(matchRules.Methods, method)
				}
			}
		}
		if hostPattern, ok := matchRulesMap["hostPattern"].(string); ok {
			matchRules.HostPattern = hostPattern
		}
		if pathPattern, ok := matchRulesMap["pathPattern"].(string); ok {
			matchRules.PathPattern = pathPattern
		}
		if patternType, ok := matchRulesMap["patternType"].(string); ok {
			matchRules.PatternType = domain.PatternType(patternType)
		}
		newScript.MatchRules = matchRules
	}
	if configMap, ok := metadata["config"].(map[string]any); ok {
		if timeoutMs, ok := configMap["timeoutMs"].(float64); ok {
			newScript.Config.TimeoutMs = int(timeoutMs)
		}
		if memoryLimitMB, ok := configMap["memoryLimitMB"].(float64); ok {
			newScript.Config.MemoryLimitMB = int(memoryLimitMB)
		}
		if allowedHosts, ok := configMap["allowedHosts"].([]any); ok {
			hosts := make([]string, 0, len(allowedHosts))
			for _, h := range allowedHosts {
				if host, ok := h.(string); ok {
					hosts = append(hosts, host)
				}
			}
			newScript.Config.AllowedHosts = hosts
		}
	}

	if err := validateScriptData(&newScript); err != nil {
		return domain.Script{}, &scriptAdminError{Status: http.StatusBadRequest, Message: "validation failed: " + err.Error()}
	}
	return newScript, nil
}

func validateScriptData(script *domain.Script) error {
	if script.Name == "" {
		return errors.New("name is required")
	}
	if script.Language == "" {
		return errors.New("language is required")
	}

	validLanguages := []string{"rust", "go", "javascript", "typescript", "dart", "python", "zig", "kotlin", "swift", "c", "cpp"}
	validLang := false
	for _, lang := range validLanguages {
		if strings.EqualFold(script.Language, lang) {
			validLang = true
			break
		}
	}
	if !validLang {
		return fmt.Errorf("invalid language: %s", script.Language)
	}
	if len(script.SourceCode) > maxSourceCodeSize {
		return fmt.Errorf("source code too large (max %dKB)", maxSourceCodeSize/1024)
	}
	totalSize := len(script.SourceCode)
	for _, content := range script.Dependencies {
		totalSize += len(content)
		if len(content) > maxDependencyFileSize {
			return fmt.Errorf("dependency file too large (max %dKB per file)", maxDependencyFileSize/1024)
		}
	}
	if totalSize > maxImportProjectSize {
		return fmt.Errorf("total project size too large (max %dMB)", maxImportProjectSize/1024/1024)
	}
	return nil
}

func validateMatchRulesRegex(rules domain.MatchRules) error {
	if rules.PatternType != domain.PatternRegex {
		return nil
	}
	if rules.HostPattern != "" {
		if _, err := regexp.Compile(rules.HostPattern); err != nil {
			return fmt.Errorf("invalid regex pattern for hostPattern: %w", err)
		}
	}
	if rules.PathPattern != "" {
		if _, err := regexp.Compile(rules.PathPattern); err != nil {
			return fmt.Errorf("invalid regex pattern for pathPattern: %w", err)
		}
	}
	return nil
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func readLimitedZipContent(zipFile *zip.File, sanitizedName string) ([]byte, *scriptAdminError) {
	content, err := readZipFile(zipFile, int64(maxSingleFileSize))
	if err != nil {
		return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: fmt.Sprintf("failed to read %s: %v", sanitizedName, err)}
	}
	if len(content) > int(maxSingleFileSize) {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("file %s exceeds size limit after decompression", sanitizedName)}
	}
	return content, nil
}

func validateZipFileLimits(zipFile *zip.File, sanitizedName string) *scriptAdminError {
	if zipFile.UncompressedSize64 > maxSingleFileSize {
		return &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("file %s too large (max %dKB)", sanitizedName, maxSingleFileSize/1024)}
	}
	if zipFile.CompressedSize64 > 0 {
		ratio := zipFile.UncompressedSize64 / zipFile.CompressedSize64
		if ratio > maxCompressionRatio {
			return &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("file %s has suspicious compression ratio (max %d:1)", sanitizedName, maxCompressionRatio)}
		}
	} else if zipFile.UncompressedSize64 > maxSingleFileSize {
		return &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("stored file %s too large (max %dKB)", sanitizedName, maxSingleFileSize/1024)}
	}
	return nil
}

func writeZipEntry(zipWriter *zip.Writer, name string, payload []byte) error {
	fw, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = fw.Write(payload)
	return err
}

func writeJSONZipEntry(zipWriter *zip.Writer, name string, payload any) error {
	fw, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	return json.NewEncoder(fw).Encode(payload)
}

func readZipFile(zipFile *zip.File, maxSize int64) ([]byte, error) {
	rc, err := zipFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open: %w", err)
	}
	defer rc.Close()
	content, err := io.ReadAll(io.LimitReader(rc, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}
	return content, nil
}

func sanitizeZipEntry(name string) (string, *scriptAdminError) {
	sanitizedName, err := sanitizeZipFilename(name)
	if err != nil {
		return "", &scriptAdminError{Status: http.StatusBadRequest, Message: err.Error()}
	}
	return sanitizedName, nil
}

func sanitizeZipFilename(name string) (string, error) {
	if strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("null byte in filename not allowed")
	}
	normalized := strings.ReplaceAll(name, "\\", "/")
	if strings.Contains(normalized, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", name)
	}
	cleaned := filepath.Clean(normalized)
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute path not allowed: %s", name)
	}
	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == ".." {
			return "", fmt.Errorf("path traversal not allowed: %s", name)
		}
	}
	return cleaned, nil
}
