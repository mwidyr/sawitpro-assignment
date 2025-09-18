package handler

import (
	"net/http"
	"strconv"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// (GET /hello)
func (s *Server) GetHello(c echo.Context, params generated.GetHelloParams) error {
	return c.JSON(http.StatusOK, generated.HelloResponse{
		Message: "Hello User " + strconv.Itoa(params.Id),
	})
}

// (POST /estate)
func (s *Server) PostEstate(c echo.Context) error {
	var req generated.EstateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "invalid request"})
	}

	estate, err := s.Repository.CreateEstate(c.Request().Context(), req.Length, req.Width)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, generated.EstateResponse{
		Id: estate.ID.String(),
	})
}

// (POST /estate/{id}/tree)
func (s *Server) PostEstateIdTree(c echo.Context, id string) error {
	estateID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "invalid estate id"})
	}

	var req generated.TreeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "invalid request"})
	}

	tree, err := s.Repository.AddTree(c.Request().Context(), estateID, req.Height, req.X, req.Y)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, generated.TreeResponse{
		Id: tree.ID.String(),
	})
}

// (GET /estate/{id}/stats)
func (s *Server) GetEstateIdStats(c echo.Context, id string) error {
	estateID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "invalid estate id"})
	}

	stats, err := s.Repository.GetStats(c.Request().Context(), estateID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, generated.StatsResponse{
		Count:  stats.Count,
		Min:    stats.Min,
		Max:    stats.Max,
		Median: stats.Median,
	})
}

// (GET /estate/{id}/drone-plan)
func (s *Server) GetEstateIdDronePlan(c echo.Context, id string, params generated.GetEstateIdDronePlanParams) error {
	estateID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "invalid estate id"})
	}

	maxDistance := 0
	if params.Distance != nil {
		maxDistance = *params.Distance
	}

	total, restX, restY, err := s.Repository.GetDronePlan(c.Request().Context(), estateID, maxDistance)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: err.Error()})
	}

	resp := generated.DronePlanResponse{
		Distance: total,
	}

	// If maxDistance > 0 and we didn't finish the plan, return rest coordinates
	if maxDistance > 0 && total == maxDistance {
		resp = generated.DronePlanResponse{
			Distance: total,
			Rest: &struct {
				X int `json:"x"`
				Y int `json:"y"`
			}{
				X: restX,
				Y: restY,
			},
		}
	}

	return c.JSON(http.StatusOK, resp)
}
