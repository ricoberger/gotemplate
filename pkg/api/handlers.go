package api

import (
	"net/http"

	"github.com/ricoberger/gotemplate/pkg/api/middleware/errresponse"
	"github.com/ricoberger/gotemplate/pkg/api/request"
	"github.com/ricoberger/gotemplate/pkg/log"

	"github.com/go-chi/render"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("api")

func pingHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "pingHandler")
	defer span.End()

	data := struct {
		Pong string `json:"pong"`
	}{
		"pong",
	}

	render.JSON(w, r, data)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "requestHandler")
	defer span.End()

	url := r.URL.Query().Get("url")
	span.SetAttributes(attribute.Key("url").String(url))

	statusCode, body, err := request.Do(ctx, url)
	if err != nil {
		log.Error(ctx, "Request returned an error", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		errresponse.Render(w, r, err, http.StatusBadRequest, "Request returned an error")
		return
	}

	data := struct {
		Status int    `json:"status"`
		Body   string `json:"body"`
	}{
		statusCode,
		body,
	}

	render.JSON(w, r, data)
}
