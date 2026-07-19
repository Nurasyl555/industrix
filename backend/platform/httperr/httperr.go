// Package httperr turns errors returned by services into HTTP responses.
//
// Every module used to carry its own identical copy of this mapping, and none
// of them logged anything: an unexpected error became a bare
// {"code":"INTERNAL","message":"Something went wrong"} with the cause thrown
// away, so debugging a 500 meant reproducing it by hand. Centralising it here
// means the cause is always recorded.
package httperr

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/logger"
)

var log = logger.New("http-error")

// Respond writes the HTTP response for err.
//
// A domain error (*errors.Error) is expected — it carries its own code and a
// message meant for the user, and is returned as-is. Anything else is a bug or
// an infrastructure failure: it is logged with the route that produced it, and
// the client gets a generic 500 so internal details never leak.
func Respond(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.Error); ok {
		return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
	}

	log.Error().
		Err(err).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Msg("unhandled error serving request")

	return c.Status(http.StatusInternalServerError).
		JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}
