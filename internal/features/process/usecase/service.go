package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"network-debugger/internal/features/process/domain"
)

// Service - service for process detection and configuration management
type Service struct {
	config           domain.IConfigRepository
	iconCache        domain.IIconCacheRepository
	localDetector    domain.IProcessDetector // without privileges
	helperClient     domain.IHelperClient    // with privileges
	iconExtractor    domain.IIconExtractor
	helperInstaller  HelperInstaller // installer for helper tool
	helperBinaryPath string          // path to helper binary
	logger           *zerolog.Logger
}

// HelperInstaller - interface for installing/uninstalling helper tool
type HelperInstaller interface {
	IsInstalled() bool
	Install(helperBinaryPath string) error
	Uninstall() error
	GetVersion() string
}

// NewService - create new process detection service
func NewService(
	config domain.IConfigRepository,
	iconCache domain.IIconCacheRepository,
	localDetector domain.IProcessDetector,
	helperClient domain.IHelperClient,
	iconExtractor domain.IIconExtractor,
	helperInstaller HelperInstaller,
	helperBinaryPath string,
	logger *zerolog.Logger,
) *Service {
	return &Service{
		config:           config,
		iconCache:        iconCache,
		localDetector:    localDetector,
		helperClient:     helperClient,
		iconExtractor:    iconExtractor,
		helperInstaller:  helperInstaller,
		helperBinaryPath: helperBinaryPath,
		logger:           logger,
	}
}

// DetectForConnection - main method for detecting process for network connection
// Strategy: try helper -> fallback to local -> fallback to "Unknown"
func (s *Service) DetectForConnection(ctx context.Context, localPort uint32) (*domain.ProcessInfo, error) {
	cfg, err := s.config.Load(ctx)
	if err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		return nil, nil // detection disabled
	}

	var info *domain.ProcessInfo

	// 1. Try via helper (if enabled and available)
	if cfg.UseHelperTool && s.helperClient != nil && s.helperClient.IsRunning() {
		info, err = s.helperClient.DetectProcess(localPort)
		if err != nil {
			s.logger.Warn().Err(err).Uint32("port", localPort).Msg("Helper detection failed, falling back to local")
		} else {
			s.logger.Debug().Uint32("port", localPort).Int32("pid", info.PID).Msg("Process detected via helper")
		}
	}

	// 2. Fallback to local detection
	if info == nil {
		info, err = s.localDetector.DetectByPort(ctx, localPort)
		if err != nil {
			if cfg.FallbackEnabled {
				// Return "Unknown Process"
				s.logger.Debug().Err(err).Uint32("port", localPort).Msg("Local detection failed, using fallback")
				return &domain.ProcessInfo{
					Name:       "Unknown Process",
					DetectedAt: time.Now(),
				}, nil
			}
			return nil, err
		}
		s.logger.Debug().Uint32("port", localPort).Int32("pid", info.PID).Msg("Process detected locally")
	}

	// 3. Get icon
	if info != nil && info.PID > 0 {
		icon, err := s.getIcon(ctx, info.PID, info.ExecutablePath)
		if err != nil {
			s.logger.Debug().Err(err).Int32("pid", info.PID).Msg("Failed to get icon")
		} else {
			info.Icon = icon
		}
	}

	return info, nil
}

// getIcon - get icon with caching
func (s *Service) getIcon(ctx context.Context, pid int32, path string) (*domain.AppIcon, error) {
	cfg, _ := s.config.Load(ctx)

	if !cfg.CacheEnabled {
		return s.extractIcon(ctx, pid, path)
	}

	// Check cache
	key := fmt.Sprintf("icon:%d:%s", pid, path)
	if cached, err := s.iconCache.Get(key); err == nil {
		s.logger.Debug().Str("key", key).Msg("Icon found in cache")
		return cached, nil
	}

	// Extract icon
	icon, err := s.extractIcon(ctx, pid, path)
	if err != nil {
		return nil, err
	}

	// Save to cache
	ttl := time.Duration(cfg.CacheTTLSeconds) * time.Second
	if err := s.iconCache.Set(key, icon, ttl); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to cache icon")
	} else {
		s.logger.Debug().Str("key", key).Dur("ttl", ttl).Msg("Icon cached")
	}

	return icon, nil
}

// extractIcon - extract icon (helper or local)
func (s *Service) extractIcon(ctx context.Context, pid int32, path string) (*domain.AppIcon, error) {
	cfg, _ := s.config.Load(ctx)

	// Try via helper
	if cfg.UseHelperTool && s.helperClient != nil && s.helperClient.IsRunning() {
		icon, err := s.helperClient.ExtractIcon(pid)
		if err == nil {
			s.logger.Debug().Int32("pid", pid).Msg("Icon extracted via helper")
			return icon, nil
		}
		s.logger.Debug().Err(err).Msg("Helper icon extraction failed, trying local")
	}

	// Fallback to local extraction
	return s.iconExtractor.ExtractByPID(ctx, pid)
}

// GetConfig - get current detection configuration
func (s *Service) GetConfig(ctx context.Context) (*domain.DetectionConfig, error) {
	return s.config.Load(ctx)
}

// SaveConfig - save detection configuration
func (s *Service) SaveConfig(ctx context.Context, cfg *domain.DetectionConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	s.logger.Info().
		Bool("enabled", cfg.Enabled).
		Bool("use_helper", cfg.UseHelperTool).
		Msg("Saving process detection config")

	return s.config.Save(ctx, cfg)
}

// CheckHelperStatus - check helper tool status
func (s *Service) CheckHelperStatus() HelperStatus {
	if s.helperClient == nil {
		return HelperStatus{
			Installed: false,
			Running:   false,
		}
	}

	installed := false
	version := ""
	if s.helperInstaller != nil {
		installed = s.helperInstaller.IsInstalled()
		if installed {
			version = s.helperInstaller.GetVersion()
		}
	}

	running := s.helperClient.IsRunning()
	s.logger.Debug().Bool("installed", installed).Bool("running", running).Msg("Helper status checked")

	return HelperStatus{
		Installed: installed,
		Running:   running,
		Version:   version,
	}
}

// InstallHelper - install helper tool
func (s *Service) InstallHelper(ctx context.Context) error {
	if s.helperInstaller == nil {
		return fmt.Errorf("helper installer not available")
	}

	if s.helperBinaryPath == "" {
		return fmt.Errorf("helper binary path not configured")
	}

	s.logger.Info().Str("binaryPath", s.helperBinaryPath).Msg("Installing helper tool")

	if err := s.helperInstaller.Install(s.helperBinaryPath); err != nil {
		s.logger.Error().Err(err).Msg("Helper installation failed")
		return fmt.Errorf("failed to install helper: %w", err)
	}

	s.logger.Info().Msg("Helper tool installed successfully")
	return nil
}

// HelperStatus - helper tool status
type HelperStatus struct {
	Installed bool
	Running   bool
	Version   string
}
