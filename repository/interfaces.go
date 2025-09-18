// This file contains the interfaces for the repository layer.
// The repository layer is responsible for interacting with the database.
// For testing purpose we will generate mock implementations of these
// interfaces using mockgen. See the Makefile for more information.
package repository

import (
	"context"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	GetTestById(ctx context.Context, input GetTestByIdInput) (output GetTestByIdOutput, err error)
	CreateEstate(ctx context.Context, length, width int) (Estate, error)
	AddTree(ctx context.Context, estateID uuid.UUID, height, x, y int) (Tree, error)
	GetStats(ctx context.Context, estateID uuid.UUID) (Stats, error)
	GetDronePlan(ctx context.Context, estateID uuid.UUID, maxDistance int) (total int, lastX, lastY int, err error)
}
