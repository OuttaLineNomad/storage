package storage

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/api/drive/v3"
)

// func (ds *DriverService) GetFile(id string) (*drive.File, error) {
// 	file, err := ds.Files.Get(id).Do()
// 	if err != nil {
// 		return nil, &Error{"GetFile", "files get call", err}
// 	}
// 	file.
// }

// GetParentID grabs the folder ID for folder name.
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

// DriveFile struct to return to user so user has Name and ID to use.
type DriveFile struct {
	Name string
	ID   string
}

// GetFileIDs returns all google sheets found infolder.
func (ds *DriverService) GetFileIDs(folderID string, mimeType ...string) ([]DriveFile, error) {
	query := `'` + folderID + `' in parents and trashed = false `
	mCount := len(mimeType)
	if mCount > 0 {
		query += `and (mimeType = `
		for i, m := range mimeType {
			query += `'` + m + `'`
			if mCount > 1 && i != mCount-1 {
				query += ` or mimeType =`
			}
		}
	}
	query += ")"

	fmt.Println(query)
	fzs, err := ds.Files.List().Q(query).Do()
	if err != nil {
		return nil, err
	}

	filez := []DriveFile{}
	for _, f := range fzs.Files {
		filez = append(filez, DriveFile{f.Name, f.Id})
	}

	return filez, nil
}

// GetFile downloads file from Drive
func (ds *DriverService) GetFile(ID string) (*http.Response, error) {
	return ds.Files.Get(ID).Download()
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
