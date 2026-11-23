package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type problem struct {
	Type     string           `json:"type"`
	Title    string           `json:"title"`
	Status   int              `json:"status"`
	Detail   string           `json:"detail,omitempty"`
	Instance string           `json:"instance,omitempty"`
	Code     string           `json:"code,omitempty"`
	Errors   []map[string]any `json:"errors,omitempty"`
}

func ProblemErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	m runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	st, _ := status.FromError(err)
	httpStatus := runtime.HTTPStatusFromCode(st.Code())

	p := problem{
		Type:     "about:blank",
		Title:    http.StatusText(httpStatus),
		Status:   httpStatus,
		Detail:   st.Message(),
		Instance: r.URL.Path,
		Code:     st.Code().String(),
	}

	for _, d := range st.Details() {
		switch t := d.(type) {
		case *errdetails.BadRequest:
			for _, v := range t.FieldViolations {
				p.Errors = append(p.Errors, map[string]any{
					"field":   v.Field,
					"message": v.Description,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(p)
}
