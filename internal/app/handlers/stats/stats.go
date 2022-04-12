package stats

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/repository"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
)

// Handler struct
type Handler struct {
	s repository.Repository
	l *zap.Logger
}

// NewStats implement stats handler
func NewStats(s repository.Repository, l *zap.Logger) *Handler {
	return &Handler{s, l}
}

// ServeHTTP implement logic for ping handler
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ts, _ := configs.Instance().Param(configs.TrustedSubnet)

	if ts != "" {
		h.l.Info("TS: " + ts)
		_, ipv4Net, err := net.ParseCIDR(ts)
		if err != nil {
			h.l.Info("Can't parse CIDR")
			http.Error(w, er.ErrBadResponse.Error(), http.StatusForbidden)
			return
		}

		ip, err := getIP(r)
		if err != nil {
			h.l.Info("Can't parse IP")
			http.Error(w, er.ErrBadResponse.Error(), http.StatusForbidden)
			return
		}

		if !ipv4Net.Contains(ip) {
			h.l.Info("Can't contain IP" + ip.String())
			http.Error(w, er.ErrBadResponse.Error(), http.StatusForbidden)
			return
		}
	}

	// Main logic
	urlCount := h.s.UrlCount()
	userCount := h.s.UserCount()

	result := struct {
		Urls  int `json:"urls"`
		Users int `json:"users"`
	}{Urls: urlCount, Users: userCount}

	body, err := json.Marshal(result)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Prepare response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
	}
}

// getIP get current ip address for user
func getIP(r *http.Request) (net.IP, error) {
	// use request method
	radr := r.RemoteAddr
	fmt.Println(radr)
	// return format host:port
	// need only host
	ip, _, err := net.SplitHostPort(radr)
	if err != nil {
		return nil, err
	}
	// parse another format
	ip1 := net.ParseIP(ip)
	// get header  X-Real-IP
	ip = r.Header.Get("X-Real-IP")

	// unknown format
	ip2 := net.ParseIP(ip)
	if ip2 == nil {
		// if X-Real-IP empty, try X-Forwarded-For
		ips := r.Header.Get("X-Forwarded-For")
		// slit addresses
		splitIps := strings.Split(ips, ",")
		// only first
		ip = splitIps[0]
		// parse ip address
		ip2 = net.ParseIP(ip)
	}

	if ip1.Equal(ip2) {
		return ip1, nil
	}

	return nil, errors.New("no guaranteed ip")
}
