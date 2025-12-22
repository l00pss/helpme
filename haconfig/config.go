package haconfig

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration manager
type Config struct {
	data       interface{}
	envPrefix  string
	yamlFile   string
	envMapping map[string]string
}

// ConfigOption represents configuration options
type ConfigOption func(*Config)

// WithEnvPrefix sets the environment variable prefix
func WithEnvPrefix(prefix string) ConfigOption {
	return func(c *Config) {
		c.envPrefix = prefix
	}
}

// WithYAMLFile sets the YAML file path
func WithYAMLFile(file string) ConfigOption {
	return func(c *Config) {
		c.yamlFile = file
	}
}

// WithEnvMapping allows custom environment variable mappings
func WithEnvMapping(mapping map[string]string) ConfigOption {
	return func(c *Config) {
		c.envMapping = mapping
	}
}

// New creates a new configuration manager
func New(opts ...ConfigOption) *Config {
	config := &Config{
		envPrefix:  "",
		envMapping: make(map[string]string),
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// Load loads configuration into the provided struct
func (c *Config) Load(cfg interface{}) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to struct")
	}

	// First load from YAML file if specified
	if c.yamlFile != "" {
		if err := c.loadFromYAML(cfg); err != nil {
			return fmt.Errorf("failed to load YAML: %w", err)
		}
	}

	// Then override with environment variables
	if err := c.loadFromEnv(cfg); err != nil {
		return fmt.Errorf("failed to load from env: %w", err)
	}

	return nil
}

// loadFromYAML loads configuration from YAML file
func (c *Config) loadFromYAML(cfg interface{}) error {
	data, err := os.ReadFile(c.yamlFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, skip YAML loading
			return nil
		}
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv(cfg interface{}) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	return c.processStruct(v, t, "")
}

// processStruct processes a struct and its nested fields
func (c *Config) processStruct(v reflect.Value, t reflect.Type, prefix string) error {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get field name and environment variable name
		fieldName := fieldType.Name

		// Handle nested structs
		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			newPrefix := c.buildPrefix(prefix, fieldName)
			if err := c.processStruct(field, fieldType.Type, newPrefix); err != nil {
				return err
			}
			continue
		}

		// Handle pointers to structs
		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			// Check if any env var exists for this nested struct before creating it
			newPrefix := c.buildPrefix(prefix, fieldName)
			if c.hasAnyEnvVar(field.Type().Elem(), newPrefix) {
				if field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
				}
				if err := c.processStruct(field.Elem(), field.Type().Elem(), newPrefix); err != nil {
					return err
				}
			}
			continue
		}

		envName := c.getEnvName(fieldName, prefix)

		// Check for custom mapping
		if customName, exists := c.envMapping[fieldName]; exists {
			envName = customName
		}

		// Set value from environment variable
		if err := c.setFieldFromEnv(field, envName); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldName, err)
		}
	}

	return nil
}

// hasAnyEnvVar checks if any environment variable exists for a struct type
func (c *Config) hasAnyEnvVar(structType reflect.Type, prefix string) bool {
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldName := fieldType.Name

		// Check nested structs recursively
		if fieldType.Type.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			newPrefix := c.buildPrefix(prefix, fieldName)
			if c.hasAnyEnvVar(fieldType.Type, newPrefix) {
				return true
			}
		} else if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
			newPrefix := c.buildPrefix(prefix, fieldName)
			if c.hasAnyEnvVar(fieldType.Type.Elem(), newPrefix) {
				return true
			}
		} else {
			envName := c.getEnvName(fieldName, prefix)
			if customName, exists := c.envMapping[fieldName]; exists {
				envName = customName
			}
			if os.Getenv(envName) != "" {
				return true
			}
		}
	}
	return false
}

// getEnvName generates environment variable name
func (c *Config) getEnvName(fieldName, prefix string) string {
	envName := c.toSnakeCase(fieldName)

	if prefix != "" {
		envName = prefix + "_" + envName
	}

	if c.envPrefix != "" {
		envName = c.envPrefix + "_" + envName
	}

	return strings.ToUpper(envName)
}

// buildPrefix builds prefix for nested structs
func (c *Config) buildPrefix(currentPrefix, fieldName string) string {
	snakeName := c.toSnakeCase(fieldName)
	if currentPrefix == "" {
		return snakeName
	}
	return currentPrefix + "_" + snakeName
}

// toSnakeCase converts CamelCase to snake_case using regex
func (c *Config) toSnakeCase(str string) string {
	if str == "" {
		return ""
	}

	// First, handle the transition from lowercase to uppercase
	re1 := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	result := re1.ReplaceAllString(str, `${1}_${2}`)

	// Then handle the transition from multiple uppercase to lowercase
	re2 := regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	result = re2.ReplaceAllString(result, `${1}_${2}`)

	return strings.ToLower(result)
}

// setFieldFromEnv sets field value from environment variable
func (c *Config) setFieldFromEnv(field reflect.Value, envName string) error {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return nil // No environment variable set
	}

	return c.setFieldValue(field, envValue)
}

// setFieldValue sets field value based on its type
func (c *Config) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value: %s", value)
		}
		field.SetBool(boolVal)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration value: %s", value)
			}
			field.Set(reflect.ValueOf(duration))
			return nil
		}

		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int value: %s", value)
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint value: %s", value)
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(floatVal)

	case reflect.Slice:
		return c.setSliceValue(field, value)

	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return c.setFieldValue(field.Elem(), value)

	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			timeVal, err := time.Parse(time.RFC3339, value)
			if err != nil {
				// Try different time formats
				formats := []string{
					"2006-01-02 15:04:05",
					"2006-01-02",
					"15:04:05",
				}
				for _, format := range formats {
					if timeVal, err = time.Parse(format, value); err == nil {
						break
					}
				}
				if err != nil {
					return fmt.Errorf("invalid time value: %s", value)
				}
			}
			field.Set(reflect.ValueOf(timeVal))
		}

	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}

// setSliceValue sets slice value from comma-separated string
func (c *Config) setSliceValue(field reflect.Value, value string) error {
	parts := strings.Split(value, ",")
	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		elem := slice.Index(i)
		if err := c.setFieldValue(elem, part); err != nil {
			return fmt.Errorf("invalid slice element at index %d: %w", i, err)
		}
	}

	field.Set(slice)
	return nil
}

// LoadFromFile loads configuration from a specific YAML file
func LoadFromFile(file string, cfg interface{}, opts ...ConfigOption) error {
	options := append([]ConfigOption{WithYAMLFile(file)}, opts...)
	config := New(options...)
	return config.Load(cfg)
}

// LoadFromEnv loads configuration only from environment variables
func LoadFromEnv(cfg interface{}, opts ...ConfigOption) error {
	config := New(opts...)
	return config.loadFromEnv(cfg)
}

// MustLoad loads configuration and panics on error
func (c *Config) MustLoad(cfg interface{}) {
	if err := c.Load(cfg); err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
}

// Validate validates the configuration
func (c *Config) Validate(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return c.validateStruct(v, v.Type(), "")
}

// validateStruct validates struct fields
func (c *Config) validateStruct(v reflect.Value, t reflect.Type, prefix string) error {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Check required tag
		if tag := fieldType.Tag.Get("required"); tag == "true" {
			if c.isZeroValue(field) {
				fieldPath := c.buildFieldPath(prefix, fieldType.Name)
				return fmt.Errorf("required field is empty: %s", fieldPath)
			}
		}

		// Validate nested structs
		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			newPrefix := c.buildFieldPath(prefix, fieldType.Name)
			if err := c.validateStruct(field, fieldType.Type, newPrefix); err != nil {
				return err
			}
		}

		// Validate pointer to structs
		if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			newPrefix := c.buildFieldPath(prefix, fieldType.Name)
			if err := c.validateStruct(field.Elem(), field.Type().Elem(), newPrefix); err != nil {
				return err
			}
		}
	}

	return nil
}

// isZeroValue checks if field has zero value
func (c *Config) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.Struct:
		return v.Interface() == reflect.Zero(v.Type()).Interface()
	default:
		return false
	}
}

// buildFieldPath builds field path for validation errors
func (c *Config) buildFieldPath(prefix, fieldName string) string {
	if prefix == "" {
		return fieldName
	}
	return prefix + "." + fieldName
}
