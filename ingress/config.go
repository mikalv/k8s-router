package ingress

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/30x/k8s-pods-ingress/utils"

	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/validation"
)

const (
	// DefaultAPIKeySecret is the default value for the first portion of the DefaultAPIKeySecretLocation (routing)
	DefaultAPIKeySecret = "routing"
	// DefaultAPIKeySecretDataField is the default value for the second portion of the DefaultAPIKeySecretDataField (api-key)
	DefaultAPIKeySecretDataField = "api-key"
	// DefaultAPIKeySecretLocation is the default value for the EnvVarAPIKeySecretLocation (routing:api-key)
	DefaultAPIKeySecretLocation = DefaultAPIKeySecret + ":" + DefaultAPIKeySecretDataField
	// DefaultHostsAnnotation is the default value for EnvVarHostsAnnotation (routingHosts)
	DefaultHostsAnnotation = "routingHosts"
	// DefaultPathsAnnotation is the default value for the EnvVarHostsAnnotation (routingPaths)
	DefaultPathsAnnotation = "routingPaths"
	// DefaultPort is the default value for the EnvVarPort (80)
	DefaultPort = 80
	// DefaultRoutableLabelSelector is the default value for EnvVarRoutableLabelSelector (routable=true)
	DefaultRoutableLabelSelector = "routable=true"
	// EnvVarAPIKeySecretLocation Environment variable name for providing the location of the secret (name:field) to identify API Key secrets
	EnvVarAPIKeySecretLocation = "API_KEY_SECRET_LOCATION"
	// EnvVarHostsAnnotation Environment variable name for providing the name of the hosts annotation
	EnvVarHostsAnnotation = "HOSTS_ANNOTATION"
	// EnvVarPathsAnnotation Environment variable name for providing the the name of the paths annotation
	EnvVarPathsAnnotation = "PATHS_ANNOTATION"
	// EnvVarPort Environment variable for providing the port nginx should listen on
	EnvVarPort = "PORT"
	// EnvVarRoutableLabelSelector Environment variable name for providing the label selector for identifying routable objects
	EnvVarRoutableLabelSelector = "ROUTABLE_LABEL_SELECTOR"
	// ErrMsgTmplInvalidAnnotationName is the error message template for an invalid annotation name
	ErrMsgTmplInvalidAnnotationName = "%s has an invalid annotation name: %s"
	// ErrMsgTmplInvalidAPIKeySecretLocation is the error message template for invalid API Key Secret location environment variable values
	ErrMsgTmplInvalidAPIKeySecretLocation = "%s is not in the format of {API_KEY_SECRET_NAME}:{API_KEY_SECRET_DATA_FIELD_NAME}"
	// ErrMsgTmplInvalidLabelSelector is the error message template for an invalid label selector
	ErrMsgTmplInvalidLabelSelector = "%s has an invalid label selector: %s\n"
	// ErrMsgTmplInvalidPort is the error message template for an invalid port
	ErrMsgTmplInvalidPort = "%s is an invalid port: %s\n"
)

/*
ConfigFromEnv returns the configuration based on the environment variables and validates the values
*/
func ConfigFromEnv() (*Config, error) {
	config := &Config{
		HostsAnnotation: os.Getenv(EnvVarHostsAnnotation),
		PathsAnnotation: os.Getenv(EnvVarPathsAnnotation),
	}

	// Apply defaults
	if config.HostsAnnotation == "" {
		config.HostsAnnotation = DefaultHostsAnnotation
	}

	if config.PathsAnnotation == "" {
		config.PathsAnnotation = DefaultPathsAnnotation
	}

	// Validate configuration
	apiKeySecretLocation := os.Getenv(EnvVarAPIKeySecretLocation)
	var apiKeySecretLocationParts []string

	if apiKeySecretLocation == "" {
		// No need to validate, just use the default
		config.APIKeySecret = DefaultAPIKeySecret
		config.APIKeySecretDataField = DefaultAPIKeySecretDataField
	} else {
		apiKeySecretLocationParts = strings.Split(apiKeySecretLocation, ":")

		if len(apiKeySecretLocationParts) == 2 {
			config.APIKeySecret = apiKeySecretLocationParts[0]
			config.APIKeySecretDataField = apiKeySecretLocationParts[1]
		} else {
			return nil, fmt.Errorf(ErrMsgTmplInvalidAPIKeySecretLocation, EnvVarAPIKeySecretLocation)
		}
	}

	if !validation.IsQualifiedName(strings.ToLower(config.HostsAnnotation)) {
		return nil, fmt.Errorf(ErrMsgTmplInvalidAnnotationName, EnvVarHostsAnnotation, config.HostsAnnotation)
	} else if !validation.IsQualifiedName(strings.ToLower(config.PathsAnnotation)) {
		return nil, fmt.Errorf(ErrMsgTmplInvalidAnnotationName, EnvVarPathsAnnotation, config.PathsAnnotation)
	}

	portStr := os.Getenv(EnvVarPort)

	if portStr == "" {
		config.Port = DefaultPort
	} else {
		port, err := strconv.Atoi(portStr)

		if err != nil || !utils.IsValidPort(port) {
			return nil, fmt.Errorf(ErrMsgTmplInvalidPort, EnvVarPort, portStr)
		}

		config.Port = port
	}

	routableLabelSelector := os.Getenv(EnvVarRoutableLabelSelector)

	if routableLabelSelector == "" {
		routableLabelSelector = DefaultRoutableLabelSelector
	}

	selector, err := labels.Parse(routableLabelSelector)

	if err == nil {
		config.RoutableLabelSelector = selector
	} else {
		return nil, fmt.Errorf(ErrMsgTmplInvalidLabelSelector, EnvVarRoutableLabelSelector, routableLabelSelector)
	}

	return config, nil
}