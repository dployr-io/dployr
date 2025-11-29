// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package shared

import "net/http"

// ErrorCode is an identifier for an error that can be surfaced to clients
// and logged across services.
type ErrorCode string

// ErrorDescriptor defines a specific error, with its code, HTTP status and
// human-readable message. It's the source of truth for HTTP responses and
// OpenAPI error objects.
type ErrorDescriptor struct {
	Code       ErrorCode
	HTTPStatus int
	Message    string
}

const (
	// Request-level errors
	ErrorRequestMissingParams    ErrorCode = "request.missing_params"
	ErrorRequestBadRequest       ErrorCode = "request.bad_request"
	ErrorRequestMethodNotAllowed ErrorCode = "request.method_not_allowed"

	// Auth errors
	ErrorAuthUnauthorized ErrorCode = "auth.unauthorized"
	ErrorAuthForbidden    ErrorCode = "auth.forbidden"

	// Resource errors
	ErrorResourceNotFound ErrorCode = "resource.not_found"

	// Runtime / internal errors
	ErrorRuntimeInternalServer      ErrorCode = "runtime.internal_server_error"
	ErrorInstanceRegistrationFailed ErrorCode = "instance.registration_failed"
)

// Errors is a structured catalog of application errors that are currently used
// across the HTTP handlers.
var Errors = struct {
	Request struct {
		MissingParams    ErrorDescriptor
		BadRequest       ErrorDescriptor
		MethodNotAllowed ErrorDescriptor
	}
	Auth struct {
		Unauthorized ErrorDescriptor
		Forbidden    ErrorDescriptor
	}
	Resource struct {
		NotFound ErrorDescriptor
	}
	Instance struct {
		RegistrationFailed ErrorDescriptor
	}
	Runtime struct {
		InternalServer ErrorDescriptor
	}
}{
	Request: struct {
		MissingParams    ErrorDescriptor
		BadRequest       ErrorDescriptor
		MethodNotAllowed ErrorDescriptor
	}{
		MissingParams:    ErrorDescriptor{Code: ErrorRequestMissingParams, HTTPStatus: http.StatusBadRequest, Message: "Missing required parameters"},
		BadRequest:       ErrorDescriptor{Code: ErrorRequestBadRequest, HTTPStatus: http.StatusBadRequest, Message: "Bad request"},
		MethodNotAllowed: ErrorDescriptor{Code: ErrorRequestMethodNotAllowed, HTTPStatus: http.StatusMethodNotAllowed, Message: "Method not allowed"},
	},
	Auth: struct {
		Unauthorized ErrorDescriptor
		Forbidden    ErrorDescriptor
	}{
		Unauthorized: ErrorDescriptor{Code: ErrorAuthUnauthorized, HTTPStatus: http.StatusUnauthorized, Message: "Unauthorized"},
		Forbidden:    ErrorDescriptor{Code: ErrorAuthForbidden, HTTPStatus: http.StatusForbidden, Message: "Forbidden"},
	},
	Resource: struct {
		NotFound ErrorDescriptor
	}{
		NotFound: ErrorDescriptor{Code: ErrorResourceNotFound, HTTPStatus: http.StatusNotFound, Message: "Resource not found"},
	},
	Runtime: struct {
		InternalServer ErrorDescriptor
	}{
		InternalServer: ErrorDescriptor{Code: ErrorRuntimeInternalServer, HTTPStatus: http.StatusInternalServerError, Message: "Internal server error"},
	},
	Instance: struct {
		RegistrationFailed ErrorDescriptor
	}{
		RegistrationFailed: ErrorDescriptor{Code: ErrorInstanceRegistrationFailed, HTTPStatus: http.StatusInternalServerError, Message: "Instance registration failed"},
	},
}
