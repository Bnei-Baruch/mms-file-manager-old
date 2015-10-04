package file_manager

import (
	r "github.com/dancannon/gorethink"
	"path/filepath"
	"time"
)

type File struct {
	Id        string    `gorethink:"id,omitempty"`
	FilePath  string    `gorethink:"file_path"`
	FileName  string    `gorethink:"file_name"`
	Status    string    `gorethink:"status"`
	CreatedAt time.Time `gorethink:"created_at"`
	UpdatedAt time.Time `gorethink:"updated_at"`
}

const fileTableName = "files"

const (
	NewFile = iota
	InvalidFile
)

var FileStatuses = [...]string{
	"NEW",
	"INVALID",
}

func (fm *FileManager) FindOneFile(fileName string) (*File, error) {
	cursor, err := r.DB(fm.services.DbName).Table(fileTableName).Filter(r.Row.Field("file_name").Eq(fileName)).Run(fm.services.DB)
	if err != nil {
		l.Println(err)
		return nil, err
	}
	defer cursor.Close()

	if cursor.IsNil() {
		l.Print("Row not found")
		return nil, nil
	}

	file := File{}
	cursor.One(&file)

	return &file, nil
}

func (fm *FileManager) CreateFileRecord(filePath string) (*File, error) {
	defer func() {
		if r := recover(); r != nil {
			l.Println("Recovered in f", r)
		}
	}()

	file := File{
		FilePath:  filePath,
		FileName:  filepath.Base(filePath),
		Status:    FileStatuses[NewFile],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	res, err := r.Table(fileTableName).Insert(file).RunWrite(fm.services.DB)

	if err != nil {
		l.Println("Create file record issue", err, res)
		return nil, err
	}
	file.Id = res.GeneratedKeys[0]
	l.Println("File was created: ", file)

	return &file, nil
}
