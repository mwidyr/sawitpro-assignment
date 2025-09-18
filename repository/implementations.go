package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

func (r *Repository) GetTestById(ctx context.Context, input GetTestByIdInput) (GetTestByIdOutput, error) {
	var output GetTestByIdOutput

	if strings.TrimSpace(input.Id) == "" {
		return output, WrapErr("invalid id: cannot be empty")
	}

	err := r.Db.QueryRowContext(ctx,
		`SELECT name FROM test WHERE id = $1`, input.Id,
	).Scan(&output.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return output, WrapErr(fmt.Sprintf("test with id %s not found", input.Id))
		}
		return output, WrapErrf("query test by id failed: %w", err)
	}

	return output, nil
}

func (r *Repository) CreateEstate(ctx context.Context, length, width int) (Estate, error) {
	if length < 1 || length > 50000 || width < 1 || width > 50000 {
		return Estate{}, WrapErr("estate dimensions must be between 1 and 50000")
	}

	id := uuid.New()
	_, err := r.Db.ExecContext(ctx,
		`INSERT INTO estates (estate_id, estate_length, estate_width) VALUES ($1, $2, $3)`,
		id, length, width,
	)
	if err != nil {
		return Estate{}, WrapErrf("failed to insert estate: %w", err)
	}

	return Estate{ID: id, Length: length, Width: width}, nil
}

func (r *Repository) AddTree(ctx context.Context, estateID uuid.UUID, height, x, y int) (Tree, error) {
	if estateID == uuid.Nil {
		return Tree{}, WrapErr("estateID cannot be empty")
	}
	if height < 1 || height > 30 {
		return Tree{}, WrapErr("tree height must be between 1 and 30")
	}
	if x < 1 || y < 1 {
		return Tree{}, WrapErr("tree coordinates must be positive")
	}

	var estate Estate
	err := r.Db.QueryRowContext(ctx,
		`SELECT estate_id, estate_length, estate_width FROM estates WHERE estate_id = $1`, estateID,
	).Scan(&estate.ID, &estate.Length, &estate.Width)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tree{}, WrapErr(fmt.Sprintf("estate %s not found", estateID))
		}
		return Tree{}, WrapErrf("failed to fetch estate: %w", err)
	}

	// Boundary check
	if x > estate.Length || y > estate.Width {
		return Tree{}, WrapErr("tree coordinates out of estate bounds")
	}

	// Check duplicate tree
	var exists bool
	err = r.Db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM trees WHERE estate_id=$1 AND coordinate_x=$2 AND coordinate_y=$3
		)`,
		estateID, x, y,
	).Scan(&exists)
	if err != nil {
		return Tree{}, WrapErrf("failed to check existing tree: %w", err)
	}
	if exists {
		return Tree{}, WrapErr("tree already exists at this plot")
	}

	id := uuid.New()
	_, err = r.Db.ExecContext(ctx,
		`INSERT INTO trees (tree_id, estate_id, tree_height, coordinate_x, coordinate_y) 
		 VALUES ($1, $2, $3, $4, $5)`,
		id, estateID, height, x, y,
	)
	if err != nil {
		return Tree{}, WrapErrf("failed to insert tree: %w", err)
	}

	return Tree{ID: id, EstateID: estateID, Height: height, X: x, Y: y}, nil
}

func (r *Repository) GetStats(ctx context.Context, estateID uuid.UUID) (Stats, error) {
	if estateID == uuid.Nil {
		return Stats{}, WrapErr("estateID cannot be empty")
	}

	rows, err := r.Db.QueryContext(ctx,
		`SELECT tree_height FROM trees WHERE estate_id = $1`, estateID,
	)
	if err != nil {
		return Stats{}, WrapErrf("failed to fetch stats: %w", err)
	}
	defer rows.Close()

	var heights []int
	for rows.Next() {
		var h int
		if err := rows.Scan(&h); err != nil {
			return Stats{}, WrapErrf("failed to scan tree_height: %w", err)
		}
		heights = append(heights, h)
	}

	if len(heights) == 0 {
		return Stats{Count: 0, Min: 0, Max: 0, Median: 0}, nil
	}

	sort.Ints(heights)
	count := len(heights)
	min := heights[0]
	max := heights[count-1]

	var median float64
	if count%2 == 1 {
		median = float64(heights[count/2])
	} else {
		median = (float64(heights[count/2-1]) + float64(heights[count/2])) / 2.0
	}

	return Stats{Count: count, Min: min, Max: max, Median: median}, nil
}

// If maxDistance > 0, it will stop early once that budget is reached and return the last coordinates.
func (r *Repository) GetDronePlan(ctx context.Context, estateID uuid.UUID, maxDistance int) (total int, lastX, lastY int, err error) {
	if estateID == uuid.Nil {
		return 0, 0, 0, WrapErr("estateID cannot be empty")
	}
	if maxDistance < 0 {
		return 0, 0, 0, WrapErr("maxDistance must be non-negative")
	}

	var estate Estate
	err = r.Db.QueryRowContext(ctx,
		`SELECT estate_id, estate_length, estate_width FROM estates WHERE estate_id = $1`, estateID,
	).Scan(&estate.ID, &estate.Length, &estate.Width)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, 0, WrapErr(fmt.Sprintf("estate %s not found", estateID))
		}
		return 0, 0, 0, WrapErrf("failed to fetch estate: %w", err)
	}

	// Load trees into map
	rows, err := r.Db.QueryContext(ctx,
		`SELECT coordinate_x, coordinate_y, tree_height FROM trees WHERE estate_id=$1`, estateID,
	)
	if err != nil {
		return 0, 0, 0, WrapErrf("failed to fetch trees: %w", err)
	}
	defer rows.Close()

	type coord struct{ X, Y int }
	trees := make(map[coord]int)
	for rows.Next() {
		var x, y, h int
		if err := rows.Scan(&x, &y, &h); err != nil {
			return 0, 0, 0, WrapErrf("failed to scan tree: %w", err)
		}
		trees[coord{X: x, Y: y}] = h
	}

	// Simulation
	x, y, z := 1, 1, 0
	dir := 1 // east = 1, west = -1

	addCost := func(c int) bool {
		total += c
		if maxDistance > 0 && total >= maxDistance {
			lastX, lastY = x, y
			total = maxDistance
			return true // stop
		}
		return false
	}

	for y <= estate.Width {
		for (dir == 1 && x <= estate.Length) || (dir == -1 && x >= 1) {
			target := 1
			if h, ok := trees[coord{X: x, Y: y}]; ok {
				target = h + 1
			}

			if z < target {
				if addCost(target - z) {
					return total, lastX, lastY, nil
				}
				z = target
			} else if z > target {
				if addCost(z - target) {
					return total, lastX, lastY, nil
				}
				z = target
			}

			lastX, lastY = x, y
			if (dir == 1 && x < estate.Length) || (dir == -1 && x > 1) {
				if addCost(10) {
					return total, lastX, lastY, nil
				}
				x += dir
			} else {
				break
			}
		}

		if y < estate.Width {
			if addCost(10) {
				return total, lastX, lastY, nil
			}
			y++
			dir *= -1
		} else {
			break
		}
	}

	// descend
	if z > 0 {
		if addCost(z) {
			return total, lastX, lastY, nil
		}
	}

	return total, lastX, lastY, nil
}
