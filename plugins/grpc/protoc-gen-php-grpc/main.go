// MIT License
//
// Copyright (c) 2018 SpiralScout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/spiral/roadrunner/v2/plugins/grpc/protoc-gen-php-grpc/php"
	"google.golang.org/protobuf/proto"
	plugin "google.golang.org/protobuf/types/pluginpb"
)

func main() {
	req, err := readRequest(os.Stdin)
	if err != nil {
		panic(err)
	}

	if err = writeResponse(os.Stdout, php.Generate(req)); err != nil {
		panic(err)
	}
}

func readRequest(in io.Reader) (*plugin.CodeGeneratorRequest, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(data, req); err != nil {
		return nil, err
	}

	return req, nil
}

func writeResponse(out io.Writer, resp *plugin.CodeGeneratorResponse) error {
	data, err := proto.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = out.Write(data)
	return err
}
