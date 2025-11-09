package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"network-debugger/internal/features/process/domain"
)

// Service - сервис детекции процессов и управления конфигурацией
type Service struct {
	config           domain.IConfigRepository
	iconCache        domain.IIconCacheRepository
	localDetector    domain.IProcessDetector // без привилегий
	helperClient     domain.IHelperClient    // с привилегиями
	iconExtractor    domain.IIconExtractor
	helperInstaller  HelperInstaller // installer для helper tool
	helperBinaryPath string          // путь к helper binary
	logger           *zerolog.Logger
}

// HelperInstaller - интерфейс для установки/удаления helper tool
type HelperInstaller interface {
	IsInstalled() bool
	Install(helperBinaryPath string) error
	Uninstall() error
	GetVersion() string
}

// NewService - создать новый сервис детекции процессов
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

// DetectForConnection - главный метод детекции процесса для сетевого соединения
// Стратегия: пробуем helper → fallback на local → fallback на "Unknown"
func (s *Service) DetectForConnection(ctx context.Context, localPort uint32) (*domain.ProcessInfo, error) {
	cfg, err := s.config.Load(ctx)
	if err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		return nil, nil // детекция отключена
	}

	var info *domain.ProcessInfo

	// 1. Попытка через helper (если включен и доступен)
	if cfg.UseHelperTool && s.helperClient != nil && s.helperClient.IsRunning() {
		info, err = s.helperClient.DetectProcess(localPort)
		if err != nil {
			s.logger.Warn().Err(err).Uint32("port", localPort).Msg("Helper detection failed, falling back to local")
		} else {
			s.logger.Debug().Uint32("port", localPort).Int32("pid", info.PID).Msg("Process detected via helper")
		}
	}

	// 2. Fallback на локальную детекцию
	if info == nil {
		info, err = s.localDetector.DetectByPort(ctx, localPort)
		if err != nil {
			if cfg.FallbackEnabled {
				// Вернуть "Unknown Process"
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

	// 3. Получить иконку
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

// getIcon - получение иконки с кешированием
func (s *Service) getIcon(ctx context.Context, pid int32, path string) (*domain.AppIcon, error) {
	cfg, _ := s.config.Load(ctx)

	if !cfg.CacheEnabled {
		return s.extractIcon(ctx, pid, path)
	}

	// Проверить кеш
	key := fmt.Sprintf("icon:%d:%s", pid, path)
	if cached, err := s.iconCache.Get(key); err == nil {
		s.logger.Debug().Str("key", key).Msg("Icon found in cache")
		return cached, nil
	}

	// Извлечь иконку
	icon, err := s.extractIcon(ctx, pid, path)
	if err != nil {
		return nil, err
	}

	// Сохранить в кеш
	ttl := time.Duration(cfg.CacheTTLSeconds) * time.Second
	if err := s.iconCache.Set(key, icon, ttl); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to cache icon")
	} else {
		s.logger.Debug().Str("key", key).Dur("ttl", ttl).Msg("Icon cached")
	}

	return icon, nil
}

// extractIcon - извлечение иконки (helper или local)
func (s *Service) extractIcon(ctx context.Context, pid int32, path string) (*domain.AppIcon, error) {
	cfg, _ := s.config.Load(ctx)

	// Попытка через helper
	if cfg.UseHelperTool && s.helperClient != nil && s.helperClient.IsRunning() {
		icon, err := s.helperClient.ExtractIcon(pid)
		if err == nil {
			s.logger.Debug().Int32("pid", pid).Msg("Icon extracted via helper")
			return icon, nil
		}
		s.logger.Debug().Err(err).Msg("Helper icon extraction failed, trying local")
	}

	// Fallback на локальное извлечение
	return s.iconExtractor.ExtractByPID(ctx, pid)
}

// GetConfig - получить текущую конфигурацию детекции
func (s *Service) GetConfig(ctx context.Context) (*domain.DetectionConfig, error) {
	return s.config.Load(ctx)
}

// SaveConfig - сохранить конфигурацию детекции
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

// CheckHelperStatus - проверить статус helper tool
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

// InstallHelper - установить helper tool
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

// HelperStatus - статус helper tool
type HelperStatus struct {
	Installed bool
	Running   bool
	Version   string
}
