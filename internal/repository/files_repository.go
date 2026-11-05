package repository

import (
	"database/sql"
	"tools-ecg-backend/internal/models"
)

const (
	getStandplanFilesQuery = "SELECT * FROM files"

	createStandplanFileQuery = "INSERT INTO files (title, description, url) VALUES ($1, $2, $3) RETURNING id"
)

type FilesRepository struct {
	// Add any necessary fields for the repository, such as a database connection
	DB *sql.DB
}

func NewFilesRepository(db *sql.DB) *FilesRepository {
	return &FilesRepository{
		DB: db,
	}
}

func (f *FilesRepository) GetStandplanFiles() ([]models.File, error) {
	var files []models.File

	rows, err := f.DB.Query(getStandplanFilesQuery)
	if err != nil {
		return nil, err
	}
	defer func () {
		err = rows.Close()
	}()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var file models.File
		err := rows.Scan(&file.Id, &file.Title, &file.Description, &file.Url)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return files, nil
}

func (f *FilesRepository) CreateStandplanFile(file models.File)	(int, error) {
	var id int
	err := f.DB.QueryRow(createStandplanFileQuery, file.Title, file.Description, file.Url).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}