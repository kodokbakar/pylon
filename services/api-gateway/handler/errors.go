package handler

import (
	"errors"
	"net/http"

	"connectrpc.com/connect"

	"github.com/kodokbakar/pylon/internal/response"
)

func writeConnectError(w http.ResponseWriter, err error) {
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		response.Error(w, http.StatusBadGateway, "upstream_error", "internal service error")
		return
	}

	switch connectErr.Code() {
	case connect.CodeInvalidArgument:
		response.Error(w, http.StatusBadRequest, "bad_request", connectErr.Message())
	case connect.CodeUnauthenticated:
		response.Error(w, http.StatusUnauthorized, "unauthorized", connectErr.Message())
	case connect.CodePermissionDenied:
		response.Error(w, http.StatusForbidden, "forbidden", connectErr.Message())
	case connect.CodeNotFound:
		response.Error(w, http.StatusNotFound, "not_found", connectErr.Message())
	case connect.CodeAlreadyExists:
		response.Error(w, http.StatusConflict, "already_exists", connectErr.Message())
	case connect.CodeFailedPrecondition:
		response.Error(w, http.StatusPreconditionFailed, "failed_precondition", connectErr.Message())
	case connect.CodeUnavailable:
		response.Error(w, http.StatusBadGateway, "service_unavailable", "internal service unavailable")
	default:
		response.Error(w, http.StatusBadGateway, "upstream_error", "internal service error")
	}
}
