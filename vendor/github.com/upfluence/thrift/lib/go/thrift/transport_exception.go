/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package thrift

import (
	"errors"
	"io"
)

type timeoutable interface {
	Timeout() bool
}

// Thrift Transport exception
type TTransportException interface {
	TException
	TypeId() int
	Err() error
}

const (
	UNKNOWN_TRANSPORT_EXCEPTION = 0
	NOT_OPEN                    = 1
	ALREADY_OPEN                = 2
	TIMED_OUT                   = 3
	END_OF_FILE                 = 4
)

type tTransportException struct {
	typeId int
	err    error
}

func (te *tTransportException) Timeout() bool {
	return te.typeId == TIMED_OUT
}

func (te *tTransportException) TypeId() int {
	return te.typeId
}

func (te *tTransportException) Error() string {
	return te.err.Error()
}

func (te *tTransportException) Err() error {
	return te.err
}

func NewTTransportException(t int, e string) TTransportException {
	return &tTransportException{typeId: t, err: errors.New(e)}
}

func NewTTransportExceptionFromError(err error) TTransportException {
	var cause = Cause(err)

	switch cause {
	case nil:
		return nil
	case io.EOF:
		return &tTransportException{typeId: END_OF_FILE, err: err}
	}

	switch v := cause.(type) {
	case TTransportException:
		return v
	case timeoutable:
		if v.Timeout() {
			return &tTransportException{typeId: TIMED_OUT, err: err}
		}
	}

	return &tTransportException{typeId: UNKNOWN_TRANSPORT_EXCEPTION, err: err}
}
