// This file contains types that are used in the repository layer.
package repository

import "github.com/google/uuid"

type GetTestByIdInput struct {
	Id string
}

type GetTestByIdOutput struct {
	Name string
}

type Estate struct {
	ID     uuid.UUID
	Length int
	Width  int
}

type Tree struct {
	ID       uuid.UUID
	EstateID uuid.UUID
	Height   int
	X        int
	Y        int
}

type Stats struct {
	Count  int
	Min    int
	Max    int
	Median float64
}
