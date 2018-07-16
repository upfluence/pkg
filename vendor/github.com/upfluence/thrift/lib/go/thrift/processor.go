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

// A processor is a generic object which operates upon an input stream and
// writes to some output stream.
type TProcessor interface {
	Process(ctx Context, in, out TProtocol) (bool, TException)
	GetMiddlewares() []TMiddleware
	AddProcessor(string, TProcessorFunction)
}

type TStandardProcessor struct {
	ProcessorMap map[string]TProcessorFunction
	Middlewares  []TMiddleware
}

func NewTStandardProcessor(ms []TMiddleware) *TStandardProcessor {
	return &TStandardProcessor{
		Middlewares:  ms,
		ProcessorMap: make(map[string]TProcessorFunction),
	}
}

func (p *TStandardProcessor) GetMiddlewares() []TMiddleware {
	return p.Middlewares
}

func (p *TStandardProcessor) AddProcessor(fname string, fn TProcessorFunction) {
	p.ProcessorMap[fname] = fn
}

func (p *TStandardProcessor) Process(ctx Context, in, out TProtocol) (bool, TException) {
	name, _, seqID, err := in.ReadMessageBegin()

	if err != nil {
		return false, err
	}

	if processor, ok := p.ProcessorMap[name]; ok {
		return processor.Process(ctx, seqID, in, out)
	}

	in.Skip(STRUCT)
	in.ReadMessageEnd()
	x5 := NewTApplicationException(UNKNOWN_METHOD, "Unknown function "+name)
	out.WriteMessageBegin(name, EXCEPTION, seqID)
	x5.Write(out)
	out.WriteMessageEnd()
	out.Flush()
	return false, x5
}

type TProcessorFunction interface {
	Process(ctx Context, seqID int32, in, out TProtocol) (bool, TException)
}

type TBaseProcessorFunction struct {
	fname       string
	argBuilder  func() TRequest
	middlewares []TMiddleware
}

func NewTBaseProcessorFunction(p TProcessor, fname string, builder func() TRequest) *TBaseProcessorFunction {
	return &TBaseProcessorFunction{
		fname:       fname,
		argBuilder:  builder,
		middlewares: p.GetMiddlewares(),
	}
}

func (p *TBaseProcessorFunction) readRequest(in TProtocol) (TRequest, error) {
	args := p.argBuilder()

	if err := args.Read(in); err != nil {
		in.ReadMessageEnd()
		return nil, err
	}

	in.ReadMessageEnd()

	return args, nil
}

type TBinaryHandler interface {
	Handle(Context, TRequest) (TResponse, error)
}

type TBinaryProcessorFunction struct {
	*TBaseProcessorFunction
	handler TBinaryHandler
}

func NewTBinaryProcessorFunction(p TProcessor, fname string, builder func() TRequest, handler TBinaryHandler) *TBinaryProcessorFunction {
	return &TBinaryProcessorFunction{
		TBaseProcessorFunction: NewTBaseProcessorFunction(p, fname, builder),
		handler:                handler,
	}
}

func (p *TBinaryProcessorFunction) Process(ctx Context, seqID int32, in, out TProtocol) (bool, TException) {
	var args, err = p.readRequest(in)

	if err != nil {
		x := NewTApplicationException(PROTOCOL_ERROR, err.Error())
		out.WriteMessageBegin(p.fname, EXCEPTION, seqID)
		x.Write(out)
		out.WriteMessageEnd()
		out.Flush()
		return false, err
	}

	call := func(ctx Context, req TRequest) (TResponse, error) {
		return p.handler.Handle(ctx, req)
	}

	for i := len(p.middlewares); i > 0; i-- {
		j := i
		next := call
		call = func(ctx Context, req TRequest) (TResponse, error) {
			return p.middlewares[j-1].HandleBinaryRequest(
				ctx,
				p.fname,
				seqID,
				req,
				next,
			)
		}
	}

	res, err := call(ctx, args)

	if err != nil {
		x := NewTApplicationException(INTERNAL_ERROR, "Internal error processing perform_void: "+err.Error())
		out.WriteMessageBegin(p.fname, EXCEPTION, seqID)
		x.Write(out)
		out.WriteMessageEnd()
		out.Flush()
		return true, err
	}

	if err2 := out.WriteMessageBegin(p.fname, REPLY, seqID); err2 != nil {
		err = err2
	}

	if err2 := res.Write(out); err == nil && err2 != nil {
		err = err2
	}

	if err2 := out.WriteMessageEnd(); err == nil && err2 != nil {
		err = err2
	}

	if err2 := out.Flush(); err == nil && err2 != nil {
		err = err2
	}

	return true, err
}

type TUnaryHandler interface {
	Handle(Context, TRequest) error
}

type TUnaryProcessorFunction struct {
	*TBaseProcessorFunction
	handler TUnaryHandler
}

func NewTUnaryProcessorFunction(p TProcessor, fname string, builder func() TRequest, handler TUnaryHandler) *TUnaryProcessorFunction {
	return &TUnaryProcessorFunction{
		TBaseProcessorFunction: NewTBaseProcessorFunction(p, fname, builder),
		handler:                handler,
	}
}

func (p *TUnaryProcessorFunction) Process(ctx Context, seqID int32, in, out TProtocol) (bool, TException) {
	var args, err = p.readRequest(in)

	if err != nil {
		return false, err
	}

	call := func(ctx Context, req TRequest) error {
		return p.handler.Handle(ctx, req)
	}

	for i := len(p.middlewares); i > 0; i-- {
		j := i
		next := call
		call = func(ctx Context, req TRequest) error {
			return p.middlewares[j-1].HandleUnaryRequest(
				ctx,
				p.fname,
				seqID,
				req,
				next,
			)
		}
	}

	return true, call(ctx, args)
}
