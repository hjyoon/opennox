// Command noxoracle seals and verifies a local Nox installation used as a
// private compatibility oracle. It records content only; original game assets
// are never copied into the source tree.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const manifestSchema = "opennox-nox-oracle-v1"

type fileRecord struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type manifest struct {
	Schema     string       `json:"schema"`
	ID         string       `json:"id"`
	RootName   string       `json:"root_name"`
	FileCount  int          `json:"file_count"`
	TotalBytes int64        `json:"total_bytes"`
	TreeSHA256 string       `json:"tree_sha256"`
	Files      []fileRecord `json:"files"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "noxoracle:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return usageError()
	}
	switch args[0] {
	case "seal":
		fs := flag.NewFlagSet("seal", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		root := fs.String("root", "", "Nox data directory")
		out := fs.String("out", "", "new manifest path")
		id := fs.String("id", "", "stable identifier for this reference copy")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *root == "" || *out == "" || *id == "" {
			return errors.New("seal requires -root, -out, and -id")
		}
		m, err := scanTree(*root, *id, filepath.Base(filepath.Clean(*root)))
		if err != nil {
			return err
		}
		if err := writeManifestExclusive(*out, m); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "sealed %s: %d files, %d bytes, tree SHA-256 %s\n", m.ID, m.FileCount, m.TotalBytes, m.TreeSHA256)
		return nil

	case "verify":
		fs := flag.NewFlagSet("verify", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		root := fs.String("root", "", "Nox data directory")
		manifestPath := fs.String("manifest", "", "sealed manifest path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *root == "" || *manifestPath == "" {
			return errors.New("verify requires -root and -manifest")
		}
		expected, err := readManifest(*manifestPath)
		if err != nil {
			return err
		}
		actual, err := scanTree(*root, expected.ID, expected.RootName)
		if err != nil {
			return err
		}
		if err := compareManifests(expected, actual); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "verified %s: %d files, %d bytes, tree SHA-256 %s\n", expected.ID, expected.FileCount, expected.TotalBytes, expected.TreeSHA256)
		return nil

	case "code-verify":
		fs := flag.NewFlagSet("code-verify", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		root := fs.String("root", "", "Nox data directory")
		manifestPath := fs.String("manifest", "", "sealed GAME.EXE code-range manifest")
		oracleManifestPath := fs.String("oracle-manifest", "", "sealed full-tree oracle manifest")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *root == "" || *manifestPath == "" || *oracleManifestPath == "" {
			return errors.New("code-verify requires -root, -manifest, and -oracle-manifest")
		}
		m, err := verifyCodeOracle(*root, *manifestPath, *oracleManifestPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "verified %s: %d GAME.EXE code ranges, %d data ranges\n", m.OracleID, len(m.Functions), len(m.Data))
		return nil

	default:
		return usageError()
	}
}

func usageError() error {
	return errors.New("usage: noxoracle seal -root DIR -out FILE -id ID | noxoracle verify -root DIR -manifest FILE | noxoracle code-verify -root DIR -manifest FILE -oracle-manifest FILE")
}

func scanTree(root, id, rootName string) (*manifest, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("manifest id is empty")
	}
	info, err := os.Lstat(root)
	if err != nil {
		return nil, fmt.Errorf("inspect root: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, fmt.Errorf("root must be a real directory, not a link or special file: %s", root)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}

	var records []fileRecord
	casePaths := make(map[string]string)
	err = filepath.WalkDir(absRoot, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == absRoot {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not permitted in an oracle: %s", name)
		}
		if entry.IsDir() {
			return nil
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if !entryInfo.Mode().IsRegular() {
			return fmt.Errorf("special files are not permitted in an oracle: %s", name)
		}
		rel, err := filepath.Rel(absRoot, name)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if err := validatePath(rel); err != nil {
			return err
		}
		folded := strings.ToLower(rel)
		if previous, ok := casePaths[folded]; ok && previous != rel {
			return fmt.Errorf("case-insensitive path collision: %q and %q", previous, rel)
		}
		casePaths[folded] = rel
		digest, size, err := hashFile(name)
		if err != nil {
			return err
		}
		records = append(records, fileRecord{Path: rel, Size: size, SHA256: digest})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", root, err)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Path < records[j].Path })
	m := &manifest{
		Schema:    manifestSchema,
		ID:        id,
		RootName:  rootName,
		FileCount: len(records),
		Files:     records,
	}
	for _, rec := range records {
		m.TotalBytes += rec.Size
	}
	m.TreeSHA256 = recordsDigest(records)
	return m, nil
}

func hashFile(name string) (string, int64, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil {
		return "", 0, err
	}
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	after, err := f.Stat()
	if err != nil {
		return "", 0, err
	}
	if n != before.Size() || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return "", 0, fmt.Errorf("file changed while hashing: %s", name)
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func recordsDigest(records []fileRecord) string {
	h := sha256.New()
	for _, rec := range records {
		io.WriteString(h, rec.Path)
		h.Write([]byte{0})
		io.WriteString(h, strconv.FormatInt(rec.Size, 10))
		h.Write([]byte{0})
		io.WriteString(h, rec.SHA256)
		h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func validatePath(name string) error {
	if name == "" || name == "." || path.IsAbs(name) || path.Clean(name) != name || strings.HasPrefix(name, "../") {
		return fmt.Errorf("invalid manifest path: %q", name)
	}
	if !utf8.ValidString(name) || strings.ContainsAny(name, "\\\x00\r\n\t") {
		return fmt.Errorf("manifest path is not portable: %q", name)
	}
	return nil
}

func writeManifestExclusive(name string, m *manifest) (retErr error) {
	if err := validateManifest(m); err != nil {
		return err
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("refusing to replace sealed manifest %s", name)
		}
		return err
	}
	defer func() {
		if err := f.Close(); retErr == nil && err != nil {
			retErr = err
		}
		if retErr != nil {
			_ = os.Remove(name)
		}
	}()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return err
	}
	return f.Sync()
}

func readManifest(name string) (*manifest, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var m manifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("manifest has trailing JSON data")
		}
		return nil, fmt.Errorf("decode manifest trailer: %w", err)
	}
	if err := validateManifest(&m); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}
	return &m, nil
}

func validateManifest(m *manifest) error {
	if m.Schema != manifestSchema {
		return fmt.Errorf("unsupported schema %q", m.Schema)
	}
	if strings.TrimSpace(m.ID) == "" || strings.TrimSpace(m.RootName) == "" {
		return errors.New("id and root_name must be non-empty")
	}
	if m.FileCount != len(m.Files) {
		return fmt.Errorf("file_count is %d, records contain %d", m.FileCount, len(m.Files))
	}
	var total int64
	for i, rec := range m.Files {
		if err := validatePath(rec.Path); err != nil {
			return err
		}
		if i > 0 && m.Files[i-1].Path >= rec.Path {
			return fmt.Errorf("file records are not strictly sorted at %q", rec.Path)
		}
		if rec.Size < 0 {
			return fmt.Errorf("negative size for %q", rec.Path)
		}
		digest, err := hex.DecodeString(rec.SHA256)
		if err != nil || len(digest) != sha256.Size || strings.ToLower(rec.SHA256) != rec.SHA256 {
			return fmt.Errorf("invalid SHA-256 for %q", rec.Path)
		}
		total += rec.Size
	}
	if total != m.TotalBytes {
		return fmt.Errorf("total_bytes is %d, records total %d", m.TotalBytes, total)
	}
	if got := recordsDigest(m.Files); got != m.TreeSHA256 {
		return fmt.Errorf("tree_sha256 mismatch: manifest has %s, records produce %s", m.TreeSHA256, got)
	}
	return nil
}

func compareManifests(expected, actual *manifest) error {
	var missing, extra, changed []string
	i, j := 0, 0
	for i < len(expected.Files) || j < len(actual.Files) {
		switch {
		case i == len(expected.Files):
			extra = append(extra, actual.Files[j].Path)
			j++
		case j == len(actual.Files):
			missing = append(missing, expected.Files[i].Path)
			i++
		case expected.Files[i].Path < actual.Files[j].Path:
			missing = append(missing, expected.Files[i].Path)
			i++
		case expected.Files[i].Path > actual.Files[j].Path:
			extra = append(extra, actual.Files[j].Path)
			j++
		default:
			if expected.Files[i].Size != actual.Files[j].Size || expected.Files[i].SHA256 != actual.Files[j].SHA256 {
				changed = append(changed, expected.Files[i].Path)
			}
			i++
			j++
		}
	}
	if len(missing) == 0 && len(extra) == 0 && len(changed) == 0 && expected.TreeSHA256 == actual.TreeSHA256 {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "oracle mismatch: missing=%d extra=%d changed=%d", len(missing), len(extra), len(changed))
	appendDiff := func(label string, names []string) {
		limit := len(names)
		if limit > 20 {
			limit = 20
		}
		for _, name := range names[:limit] {
			fmt.Fprintf(&b, "\n  %s %s", label, name)
		}
		if len(names) > limit {
			fmt.Fprintf(&b, "\n  %s ... and %d more", label, len(names)-limit)
		}
	}
	appendDiff("missing", missing)
	appendDiff("extra", extra)
	appendDiff("changed", changed)
	return errors.New(b.String())
}
