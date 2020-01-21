package handlerutil

import (
	"encoding/json"
	"errors"
	scimjson "github.com/imulab/go-scim/pkg/json"
	"github.com/imulab/go-scim/pkg/prop"
	"github.com/imulab/go-scim/pkg/spec"
	"net/http"
)

// WriteResourceToResponse writes the given resource to http.ResponseWriter, respecting the attributes or excludedAttributes
// specified through options. Any error during the process will be returned.
// Apart from writing the JSON representation of the resource to body, this method also sets Content-Type header to
// application/json+scim; sets Location header to resource's meta.location field, if any; and sets ETag header to
// resource's meta.version field, if any. This method does not set response status, which should be set before calling
// this method.
func WriteResourceToResponse(rw http.ResponseWriter, resource *prop.Resource, options ...scimjson.Options) error {
	raw, jsonErr := scimjson.Serialize(resource, options...)
	if jsonErr != nil {
		return jsonErr
	}

	rw.Header().Set("Content-Type", "application/json+scim")
	if location := resource.MetaLocationOrEmpty(); len(location) > 0 {
		rw.Header().Set("Location", location)
	}
	if version := resource.MetaVersionOrEmpty(); len(version) > 0 {
		rw.Header().Set("ETag", version)
	}

	_, writeErr := rw.Write(raw)
	return writeErr
}

// WriteError writes the error to the http.ResponseWriter. Any error during the process will be returned.
// If the cause of the error (determined using errors.Unwrap) is a *spec.Error, the cause status and scimType will be
// used together with the error's message as detail. If the cause is not a *spec.Error, spec.ErrInternal is used instead.
// This method also writes the http status with the error's defined status, and set Content-Type header to application/json+scim.
func WriteError(rw http.ResponseWriter, err error) error {
	var errMsg = struct {
		Schemas  []string `json:"schemas"`
		Status   int      `json:"status"`
		ScimType string   `json:"scimType"`
		Detail   string   `json:"detail"`
	}{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		Detail:  err.Error(),
	}

	cause := errors.Unwrap(err)
	if scimError, ok := cause.(*spec.Error); ok {
		errMsg.Status = scimError.Status
		errMsg.ScimType = scimError.Type
	} else {
		errMsg.Status = spec.ErrInternal.Status
		errMsg.ScimType = spec.ErrInternal.Type
	}

	rw.WriteHeader(errMsg.Status)
	rw.Header().Set("Content-Type", "application/json+scim")

	raw, jsonErr := json.Marshal(errMsg)
	if jsonErr != nil {
		return jsonErr
	}

	_, writeErr := rw.Write(raw)
	return writeErr
}
