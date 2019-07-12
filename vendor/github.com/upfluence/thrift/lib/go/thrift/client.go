package thrift

import (
	"fmt"
	"sync"
)

type TClientFactory interface {
	GetClient(TTransportFactory, TProtocolFactory, []TMiddleware) TClient
}

type tSyncClientFactory struct{}

func NewTDefaultClientFactory() TClientFactory {
	return &tSyncClientFactory{}
}

func (*tSyncClientFactory) GetClient(t TTransportFactory, p TProtocolFactory, ms []TMiddleware) TClient {
	return NewTSyncClient(t.GetTransport(nil), p, ms...)
}

type TClient interface {
	CallBinary(Context, string, TRequest, TResponse) error
	CallUnary(Context, string, TRequest) error
}

type TSyncClient struct {
	TTransport
	sync.Mutex

	in, out TProtocol

	seqID int32

	Middleware TMiddleware
}

func NewTSyncClient(t TTransport, f TProtocolFactory, ms ...TMiddleware) *TSyncClient {
	return &TSyncClient{
		TTransport: t,
		in:         f.GetProtocol(t),
		out:        f.GetProtocol(t),
		Middleware: WrapMiddlewares(ms),
	}
}

func send(ctx Context, oprot TProtocol, seqID int32, method string, args TRequest, mType TMessageType) error {
	if err := oprot.WriteMessageBegin(method, mType, seqID); err != nil {
		return err
	}

	if err := args.Write(oprot); err != nil {
		return err
	}

	if err := oprot.WriteMessageEnd(); err != nil {
		return err
	}

	if err := oprot.Transport().WriteContext(ctx); err != nil {
		return err
	}

	return oprot.Flush()
}

func recv(iprot TProtocol, seqID int32, method string, result TResponse) error {
	var rMethod, rTypeID, rSeqID, err = iprot.ReadMessageBegin()

	if err != nil {
		return err
	}

	if method != rMethod {
		return NewTApplicationException(
			WRONG_METHOD_NAME,
			fmt.Sprintf("%s: wrong method name", method),
		)
	} else if seqID != rSeqID {
		return NewTApplicationException(
			BAD_SEQUENCE_ID,
			fmt.Sprintf("%s: out of order sequence response", method),
		)
	} else if rTypeID == EXCEPTION {
		var (
			exception   tApplicationException
			retErr, err = exception.Read(iprot)
		)

		if err != nil {
			return err
		}

		if err := iprot.ReadMessageEnd(); err != nil {
			return err
		}

		return retErr
	} else if rTypeID != REPLY {
		return NewTApplicationException(
			INVALID_MESSAGE_TYPE_EXCEPTION,
			fmt.Sprintf("%s: invalid message type", method),
		)
	}

	if err := result.Read(iprot); err != nil {
		return err
	}

	return iprot.ReadMessageEnd()
}

func (c *TSyncClient) CallBinary(ctx Context, method string, req TRequest, res TResponse) error {
	c.Lock()
	defer c.Unlock()

	c.seqID++

	_, err := c.Middleware.HandleBinaryRequest(
		ctx,
		method,
		c.seqID,
		req,
		func(ctx Context, req TRequest) (TResponse, error) {
			if err := send(ctx, c.in, c.seqID, method, req, CALL); err != nil {
				return nil, err
			}

			return res, recv(c.out, c.seqID, method, res)
		},
	)

	return err
}

func (c *TSyncClient) CallUnary(ctx Context, method string, req TRequest) error {
	c.Lock()
	defer c.Unlock()

	c.seqID++

	return c.Middleware.HandleUnaryRequest(
		ctx,
		method,
		c.seqID,
		req,
		func(ctx Context, req TRequest) error {
			return send(ctx, c.in, c.seqID, method, req, ONEWAY)
		},
	)
}
