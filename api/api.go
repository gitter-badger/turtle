/*
 *  Turtle - Rock Solid Cluster Management
 *  Copyright DesertBit
 *  Author: Roland Singer
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/mitchellh/mapstructure"
)

const (
	// The API version.
	Version = "0.1"
)

//####################//
//### Public Types ###//
//####################//

type Status int

const (
	StatusSuccess Status = 1 << iota
	StatusError   Status = 1 << iota
)

// Request is the request body of each turtle request.
// The specific request is stored in the data value.
type Request struct {
	// The API version.
	Version string

	// The requested action.
	Type Type

	// The specific request type.
	Data interface{}
}

// Response is the response body of each turtle response.
// The specific reply is stored in the data value.
type Response struct {
	// The API version.
	Version string

	// The success state of the request,
	Status Status

	// The specific response type.
	Data interface{}
}

// ResponseError is always set to the response data, if the status value is set to error.
type ResponseError struct {
	ErrorMessage string
}

//##############//
//### Public ###//
//##############//

// NewRequest creates a new request value.
// The API Version is also set.
func NewRequest(t Type, data interface{}) *Request {
	return &Request{
		Version: Version,
		Type:    t,
		Data:    data,
	}
}

// NewRequestFromJSON constructs a new request from a JSON string or io.Reader.
func NewRequestFromJSON(jsonData interface{}) (*Request, error) {
	var jsonReader io.Reader

	switch v := jsonData.(type) {
	case string:
		jsonReader = strings.NewReader(v)
	case io.Reader:
		jsonReader = v
	default:
		return nil, fmt.Errorf("NewRequest: expected string or io.Reader type as argument.")
	}

	// Decode the JSON to the request value.
	decoder := json.NewDecoder(jsonReader)

	var r Request
	err := decoder.Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	return &r, nil
}

// MapTo maps the data values to a struct.
// val must be a pointer to a struct.
func (r *Request) MapTo(val interface{}) error {
	return decode(r.Data, val)
}

// ToJSON marshals the request to JSON.
func (r *Request) ToJSON() ([]byte, error) {
	// Marshal the reponse to a JSON byte slice.
	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request to JSON: %v", err)
	}

	return data, nil
}

// NewResponse creates a new response value.
// The API Version is also set.
func NewResponse() *Response {
	return &Response{
		Version: Version,
	}
}

// NewResponseFromJSON constructs a new response from a JSON string or io.Reader.
func NewResponseFromJSON(jsonData interface{}) (*Response, error) {
	var jsonReader io.Reader

	switch v := jsonData.(type) {
	case string:
		jsonReader = strings.NewReader(v)
	case io.Reader:
		jsonReader = v
	default:
		return nil, fmt.Errorf("NewResponse: expected string or io.Reader type as argument.")
	}

	// Decode the JSON to the request value.
	decoder := json.NewDecoder(jsonReader)

	var r Response
	err := decoder.Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	return &r, nil
}

// MapTo maps the data values to a struct.
// val must be a pointer to a struct.
func (r *Response) MapTo(val interface{}) error {
	return decode(r.Data, val)
}

// ToJSON marshals the response to JSON.
func (r *Response) ToJSON() ([]byte, error) {
	// Marshal the response to a JSON byte slice.
	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response to JSON: %v", err)
	}

	return data, nil
}

//###############//
//### Private ###//
//###############//

// decode an interface map to a struct pointer.
func decode(input interface{}, result interface{}) error {
	// Create the decode config. Set to weakly typed input decode mode.
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           result,
	}

	// Create the map to struct decoder.
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("failed to create interface map to struct decoder: %v", err)
	}

	// Decode the input map to a struct.
	err = decoder.Decode(input)
	if err != nil {
		return fmt.Errorf("failed to decode interface map to struct: %v", err)
	}

	return nil
}
