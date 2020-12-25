package config

type Configurer interface {
	// UnmarshalKey reads configuration section into configuration object.
	//
	// func (h *HttpService) Init(cp config.Configurer) error {
	//     h.config := &HttpConfig{}
	//     if err := configProvider.UnmarshalKey("http", h.config); err != nil {
	//         return err
	//     }
	// }
	UnmarshalKey(name string, out interface{}) error

	// Get used to get config section
	Get(name string) interface{}

	// Overwrite used to overwrite particular values in the unmarshalled config
	Overwrite(values map[string]interface{}) error

	// Has checks if config section exists.
	Has(name string) bool
}
