package repository

import "errors"

var ErrNotFound = errors.New("resource not found")

var ErrVersionConflict = errors.New("file version conflict")

var ErrFileAlreadyExists = errors.New("file with the same name already exists")
