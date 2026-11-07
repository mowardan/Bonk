package modes

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bonk/assets"
)

// PlayMode defines how audio files are picked upon impact.
type PlayMode int

const (
	ModeRandom PlayMode = iota
	ModeEscalation
)

// SoundPack represents a collection of sound files for reaction playback.
type SoundPack struct {
	Name     string
	Mode     PlayMode
	IsCustom bool
	Files    []string
	EmbedFS  embed.FS
	BaseDir  string
}

// LoadPack initializes a SoundPack based on requested configuration.
func LoadPack(name string, customPath string) (*SoundPack, error) {
	switch name {
	case "sexy":
		sp := &SoundPack{
			Name:    "sexy",
			Mode:    ModeEscalation,
			EmbedFS: assets.SexyFS,
			BaseDir: "sounds/sexy",
		}
		if err := sp.loadEmbeddedFiles(); err != nil {
			return nil, err
		}
		return sp, nil

	case "halo":
		sp := &SoundPack{
			Name:    "halo",
			Mode:    ModeRandom,
			EmbedFS: assets.HaloFS,
			BaseDir: "sounds/halo",
		}
		if err := sp.loadEmbeddedFiles(); err != nil {
			return nil, err
		}
		return sp, nil

	case "lizard":
		sp := &SoundPack{
			Name:    "lizard",
			Mode:    ModeEscalation,
			EmbedFS: assets.LizardFS,
			BaseDir: "sounds/lizard",
		}
		if err := sp.loadEmbeddedFiles(); err != nil {
			return nil, err
		}
		return sp, nil

	case "custom":
		if customPath == "" {
			return nil, fmt.Errorf("custom audio mode requires a directory path via --custom")
		}
		return loadCustomPack(customPath)

	default: // "normal" or "pain"
		sp := &SoundPack{
			Name:    "normal",
			Mode:    ModeRandom,
			EmbedFS: assets.PainFS,
			BaseDir: "sounds/pain",
		}
		if err := sp.loadEmbeddedFiles(); err != nil {
			return nil, err
		}
		return sp, nil
	}
}

func (sp *SoundPack) loadEmbeddedFiles() error {
	entries, err := sp.EmbedFS.ReadDir(sp.BaseDir)
	if err != nil {
		return fmt.Errorf("reading embedded pack %s: %w", sp.Name, err)
	}

	sp.Files = make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".mp3") {
			sp.Files = append(sp.Files, filepath.Join(sp.BaseDir, entry.Name()))
		}
	}
	sort.Strings(sp.Files)
	if len(sp.Files) == 0 {
		return fmt.Errorf("no MP3 files found in embedded pack %s", sp.Name)
	}
	return nil
}

func loadCustomPack(dirPath string) (*SoundPack, error) {
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("custom directory not found: %s", dirPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("custom path is not a directory: %s", dirPath)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("reading custom directory %s: %w", dirPath, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".mp3") {
			files = append(files, filepath.Join(dirPath, e.Name()))
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no .mp3 files found in custom directory: %s", dirPath)
	}

	return &SoundPack{
		Name:     "custom",
		Mode:     ModeRandom,
		IsCustom: true,
		Files:    files,
		BaseDir:  dirPath,
	}, nil
}

// GetAudioData returns the raw bytes for the file at filePath.
func (sp *SoundPack) GetAudioData(filePath string) ([]byte, error) {
	if sp.IsCustom {
		return os.ReadFile(filePath)
	}
	return sp.EmbedFS.ReadFile(filePath)
}
