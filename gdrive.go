package storage

import (
	"bytes"
	"errors"

	"google.golang.org/api/drive/v3"
)

// GetParentID graps the folder ID for folder name.
func (ds *DriverService) GetParentID(folder string) (string, error) {
	file, err := ds.Files.List().Q(`name = "` + folder + `" and '0BzaYO4E7QW9VN2ZuTjVBc1Zydzg' in parents and mimeType = 'application/vnd.google-apps.folder'`).Do()
	if err != nil {
		return "", err
	}
	if len(file.Files) > 1 {
		return "", errors.New("too broad of a search term; too many folders")
	}
	if len(file.Files) == 0 {
		return "", errors.New("can't find folder")
	}
	return file.Files[0].Id, nil
}

// Save saves file to drive.
func (ds *DriverService) Save(b []byte, f *drive.File) (*drive.File, error) {

	r := bytes.NewReader(b)

	file, err := ds.Files.Create(f).Media(r).Do()
	if err != nil {
		return nil, err
	}
	return file, nil
}

// EmptyTrash permanently deletes files from trash
func (ds *DriverService) EmptyTrash() {
	ds.Files.EmptyTrash().Do()
}
