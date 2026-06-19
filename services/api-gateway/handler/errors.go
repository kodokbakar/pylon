package handler

import (
	"net/http"

	"connectrpc.com/connect"

	"github.com/kodokbakar/pylon/internal/response"
)

func writeConnectError(w http.ResponseWriter, err error) {
	connectErr, ok := err.(*connect.Error)
	if !ok {
		response.Error(w, http.StatusBadGateway, "upstream_error", err.Error())
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
		response.Error(w, http.StatusBadGateway, "service_unavailable", connectErr.Message())
	default:
		response.Error(w, http.StatusBadGateway, "upstream_error", connectErr.Message())
	}
}
