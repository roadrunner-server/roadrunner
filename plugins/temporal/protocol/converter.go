package protocol

import (
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

type (
	// DataConverter wraps Temporal data converter to enable direct access to the payloads.
	DataConverter struct {
		fallback converter.DataConverter
	}
)

// NewDataConverter creates new data converter.
func NewDataConverter(fallback converter.DataConverter) converter.DataConverter {
	return &DataConverter{fallback: fallback}
}

// ToPayloads converts a list of values.
func (r *DataConverter) ToPayloads(values ...interface{}) (*commonpb.Payloads, error) {
	for _, v := range values {
		if aggregated, ok := v.(*commonpb.Payloads); ok {
			// bypassing
			return aggregated, nil
		}
	}

	return r.fallback.ToPayloads(values...)
}

// ToPayload converts single value to payload.
func (r *DataConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	return r.fallback.ToPayload(value)
}

// FromPayloads converts to a list of values of different types.
// Useful for deserializing arguments of function invocations.
func (r *DataConverter) FromPayloads(payloads *commonpb.Payloads, valuePtrs ...interface{}) error {
	if payloads == nil {
		return nil
	}

	if len(valuePtrs) == 1 {
		// input proxying
		if input, ok := valuePtrs[0].(**commonpb.Payloads); ok {
			*input = &commonpb.Payloads{}
			(*input).Payloads = payloads.Payloads
			return nil
		}
	}

	for i := 0; i < len(payloads.Payloads); i++ {
		err := r.FromPayload(payloads.Payloads[i], valuePtrs[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// FromPayload converts single value from payload.
func (r *DataConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	return r.fallback.FromPayload(payload, valuePtr)
}

// ToString converts payload object into human readable string.
func (r *DataConverter) ToString(input *commonpb.Payload) string {
	return r.fallback.ToString(input)
}

// ToStrings converts payloads object into human readable strings.
func (r *DataConverter) ToStrings(input *commonpb.Payloads) []string {
	return r.fallback.ToStrings(input)
}
