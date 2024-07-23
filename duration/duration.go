package duration

import "time"

type ConfigDuration struct {
	time.Duration
}

func (cd *ConfigDuration) Unmarshal(v []byte) error {
	d, err := time.ParseDuration(string(v))
	if err != nil {
		return err
	}

	cd.Duration = d
	return nil
}

func (cd *ConfigDuration) Marshal() ([]byte, error) {
	return []byte(cd.Duration.String()), nil
}

func (cd *ConfigDuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(&cd.Duration)
}

func (cd *ConfigDuration) MarshalYAML() (interface{}, error) {
	return cd.Duration.String(), nil
}

func (cd *ConfigDuration) UnmarshalText(v []byte) error {
	d, err := time.ParseDuration(string(v))
	if err != nil {
		return err
	}

	cd.Duration = d
	return nil
}
