package main

import (
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const codeManifestSchema = "opennox-game-exe-code-v1"

type codeRecord struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Size    uint32 `json:"size"`
	SHA256  string `json:"sha256"`
	Purpose string `json:"purpose"`
}

type codeManifest struct {
	Schema     string       `json:"schema"`
	OracleID   string       `json:"oracle_id"`
	Path       string       `json:"path"`
	FileSHA256 string       `json:"file_sha256"`
	Format     string       `json:"format"`
	ImageBase  string       `json:"image_base"`
	Functions  []codeRecord `json:"functions"`
}

func verifyCodeOracle(root, manifestPath, oracleManifestPath string) (*codeManifest, error) {
	m, err := readCodeManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	oracle, err := readManifest(oracleManifestPath)
	if err != nil {
		return nil, fmt.Errorf("read full-tree oracle manifest: %w", err)
	}
	if oracle.ID != m.OracleID {
		return nil, fmt.Errorf("code oracle id %q does not match full-tree oracle id %q", m.OracleID, oracle.ID)
	}
	fileBound := false
	for _, rec := range oracle.Files {
		if rec.Path == m.Path {
			fileBound = rec.SHA256 == m.FileSHA256
			break
		}
	}
	if !fileBound {
		return nil, fmt.Errorf("%s code oracle is not bound to the same file in full-tree oracle %q", m.Path, oracle.ID)
	}
	name := filepath.Join(root, filepath.FromSlash(m.Path))
	info, err := os.Lstat(name)
	if err != nil {
		return nil, fmt.Errorf("inspect %s: %w", m.Path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s must be a real regular file", m.Path)
	}
	digest, _, err := hashFile(name)
	if err != nil {
		return nil, fmt.Errorf("hash %s: %w", m.Path, err)
	}
	if digest != m.FileSHA256 {
		return nil, fmt.Errorf("%s SHA-256 mismatch: got %s, want %s", m.Path, digest, m.FileSHA256)
	}

	f, err := pe.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open %s as PE: %w", m.Path, err)
	}
	defer f.Close()
	if f.Machine != pe.IMAGE_FILE_MACHINE_I386 {
		return nil, fmt.Errorf("%s machine is %#x, want I386 (%#x)", m.Path, f.Machine, pe.IMAGE_FILE_MACHINE_I386)
	}
	oh, ok := f.OptionalHeader.(*pe.OptionalHeader32)
	if !ok || oh.Magic != 0x10b {
		return nil, fmt.Errorf("%s is not PE32", m.Path)
	}
	base, _ := parseCodeAddress(m.ImageBase)
	if uint64(oh.ImageBase) != base {
		return nil, fmt.Errorf("%s image base is %#x, want %s", m.Path, oh.ImageBase, m.ImageBase)
	}

	for _, rec := range m.Functions {
		va, _ := parseCodeAddress(rec.Address)
		data, section, err := readCodeRange(f, va, rec.Size)
		if err != nil {
			return nil, fmt.Errorf("%s at %s: %w", rec.Name, rec.Address, err)
		}
		got := sha256.Sum256(data)
		if hex.EncodeToString(got[:]) != rec.SHA256 {
			return nil, fmt.Errorf("%s at %s in %s has SHA-256 %x, want %s", rec.Name, rec.Address, section, got, rec.SHA256)
		}
	}
	return m, nil
}

func readCodeManifest(name string) (*codeManifest, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var m codeManifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("decode code manifest: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("code manifest has trailing JSON data")
		}
		return nil, fmt.Errorf("decode code manifest trailer: %w", err)
	}
	if err := validateCodeManifest(&m); err != nil {
		return nil, fmt.Errorf("invalid code manifest: %w", err)
	}
	return &m, nil
}

func validateCodeManifest(m *codeManifest) error {
	if m.Schema != codeManifestSchema {
		return fmt.Errorf("unsupported schema %q", m.Schema)
	}
	if strings.TrimSpace(m.OracleID) == "" {
		return errors.New("oracle_id must be non-empty")
	}
	if err := validatePath(m.Path); err != nil {
		return err
	}
	if err := validateSHA256(m.FileSHA256, m.Path); err != nil {
		return err
	}
	if m.Format != "PE32/I386" {
		return fmt.Errorf("unsupported executable format %q", m.Format)
	}
	base, err := parseCodeAddress(m.ImageBase)
	if err != nil {
		return fmt.Errorf("invalid image_base: %w", err)
	}
	if base != 0x00400000 {
		return fmt.Errorf("unexpected image_base %s", m.ImageBase)
	}
	if len(m.Functions) == 0 {
		return errors.New("code manifest contains no functions")
	}
	names := make(map[string]bool)
	var previousEnd uint64
	for i, rec := range m.Functions {
		if strings.TrimSpace(rec.Name) == "" || strings.TrimSpace(rec.Purpose) == "" {
			return fmt.Errorf("function %d has an empty name or purpose", i)
		}
		if names[rec.Name] {
			return fmt.Errorf("duplicate function name %q", rec.Name)
		}
		names[rec.Name] = true
		va, err := parseCodeAddress(rec.Address)
		if err != nil {
			return fmt.Errorf("invalid address for %q: %w", rec.Name, err)
		}
		if va < base || rec.Size == 0 || rec.Size > 1<<20 {
			return fmt.Errorf("invalid range for %q", rec.Name)
		}
		end := va + uint64(rec.Size)
		if end < va {
			return fmt.Errorf("range overflow for %q", rec.Name)
		}
		if i > 0 && va < previousEnd {
			return fmt.Errorf("function ranges are not sorted and disjoint at %q", rec.Name)
		}
		previousEnd = end
		if err := validateSHA256(rec.SHA256, rec.Name); err != nil {
			return err
		}
	}
	return nil
}

func parseCodeAddress(s string) (uint64, error) {
	if len(s) != 10 || !strings.HasPrefix(s, "0x") {
		return 0, fmt.Errorf("address %q must use canonical 0xXXXXXXXX form", s)
	}
	v, err := strconv.ParseUint(s[2:], 16, 32)
	if err != nil || fmt.Sprintf("0x%08X", v) != s {
		return 0, fmt.Errorf("address %q must use canonical 0xXXXXXXXX form", s)
	}
	return v, nil
}

func validateSHA256(value, subject string) error {
	digest, err := hex.DecodeString(value)
	if err != nil || len(digest) != sha256.Size || strings.ToLower(value) != value {
		return fmt.Errorf("invalid SHA-256 for %q", subject)
	}
	return nil
}

func readCodeRange(f *pe.File, va uint64, size uint32) ([]byte, string, error) {
	oh, ok := f.OptionalHeader.(*pe.OptionalHeader32)
	if !ok {
		return nil, "", errors.New("not a PE32 optional header")
	}
	base := uint64(oh.ImageBase)
	if va < base {
		return nil, "", errors.New("address precedes image base")
	}
	rva := va - base
	end := rva + uint64(size)
	if end < rva {
		return nil, "", errors.New("range overflow")
	}
	for _, section := range f.Sections {
		start := uint64(section.VirtualAddress)
		rawEnd := start + uint64(section.Size)
		if rva < start || end > rawEnd {
			continue
		}
		if section.Characteristics&pe.IMAGE_SCN_MEM_EXECUTE == 0 {
			return nil, section.Name, errors.New("range is not in an executable section")
		}
		data := make([]byte, int(size))
		n, err := section.ReadAt(data, int64(rva-start))
		if err != nil || n != len(data) {
			return nil, section.Name, fmt.Errorf("read range: got %d of %d bytes: %w", n, len(data), err)
		}
		return data, section.Name, nil
	}
	return nil, "", errors.New("range is outside file-backed PE sections")
}
