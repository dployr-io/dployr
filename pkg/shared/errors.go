package shared

type AppError string

const (
	BadRequest   AppError = "bad_request"
	Unavailable  AppError = "unavailable"
	RuntimeError AppError = "runtime_error"
)
