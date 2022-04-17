// Package grpcshortener implement interfaces for gRPC server shortener service
// It's just Facade for http handlers from origin logic
// For implementation using map ResponseWriterMap for  http.ResponseWriter struct and calling
// general http handlers
// Server work only "all" user id. For implementation session user id it's need to rewrite logic general handlers,
// but for this is test increment it's not nessasary

package grpcshortener

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/delete"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/ping"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/stats"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/repository"
	proto "github.com/triumphpc/go-musthave-shortener-tpl/pkg/api"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// ShortenerServer implement main methods for gRPC
type ShortenerServer struct {
	// need embedded type proto.Unimplemented<TypeName>
	// for related in future version
	proto.UnimplementedShortenerServer

	s  repository.Repository
	l  *zap.Logger
	db *sql.DB
	p  *worker.Pool
}

// ResponseWriterMap it's bridge for response from main handler
type ResponseWriterMap struct {
	h    http.ResponseWriter
	head http.Header
	buf  bytes.Buffer
	code int
}

// NewResponseWriterMap init new ResponseWriterMap
func NewResponseWriterMap() *ResponseWriterMap {
	rw := ResponseWriterMap{}
	rw.head = make(http.Header)

	return &rw
}

func (rw *ResponseWriterMap) Header() http.Header {
	return rw.head
}

func (rw *ResponseWriterMap) WriteHeader(statusCode int) {
	rw.code = statusCode
}

func (rw *ResponseWriterMap) Write(data []byte) (int, error) {
	return rw.buf.Write(data)
}

// New instance for gRPC server
func New(l *zap.Logger, s repository.Repository, db *sql.DB, p *worker.Pool) *ShortenerServer {
	return &ShortenerServer{proto.UnimplementedShortenerServer{}, s, l, db, p}
}

// AddLink implement add new link
func (s *ShortenerServer) AddLink(ctx context.Context, r *proto.AddLinkRequest) (*proto.AddLinkResponse, error) {
	var response proto.AddLinkResponse

	// Init http handler
	h := handlers.New(s.l, s.s)
	link := r.GetLink().Link

	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(link))
	if err != nil {
		return &response, err
	}

	resp := NewResponseWriterMap()

	h.Save(resp, req)

	response.Code = int32(resp.code)

	response.Link = &proto.ShortLink{
		Link: resp.buf.String(),
	}

	return &response, nil
}

// Ping implement method for gRPC for check DB connection
func (s *ShortenerServer) Ping(context.Context, *proto.PingRequest) (*proto.PingResponse, error) {
	var response proto.PingResponse

	resp := NewResponseWriterMap()
	req, err := http.NewRequest(http.MethodPost, "/ping", strings.NewReader(""))
	if err != nil {
		return &response, err
	}

	handler := ping.NewPing(s.db, s.l)
	handler.ServeHTTP(resp, req)

	response.Code = int32(resp.code)

	return &response, nil
}

// AddBatch implement add batch links implementation
func (s *ShortenerServer) AddBatch(ctx context.Context, r *proto.AddBatchRequest) (*proto.AddBatchResponse, error) {
	response := new(proto.AddBatchResponse)

	// Init http handler
	h := handlers.New(s.l, s.s)
	requestLinks := r.GetLinks()

	var urls []shortlink.URLs
	for _, requestLink := range requestLinks {
		urls = append(urls, shortlink.URLs{ID: requestLink.Id.Id, Origin: requestLink.Link})
	}
	body, err := json.Marshal(urls)
	if err != nil {
		return response, err
	}

	req, err := http.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(string(body)))
	if err != nil {
		return response, err
	}

	resp := NewResponseWriterMap()

	h.BunchSaveJSON(resp, req)
	response.Code = int32(resp.code)

	data := make([]shortlink.ShortURLs, len(urls))
	err = json.Unmarshal(resp.buf.Bytes(), &data)
	if err != nil {
		return response, err
	}

	// Convert to gRPC struct
	shortLinks := make([]*proto.JSONBatchShortLink, len(urls))
	for k, v := range data {
		shortLinks[k] = &proto.JSONBatchShortLink{
			Link: v.Short,
			Id:   &proto.LinkID{Id: v.ID},
		}
	}
	response.Links = shortLinks

	return response, nil
}

// AddJSONLink implement save link in JSON format
func (s *ShortenerServer) AddJSONLink(ctx context.Context, r *proto.AddJSONLinkRequest) (*proto.AddJSONLinkResponse, error) {
	response := new(proto.AddJSONLinkResponse)

	link := r.GetLink().String()

	url := shortlink.URL{URL: link}

	body, err := json.Marshal(url)
	if err != nil {
		return response, err
	}

	req, err := http.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(string(body)))
	if err != nil {
		return response, err
	}

	resp := NewResponseWriterMap()
	// Init http handler
	h := handlers.New(s.l, s.s)

	h.SaveJSON(resp, req)
	response.Code = int32(resp.code)

	data := struct {
		Result string `json:"result"`
	}{}

	err = json.Unmarshal(resp.buf.Bytes(), &data)
	if err != nil {
		return response, err
	}

	response.Link = &proto.JSONShortLink{
		Link: data.Result,
	}

	return response, nil
}

func (s *ShortenerServer) UserLinks(context.Context, *proto.JSONUserLinksRequest) (*proto.JSONUserLinksResponse, error) {
	response := new(proto.JSONUserLinksResponse)

	req, err := http.NewRequest(http.MethodGet, "/api/user/urls", nil)
	if err != nil {
		return response, err
	}

	resp := NewResponseWriterMap()

	// Init http handler
	h := handlers.New(s.l, s.s)
	h.GetUrls(resp, req)

	response.Code = int32(resp.code)

	return response, nil

}

func (s *ShortenerServer) Stats(context.Context, *proto.StatsRequest) (*proto.StatsResponse, error) {
	response := new(proto.StatsResponse)

	handler := stats.NewStats(s.s, s.l)
	req, err := http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	if err != nil {
		return response, err
	}

	resp := NewResponseWriterMap()
	handler.ServeHTTP(resp, req)

	result := struct {
		Urls  int `json:"urls"`
		Users int `json:"users"`
	}{}

	err = json.Unmarshal(resp.buf.Bytes(), &result)
	if err != nil {
		return response, err
	}

	response.Urls = int32(result.Urls)
	response.Users = int32(result.Users)

	return response, nil
}

func (s *ShortenerServer) Delete(ctx context.Context, r *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	response := new(proto.DeleteResponse)

	IDs := r.GetId()

	var linkIDs []string
	for _, id := range IDs {
		linkIDs = append(linkIDs, id.GetId())
	}
	body, err := json.Marshal(linkIDs)
	if err != nil {
		return response, err
	}

	req, err := http.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(string(body)))
	if err != nil {
		return response, err
	}

	handler := delete.New(s.l, s.p)

	resp := NewResponseWriterMap()
	handler.ServeHTTP(resp, req)

	response.Code = int32(resp.code)

	return response, nil

}

func (s *ShortenerServer) Origin(ctx context.Context, r *proto.OriginRequest) (*proto.OriginResponse, error) {
	response := new(proto.OriginResponse)
	id := r.GetLink().Link

	req, err := http.NewRequest(http.MethodGet, "/"+id, strings.NewReader(""))
	if err != nil {
		return response, err
	}
	// Init http handler
	h := handlers.New(s.l, s.s)

	resp := NewResponseWriterMap()

	//Hack to try to fake gorilla/mux vars
	vars := map[string]string{
		"id": id,
	}
	req = mux.SetURLVars(req, vars)
	h.Get(resp, req)

	response.Code = int32(resp.code)
	response.Link = &proto.Link{
		Link: resp.Header().Get("Location"),
	}

	return response, nil

}
