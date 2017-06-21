// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
)

type IResult interface {
	Code() int
}

type StatusResult struct {
	Code   int
	Result interface{}
	Type   int
}

const (
	AutoResponse = iota
	JsonResponse
	XmlResponse
)

type ResponseTyper interface {
	ResponseType() int
}

type Json struct{}

func (Json) ResponseType() int {
	return JsonResponse
}

type Xml struct{}

func (Xml) ResponseType() int {
	return XmlResponse
}

func isNil(a interface{}) bool {
	if a == nil {
		return true
	}
	aa := reflect.ValueOf(a)
	return !aa.IsValid() || (aa.Type().Kind() == reflect.Ptr && aa.IsNil())
}

type XmlError struct {
	XMLName xml.Name `xml:"err"`
	Content string   `xml:"content"`
}

type XmlString struct {
	XMLName xml.Name `xml:"string"`
	Content string   `xml:"content"`
}

func Return() HandlerFunc {
	return func(ctx *Context) {
		var rt int
		action := ctx.Action()
		if action != nil {
			if i, ok := action.(ResponseTyper); ok {
				rt = i.ResponseType()
			}
		}

		ctx.Next()

		// if no route match or has been write, then return
		if action == nil || ctx.Written() {
			return
		}

		// if there is no return value or return nil
		if isNil(ctx.Result) {
			// then we return blank page
			ctx.Result = ""
		}

		var result = ctx.Result
		var statusCode int = 0
		if res, ok := ctx.Result.(*StatusResult); ok {
			statusCode = res.Code
			result = res.Result
			rt = res.Type
		}

		if rt == JsonResponse {
			encoder := json.NewEncoder(ctx)
			switch res := result.(type) {
			case AbortError:
				if statusCode == 0 {
					statusCode = res.Code()
				}
				ctx.WriteHeader(statusCode)
				encoder.Encode(map[string]string{
					"err": res.Error(),
				})
			case error:
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				ctx.WriteHeader(statusCode)
				encoder.Encode(map[string]string{
					"err": res.Error(),
				})
			case string:
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				ctx.WriteHeader(statusCode)
				ctx.WriteString(res)
			case []byte:
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				ctx.WriteHeader(statusCode)
				ctx.WriteString(string(res))
			default:
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				ctx.WriteHeader(statusCode)
				if res == nil {
					return
				}
				err := encoder.Encode(res)
				if err != nil {
					ctx.Result = err
					encoder.Encode(map[string]string{
						"err": err.Error(),
					})
				}
			}

			return
		} else if rt == XmlResponse {
			encoder := xml.NewEncoder(ctx)
			switch res := result.(type) {
			case AbortError:
				if statusCode == 0 {
					statusCode = res.Code()
				}
				ctx.WriteHeader(statusCode)
				encoder.Encode(XmlError{
					Content: res.Error(),
				})
			case error:
				if statusCode == 0 {
					statusCode = http.StatusInternalServerError
				}
				ctx.WriteHeader(statusCode)
				encoder.Encode(XmlError{
					Content: res.Error(),
				})
			case string:
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				ctx.WriteHeader(statusCode)
				ctx.WriteString(res)
			case []byte:
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				ctx.WriteHeader(statusCode)
				ctx.Write(res)
			default:
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				ctx.WriteHeader(statusCode)
				if result == nil {
					return
				}
				err := encoder.Encode(result)
				if err != nil {
					ctx.Result = err
					encoder.Encode(XmlError{
						Content: err.Error(),
					})
				}
			}
			return
		}
		switch res := result.(type) {
		case AbortError, error:
			ctx.HandleError()
		case []byte:
			if statusCode == 0 {
				statusCode = http.StatusOK
			}
			ctx.WriteHeader(statusCode)
			ctx.Write(res)
		case string:
			if statusCode == 0 {
				statusCode = http.StatusOK
			}
			ctx.WriteHeader(statusCode)
			ctx.WriteString(res)
		default:
			if statusCode == 0 {
				statusCode = http.StatusOK
			}
			ctx.WriteHeader(statusCode)
			if res == nil {
				return
			}
			_, err := ctx.WriteString(fmt.Sprintf("%+v", res))
			if err != nil {
				ctx.Result = err
				ctx.WriteString(fmt.Sprintf("err:%v", err.Error()))
			}
		}
	}
}
