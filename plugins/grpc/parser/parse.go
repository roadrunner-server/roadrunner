package parser

import (
	"bytes"
	"io"
	"os"

	pp "github.com/emicklei/proto"
)

// Service contains information about singular GRPC service.
type Service struct {
	// Package defines service namespace.
	Package string

	// Name defines service name.
	Name string

	// Methods list.
	Methods []Method
}

// Method describes singular RPC method.
type Method struct {
	// Name is method name.
	Name string

	// StreamsRequest defines if method accept stream input.
	StreamsRequest bool

	// RequestType defines message name (from the same package) of method input.
	RequestType string

	// StreamsReturns defines if method streams result.
	StreamsReturns bool

	// ReturnsType defines message name (from the same package) of method return value.
	ReturnsType string
}

// File parses given proto file or returns error.
func File(file string, importPath string) ([]Service, error) {
	reader, _ := os.Open(file)
	defer reader.Close()

	return parse(reader, importPath)
}

// Bytes parses string into proto definition.
func Bytes(data []byte) ([]Service, error) {
	return parse(bytes.NewBuffer(data), "")
}

func parse(reader io.Reader, importPath string) ([]Service, error) {
	proto, err := pp.NewParser(reader).Parse()
	if err != nil {
		return nil, err
	}

	return parseServices(
		proto,
		parsePackage(proto),
		importPath,
	)
}

func parsePackage(proto *pp.Proto) string {
	for _, e := range proto.Elements {
		if p, ok := e.(*pp.Package); ok {
			return p.Name
		}
	}

	return ""
}

func parseServices(proto *pp.Proto, pkg string, importPath string) ([]Service, error) {
	services := make([]Service, 0)

	pp.Walk(proto, pp.WithService(func(service *pp.Service) {
		services = append(services, Service{
			Package: pkg,
			Name:    service.Name,
			Methods: parseMethods(service),
		})
	}))

	pp.Walk(proto, func(v pp.Visitee) {
		if i, ok := v.(*pp.Import); ok {
			if im, err := File(importPath+"/"+i.Filename, importPath); err == nil {
				services = append(services, im...)
			}
		}
	})

	return services, nil
}

func parseMethods(s *pp.Service) []Method {
	methods := make([]Method, 0)
	for _, e := range s.Elements {
		if m, ok := e.(*pp.RPC); ok {
			methods = append(methods, Method{
				Name:           m.Name,
				StreamsRequest: m.StreamsRequest,
				RequestType:    m.RequestType,
				StreamsReturns: m.StreamsReturns,
				ReturnsType:    m.ReturnsType,
			})
		}
	}

	return methods
}
