package kafkalib

type Config struct {
	ClientID string `env:"CLIENT_ID" json:"client_id" yaml:"client_id"`
	Brokers  string `env:"BROKERS" json:"brokers" yaml:"brokers"`

	DebugLog bool `env:"DEBUG_LOG" envDefault:"false" json:"debug_log" yaml:"debug_log" default:"false"`

	TLSEnabled             bool   `env:"TLS_ENABLED" json:"tls_enabled" yaml:"tls_enabled"`
	ClientKeyFilePath      string `env:"CLIENT_KEY_FILE_PATH" json:"client_key_file_path" yaml:"client_key_file_path"`
	ClientCertFilePath     string `env:"CLIENT_CERT_FILE_PATH" json:"client_cert_file_path" yaml:"client_cert_file_path"`
	CaFilePath             string `env:"CA_FILE_PATH" json:"ca_file_path" yaml:"ca_file_path"`
	EnableCertVerification bool   `env:"ENABLE_CERT_VERIFICATION" json:"enable_cert_verification" yaml:"enable_cert_verification"`
	SaslUser               string `env:"SAASL_USER" json:"sasl_user" yaml:"sasl_user"`
	SaslPassword           string `env:"SAASL_PASSWORD" json:"sasl_password" yaml:"sasl_password"`
}
