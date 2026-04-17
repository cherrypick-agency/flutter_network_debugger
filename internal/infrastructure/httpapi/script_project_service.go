package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

type scriptProjectFile struct {
	Filename string `json:"filename"`
	Size     int    `json:"size"`
	Type     string `json:"type"`
}

type scriptProjectFilesResponse struct {
	ScriptID  string              `json:"scriptId"`
	FileCount int                 `json:"fileCount"`
	Files     []scriptProjectFile `json:"files"`
}

type scriptProjectUploadResponse struct {
	Success   bool     `json:"success"`
	FileCount int      `json:"fileCount"`
	Files     []string `json:"files"`
	TotalSize int      `json:"totalSize"`
}

type scriptProjectDownload struct {
	Filename string
	Data     []byte
}

type scriptProjectService struct {
	service *ScriptHandlers
}

func newScriptProjectService(h *ScriptHandlers) scriptProjectService {
	return scriptProjectService{service: h}
}

func (s scriptProjectService) upload(ctx context.Context, id string, r *http.Request) (scriptProjectUploadResponse, *scriptAdminError) {
	if id == "" {
		return scriptProjectUploadResponse{}, &scriptAdminError{Status: http.StatusBadRequest, Message: "script id required"}
	}
	script, err := s.service.service.GetScript(ctx, id)
	if err != nil {
		return scriptProjectUploadResponse{}, &scriptAdminError{Status: http.StatusNotFound, Message: "script not found: " + err.Error()}
	}

	dependencies, totalSize, apiErr := parseProjectUploadArchive(r)
	if apiErr != nil {
		return scriptProjectUploadResponse{}, apiErr
	}

	script.Dependencies = dependencies
	script.UpdatedAt = time.Now()
	if err := s.service.service.UpdateScript(ctx, script); err != nil {
		return scriptProjectUploadResponse{}, &scriptAdminError{Status: http.StatusInternalServerError, Message: "failed to update script: " + err.Error()}
	}

	return scriptProjectUploadResponse{
		Success:   true,
		FileCount: len(dependencies),
		Files:     getFileList(dependencies),
		TotalSize: totalSize,
	}, nil
}

func (s scriptProjectService) download(ctx context.Context, id string) (*scriptProjectDownload, *scriptAdminError) {
	if id == "" {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "script id required"}
	}
	script, err := s.service.service.GetScript(ctx, id)
	if err != nil {
		return nil, &scriptAdminError{Status: http.StatusNotFound, Message: "script not found: " + err.Error()}
	}
	data, apiErr := buildProjectZip(script)
	if apiErr != nil {
		return nil, apiErr
	}
	return &scriptProjectDownload{
		Filename: script.ID + "-project.zip",
		Data:     data,
	}, nil
}

func (s scriptProjectService) listFiles(ctx context.Context, id string) (*scriptProjectFilesResponse, *scriptAdminError) {
	if id == "" {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "script id required"}
	}
	script, err := s.service.service.GetScript(ctx, id)
	if err != nil {
		return nil, &scriptAdminError{Status: http.StatusNotFound, Message: "script not found: " + err.Error()}
	}
	files := make([]scriptProjectFile, 0, len(script.Dependencies)+1)
	if script.SourceCode != "" {
		files = append(files, scriptProjectFile{
			Filename: getMainFilename(script.Language),
			Size:     len(script.SourceCode),
			Type:     "source",
		})
	}
	depNames := getFileList(script.Dependencies)
	for _, filename := range depNames {
		files = append(files, scriptProjectFile{
			Filename: filename,
			Size:     len(script.Dependencies[filename]),
			Type:     getDependencyType(filename),
		})
	}
	return &scriptProjectFilesResponse{
		ScriptID:  script.ID,
		FileCount: len(files),
		Files:     files,
	}, nil
}

func parseProjectUploadArchive(r *http.Request) (map[string]string, int, *scriptAdminError) {
	file, header, apiErr := parseMultipartZip(r, "file required: ", "file required")
	if apiErr != nil {
		return nil, 0, apiErr
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		return nil, 0, &scriptAdminError{Status: http.StatusBadRequest, Message: "file must be a ZIP archive"}
	}

	zipReader, apiErr := openZipUpload(file)
	if apiErr != nil {
		return nil, 0, apiErr
	}

	dependencies := make(map[string]string)
	totalSize := 0
	fileCount := 0
	seenFiles := make(map[string]bool)

	for _, zipFile := range zipReader.File {
		fileCount++
		if fileCount > maxFilesInZip {
			return nil, 0, &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("too many files in ZIP (max %d)", maxFilesInZip)}
		}
		if zipFile.FileInfo().Mode()&os.ModeSymlink != 0 {
			continue
		}
		sanitizedName, apiErr := sanitizeZipEntry(zipFile.Name)
		if apiErr != nil {
			return nil, 0, apiErr
		}
		if seenFiles[sanitizedName] {
			return nil, 0, &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("duplicate file in ZIP: %s", sanitizedName)}
		}
		seenFiles[sanitizedName] = true
		if zipFile.FileInfo().IsDir() {
			continue
		}
		if strings.HasPrefix(filepath.Base(sanitizedName), ".") || strings.Contains(sanitizedName, "__MACOSX") {
			continue
		}
		if !isAllowedProjectExtension(sanitizedName) {
			continue
		}
		if apiErr := validateZipFileLimits(zipFile, sanitizedName); apiErr != nil {
			return nil, 0, apiErr
		}
		content, apiErr := readLimitedZipContent(zipFile, sanitizedName)
		if apiErr != nil {
			return nil, 0, apiErr
		}
		totalSize += len(content)
		if totalSize > maxTotalProjectSize {
			return nil, 0, &scriptAdminError{Status: http.StatusBadRequest, Message: fmt.Sprintf("total project size exceeds %dMB", maxTotalProjectSize/1024/1024)}
		}

		filename := strippedProjectFilename(sanitizedName)
		dependencies[filename] = string(content)
	}

	if len(dependencies) == 0 {
		return nil, 0, &scriptAdminError{Status: http.StatusBadRequest, Message: "no valid source files found in ZIP"}
	}

	return dependencies, totalSize, nil
}

func buildProjectZip(script *domain.Script) ([]byte, *scriptAdminError) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	if script.SourceCode != "" {
		if err := writeZipEntry(zipWriter, getMainFilename(script.Language), []byte(script.SourceCode)); err != nil {
			return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: "failed to create zip: " + err.Error()}
		}
	}
	for filename, content := range script.Dependencies {
		if err := writeZipEntry(zipWriter, filename, []byte(content)); err != nil {
			return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: fmt.Sprintf("failed to create %s: %v", filename, err)}
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, &scriptAdminError{Status: http.StatusInternalServerError, Message: "failed to finalize zip: " + err.Error()}
	}
	return buf.Bytes(), nil
}

func parseMultipartZip(r *http.Request, formErrorPrefix, noFileMessage string) (multipart.File, *multipart.FileHeader, *scriptAdminError) {
	if err := r.ParseMultipartForm(maxUploadFormSize); err != nil {
		return nil, nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "failed to parse form: " + err.Error()}
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		msg := noFileMessage
		if formErrorPrefix != "" {
			msg = formErrorPrefix + err.Error()
		}
		return nil, nil, &scriptAdminError{Status: http.StatusBadRequest, Message: msg}
	}
	return file, header, nil
}

func openZipUpload(file multipart.File) (*zip.Reader, *scriptAdminError) {
	zipData, err := io.ReadAll(file)
	if err != nil {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "failed to read zip: " + err.Error()}
	}
	if len(zipData) < 4 || zipData[0] != 'P' || zipData[1] != 'K' {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "invalid ZIP file: bad signature"}
	}
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, &scriptAdminError{Status: http.StatusBadRequest, Message: "invalid zip file: " + err.Error()}
	}
	return zipReader, nil
}

func isAllowedProjectExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{
		".rs": true, ".go": true, ".ts": true, ".js": true,
		".toml": true, ".json": true, ".mod": true, ".sum": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".py": true, ".zig": true, ".kt": true, ".swift": true,
	}
	return allowedExts[ext]
}

func strippedProjectFilename(name string) string {
	filename := name
	if strings.Contains(filename, "/") {
		parts := strings.Split(filename, "/")
		if len(parts) > 1 {
			filename = strings.Join(parts[1:], "/")
		}
	}
	return filename
}

func getMainFilename(language string) string {
	switch strings.ToLower(language) {
	case "rust":
		return "src/lib.rs"
	case "go":
		return "main.go"
	case "javascript", "typescript":
		return "index.ts"
	case "dart":
		return "main.dart"
	case "python":
		return "main.py"
	case "zig":
		return "main.zig"
	case "kotlin":
		return "Main.kt"
	case "swift":
		return "main.swift"
	case "c":
		return "main.c"
	case "cpp", "c++":
		return "main.cpp"
	default:
		return "main.txt"
	}
}

func getDependencyType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".toml", ".json", ".mod", ".sum":
		return "config"
	case ".rs", ".go", ".ts", ".js", ".dart", ".py", ".zig", ".kt", ".swift", ".c", ".cpp", ".h", ".hpp":
		return "source"
	default:
		return "other"
	}
}

func getFileList(deps map[string]string) []string {
	files := make([]string, 0, len(deps))
	for filename := range deps {
		files = append(files, filename)
	}
	sort.Strings(files)
	return files
}
