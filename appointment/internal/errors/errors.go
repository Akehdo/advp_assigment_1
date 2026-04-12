package errors

import "errors"

var ErrDoctorServiceUnavailable = errors.New("doctor service is unavailable")
var ErrDoctorNotFound = errors.New("doctor not found")
