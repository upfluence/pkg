#!/bin/bash

thrift --gen go:thrift_import=github.com/upfluence/thrift/lib/go/thrift,package_prefix=github.com/upfluence/thrift-amqp-go/integration/gen-go/ test.thrift
