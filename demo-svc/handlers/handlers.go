package handlers

import (
	"net/http"

	"go.uber.org/zap"
)

type DemoHandler struct {
	logger  *zap.SugaredLogger
	version string
}

func NewDemoHandler(logger *zap.SugaredLogger, version string) *DemoHandler {
	return &DemoHandler{
		logger:  logger,
		version: version,
	}
}

func (fh *DemoHandler) HandleOk(w http.ResponseWriter, r *http.Request) {
	fh.logger.Infow("Received request on", "path", r.URL.Path)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "version": "` + fh.version + `"}`))
}

func (fh *DemoHandler) HandleUserError(w http.ResponseWriter, r *http.Request) {
	fh.logger.Infow("Received request on", "path", r.URL.Path)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{"status": "Bad Request", "version": "` + fh.version + `"}`))

}

func (fh *DemoHandler) HandleServerError(w http.ResponseWriter, r *http.Request) {
	fh.logger.Infow("Received request on", "path", r.URL.Path)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status": "Internal Server Error", "version": "` + fh.version + `"}`))
}
