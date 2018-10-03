// Autogenerated by Thrift Compiler (2.0.1-upfluence)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package http

import (
	"bytes"
	"fmt"
	"github.com/upfluence/thrift/lib/go/thrift"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = bytes.Equal

type HTTPProxy interface {
	// Parameters:
	//  - Req
	Perform(ctx thrift.Context, req *Request) (res *Response, err error)
}

type HTTPProxyClient struct {
	thrift.TClient
}

func NewHTTPProxyClientFactoryProvider(p thrift.TClientProvider) (*HTTPProxyClient, error) {
	cl, err := p.Build("proxy.http", "HTTPProxy")
	if err != nil {
		return nil, err
	}

	return &HTTPProxyClient{TClient: cl}, nil
}

// Parameters:
//  - Req
func (p *HTTPProxyClient) Perform(ctx thrift.Context, req *Request) (res *Response, err error) {
	args := HTTPProxyPerformArgs{
		Req: req,
	}
	result := HTTPProxyPerformResult{}
	if err = p.CallBinary(ctx, "perform", &args, &result); err != nil {
		return res, err
	}

	return result.GetSuccess(), nil
}

func NewHTTPProxyProcessorProvider(handler HTTPProxy, provider thrift.TProcessorProvider) (thrift.TProcessor, error) {
	p, err := provider.Build("proxy.http", "HTTPProxy")
	if err != nil {
		return nil, err
	}

	return NewHTTPProxyProcessorFactory(handler, p), nil
}

func NewHTTPProxyProcessor(handler HTTPProxy, middlewares []thrift.TMiddleware) thrift.TProcessor {
	p := thrift.NewTStandardProcessor(middlewares)
	return NewHTTPProxyProcessorFactory(handler, p)
}

func NewHTTPProxyProcessorFactory(handler HTTPProxy, p thrift.TProcessor) thrift.TProcessor {
	p.AddProcessor(
		"perform",
		thrift.NewTBinaryProcessorFunction(p, "perform", func() thrift.TRequest { return &HTTPProxyPerformArgs{} }, &hTTPProxyProcessorPerform{handler: handler}),
	)
	return p
}

type hTTPProxyProcessorPerform struct {
	handler HTTPProxy
}

func (p *hTTPProxyProcessorPerform) Handle(ctx thrift.Context, req thrift.TRequest) (thrift.TResponse, error) {
	args := req.(*HTTPProxyPerformArgs)
	retval, err2 := p.handler.Perform(ctx, args.Req)
	result := &HTTPProxyPerformResult{}
	if err2 != nil {
		return nil, err2
	}

	result.Success = retval
	return result, nil
}

// HELPER FUNCTIONS AND STRUCTURES

// Attributes:
//  - Req
type HTTPProxyPerformArgs struct {
	Req *Request `thrift:"req,1" json:"req"`
}

func NewHTTPProxyPerformArgs() *HTTPProxyPerformArgs {
	return &HTTPProxyPerformArgs{}
}

var HTTPProxyPerformArgs_Req_DEFAULT *Request

func (p *HTTPProxyPerformArgs) GetReq() *Request {
	if !p.IsSetReq() {
		return HTTPProxyPerformArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *HTTPProxyPerformArgs) SetReq(v *Request) {
	p.Req = v
}
func (p *HTTPProxyPerformArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *HTTPProxyPerformArgs) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *HTTPProxyPerformArgs) ReadField1(iprot thrift.TProtocol) error {
	p.Req = NewRequest()
	if err := p.Req.Read(iprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
	}
	return nil
}

func (p *HTTPProxyPerformArgs) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("perform_args"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *HTTPProxyPerformArgs) writeField1(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("req", thrift.STRUCT, 1); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err)
	}
	if err := p.Req.Write(oprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err)
	}
	return err
}

func (p *HTTPProxyPerformArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("HTTPProxyPerformArgs(%+v)", *p)
}

func (p *HTTPProxyPerformResult) GetResult() interface{} {
	return p.GetSuccess()
}

func (p *HTTPProxyPerformResult) GetError() error {
	return nil

}

// Attributes:
//  - Success
type HTTPProxyPerformResult struct {
	Success *Response `thrift:"success,0" json:"success,omitempty"`
}

func NewHTTPProxyPerformResult() *HTTPProxyPerformResult {
	return &HTTPProxyPerformResult{}
}

var HTTPProxyPerformResult_Success_DEFAULT *Response

func (p *HTTPProxyPerformResult) GetSuccess() *Response {
	if !p.IsSetSuccess() {
		return HTTPProxyPerformResult_Success_DEFAULT
	}
	return p.Success
}

func (p *HTTPProxyPerformResult) SetSuccess(v *Response) {
	p.Success = v
}
func (p *HTTPProxyPerformResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *HTTPProxyPerformResult) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 0:
			if err := p.ReadField0(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *HTTPProxyPerformResult) ReadField0(iprot thrift.TProtocol) error {
	p.Success = NewResponse()
	if err := p.Success.Read(iprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
	}
	return nil
}

func (p *HTTPProxyPerformResult) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("perform_result"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField0(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *HTTPProxyPerformResult) writeField0(oprot thrift.TProtocol) (err error) {
	if p.IsSetSuccess() {
		if err := oprot.WriteFieldBegin("success", thrift.STRUCT, 0); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err)
		}
		if err := p.Success.Write(oprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err)
		}
	}
	return err
}

func (p *HTTPProxyPerformResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("HTTPProxyPerformResult(%+v)", *p)
}
