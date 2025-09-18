package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupEchoServer(ctrl *gomock.Controller) (*Server, *repository.MockRepositoryInterface) {
	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	server := &Server{
		Repository: mockRepo,
	}
	return server, mockRepo
}

func TestGetHello(t *testing.T) {
	e := echo.New()
	s := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	params := generated.GetHelloParams{Id: 42}

	if assert.NoError(t, s.GetHello(c, params)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var resp generated.HelloResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Hello User 42", resp.Message)
	}
}

func TestPostEstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	e := echo.New()
	s, mockRepo := setupEchoServer(ctrl)

	estateID := uuid.New()
	mockRepo.EXPECT().
		CreateEstate(gomock.Any(), 10, 20).
		Return(repository.Estate{ID: estateID, Length: 10, Width: 20}, nil)

	body := generated.EstateRequest{Length: 10, Width: 20}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/estate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := s.PostEstate(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp generated.EstateResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, estateID.String(), resp.Id)
}

func TestPostEstateIdTree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	e := echo.New()
	s, mockRepo := setupEchoServer(ctrl)

	estateID := uuid.New()
	treeID := uuid.New()
	mockRepo.EXPECT().
		AddTree(gomock.Any(), estateID, 5, 1, 1).
		Return(repository.Tree{ID: treeID, EstateID: estateID, Height: 5, X: 1, Y: 1}, nil)

	body := generated.TreeRequest{Height: 5, X: 1, Y: 1}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/estate/"+estateID.String()+"/tree", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := s.PostEstateIdTree(c, estateID.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp generated.TreeResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, treeID.String(), resp.Id)
}

func TestGetEstateIdStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	e := echo.New()
	s, mockRepo := setupEchoServer(ctrl)

	estateID := uuid.New()
	mockRepo.EXPECT().
		GetStats(gomock.Any(), estateID).
		Return(repository.Stats{Count: 2, Min: 5, Max: 10, Median: 7.5}, nil)

	req := httptest.NewRequest(http.MethodGet, "/estate/"+estateID.String()+"/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := s.GetEstateIdStats(c, estateID.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp generated.StatsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Count)
	assert.Equal(t, 5, resp.Min)
	assert.Equal(t, 10, resp.Max)
	assert.Equal(t, 7.5, resp.Median)
}

func TestGetEstateIdDronePlan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	e := echo.New()
	s, mockRepo := setupEchoServer(ctrl)

	estateID := uuid.New()
	maxDistance := 50
	mockRepo.EXPECT().
		GetDronePlan(gomock.Any(), estateID, maxDistance).
		Return(maxDistance, 3, 2, nil)

	params := generated.GetEstateIdDronePlanParams{Distance: &maxDistance}
	req := httptest.NewRequest(http.MethodGet, "/estate/"+estateID.String()+"/drone-plan", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := s.GetEstateIdDronePlan(c, estateID.String(), params)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp generated.DronePlanResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, maxDistance, resp.Distance)
	assert.NotNil(t, resp.Rest)
	assert.Equal(t, 3, resp.Rest.X)
	assert.Equal(t, 2, resp.Rest.Y)
}
