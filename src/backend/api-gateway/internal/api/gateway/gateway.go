package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

func CookieSetter() runtime.ServeMuxOption {
	return runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
		if strings.EqualFold(key, "set-cookie") {
			return "Set-Cookie", true
		}
		return runtime.DefaultHeaderMatcher(key)
	})
}

func CookieGetter() runtime.ServeMuxOption {
	return runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
		cookies := r.Cookies()
		if len(cookies) == 0 {
			return nil
		}

		md := metadata.MD{}
		for _, c := range cookies {
			key := "cookie-" + strings.ToLower(c.Name)
			md.Append(key, c.Value)
		}
		return md
	})
}

// TODO: Attempt to move dependency wiring here later
