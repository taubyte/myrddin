package myrddin

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
	"github.com/spf13/afero/zipfs"
)

func (m *Myrddin) ReadFile(uri string) ([]byte, error) {
	if strings.HasPrefix(uri, "myrddin+") == false {
		return nil, fmt.Errorf("Uri missing myrddin+ prefix")
	}

	_uri, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("Myrddin parsing uri(`%s`) failed with: %w", uri, err)
	}

	switch _uri.Scheme {
	case "myrddin+file":
		content, err := m.readFileOS(_uri.Path)
		if err != nil {
			return nil, fmt.Errorf("Myrddin parsing uri(`%s`) failed with: %w", uri, err)
		}
		return content, nil

	default:
		return nil, fmt.Errorf("I don't know how to get %s", _uri.Scheme)
	}

}

func (m *Myrddin) Load(uri string) error {
	_uri, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("Myrddin parsing uri(`%s`) failed with: %w", uri, err)
	}

	if _uri.Scheme == "" {
		absPath, err := filepath.Abs(uri)
		if err != nil {
			return fmt.Errorf("Myrddin parsing uri(`%s`) failed with: %w", uri, err)
		}
		_uri, err = url.Parse("file://" + absPath)
		if err != nil {
			return fmt.Errorf("Myrddin parsing uri(`%s`) failed with: %w", uri, err)
		}

	}

	switch _uri.Scheme {
	case "file":
		osfs := afero.NewOsFs()
		_path := _uri.Path
		isdir, err := afero.IsDir(osfs, _path)
		if err != nil {
			return fmt.Errorf("Myrddin loading uri(`%s`) failed with: %w", uri, err)
		}
		if isdir == true {
			m.store = afero.NewReadOnlyFs(afero.NewBasePathFs(osfs, _path))
		} else {
			f, err := osfs.Open(_path)
			if err != nil {
				return fmt.Errorf("Myrddin opening uri(`%s`) target failed with: %w", uri, err)
			}
			typeBuff := make([]byte, 512)
			_, err = f.Read(typeBuff)
			if err != nil {
				return fmt.Errorf("Myrddin reading uri(`%s`) target failed with: %w", uri, err)
			}

			//rewind f
			f.Seek(0, 0)

			contentType, err := filetype.Match(typeBuff)
			if err != nil {
				return fmt.Errorf("Myrddin file type of uri(`%s`) target failed with: %w", uri, err)
			}

			switch contentType {
			case matchers.TypeZip:
				zrc, err := zip.OpenReader(_path)
				if err != nil {
					return fmt.Errorf("Myrddin reading uri(`%s`) as zip file failed with: %w", uri, err)
				}
				m.store = zipfs.New(&zrc.Reader)
			case matchers.TypeTar:
				m.store = tarfs.New(tar.NewReader(f))
			case matchers.TypeGz:
				gzf, err := gzip.NewReader(f)
				if err != nil {
					return fmt.Errorf("Myrddin reading uri(`%s`) as Gzip file failed with: %w", uri, err)
				}
				m.store = tarfs.New(tar.NewReader(gzf))
			default:
				return fmt.Errorf("Myrddin unsuported uri(`%s`) file type: %s", _uri, contentType.Extension)
			}
		}
	default:
		return fmt.Errorf("Myrddin parsing uri(`%s`) error: unknown uri scheme `%s`", _uri, _uri.Scheme)
	}

	if m.store == nil {
		return errors.New("Failed to open URI")
	}

	return nil
}
