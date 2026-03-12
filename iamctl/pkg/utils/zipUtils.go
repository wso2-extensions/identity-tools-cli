package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ZipAndDeleteExports(outputDir string, sourceDirs []string) error {
	exported, err := zipExports(outputDir, sourceDirs)
	if err != nil {
		return fmt.Errorf("failed to create zip: %w", err)
	}

	if exported {
		for _, dir := range sourceDirs {
			fullPath := filepath.Join(outputDir, dir)
			if err := os.RemoveAll(fullPath); err != nil {
				fmt.Printf("Warning: Error deleting directory %s: %v\n", fullPath, err)
			} else {
				fmt.Printf("Deleted directory: %s\n", fullPath)
			}
		}
	}
	return nil
}

func zipExports(outputDir string, sourceDirs []string) (bool, error) {
	if allDirsEmpty(outputDir, sourceDirs) {
		fmt.Println("No exports found to zip.")
		return false, nil
	}

	archiveDir := filepath.Join(outputDir, "archives")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return false, fmt.Errorf("could not create archive directory: %w", err)
	}

	zipPath := filepath.Join(archiveDir, fmt.Sprintf("export_%d.zip", time.Now().Unix()))
	fmt.Printf("Creating archive: %s\n", zipPath)

	f, err := os.Create(zipPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	archive := zip.NewWriter(f)

	for _, sourceDir := range sourceDirs {
		fullPath := filepath.Join(outputDir, sourceDir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("Adding contents of '%s'...\n", fullPath)
		if err := addDirToZip(archive, fullPath); err != nil {
			archive.Close()
			return false, err
		}
	}

	if err := archive.Close(); err != nil {
		return false, fmt.Errorf("failed to finalize zip: %w", err)
	}

	fmt.Println("Zip created successfully!")
	return true, nil
}

func addDirToZip(archive *zip.Writer, source string) error {
	baseDir := filepath.Base(source)
	allowedExts := map[string]bool{
		".yaml": true, ".yml": true, ".xml": true, ".json": true,
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !allowedExts[ext] {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(filepath.Join(baseDir, relPath))
		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

func allDirsEmpty(outputDir string, dirs []string) bool {
	for _, dir := range dirs {
		fullPath := filepath.Join(outputDir, dir)
		info, err := os.Stat(fullPath)

		if err == nil && info.IsDir() && !isDirEmpty(fullPath) {
			return false
		}
	}
	return true
}

func isDirEmpty(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return true
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF
}
