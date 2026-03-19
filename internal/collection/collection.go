package collection

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

var (
	placeholderRegexp = regexp.MustCompile(`\{([^}]+)\}`)
)

type Vars map[string]string
type RequestQuery map[string]any
type RequestHeaders map[string]string
type RequestMethod string

const (
	RequestMethodGet    = http.MethodGet
	RequestMethodPost   = http.MethodPost
	RequestMethodPut    = http.MethodPut
	RequestMethodPatch  = http.MethodPatch
	RequestMethodDelete = http.MethodDelete
)

func (m RequestMethod) validate() error {
	valid := []RequestMethod{RequestMethodGet, RequestMethodPost,
		RequestMethodPut, RequestMethodPatch, RequestMethodDelete}

	if !slices.Contains(valid, m) {
		return fmt.Errorf("invalid request method: %s", m)
	}

	return nil
}

type RequestEntry struct {
	Name    string         `toml:"name"`
	Url     string         `toml:"url"`
	Method  RequestMethod  `toml:"method"`
	Query   RequestQuery   `toml:"query"`   // optional.
	Headers RequestHeaders `toml:"headers"` // optional.
	Body    any            `toml:"body"`    // optional.
}

func (r *RequestEntry) validateAndParseUrl(vars Vars) error {
	if r.Url == "" {
		return fmt.Errorf("request url is missing")
	}

	var err error
	r.Url, err = replacePlaceholders(vars, r.Url, "request url")
	return err
}

func (r *RequestEntry) validateAndParseQuery(vars Vars) error {
	if r.Query == nil {
		return nil
	}

	var err error
	for k, v := range r.Query {
		switch vt := v.(type) {
		case int:
		case int64:
		case float64:
			return nil

		case string:
			r.Query[k], err = replacePlaceholders(vars, vt, "request query")
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("only int, float and string types are supported for request query values")
		}
	}

	return nil
}

func (r *RequestEntry) validateAndParseHeaders(vars Vars) error {
	if r.Headers == nil {
		return nil
	}

	var err error
	for k, v := range r.Headers {
		r.Headers[k], err = replacePlaceholders(vars, v, "request header")
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RequestEntry) validateAndParseBody(vars Vars) error {
	switch t := r.Body.(type) {
	case nil:
		return nil

	case []any:
		for _, entry := range t {
			if err := recursivelyParseBody(vars, entry); err != nil {
				return err
			}
		}
		return nil

	case any:
		return recursivelyParseBody(vars, t)

	default:
		return fmt.Errorf("only array and object types are supported for body")
	}
}

func recursivelyParseBody(vars Vars, body any) error {
	var err error
	switch t := body.(type) {
	case map[string]any:
		for k, v := range t {
			switch vt := v.(type) {
			case string:
				t[k], err = replacePlaceholders(vars, vt, "body value")
				if err != nil {
					return err
				}

			case map[string]any:
				return recursivelyParseBody(vars, vt)
			}
		}

	default:
		return nil
	}

	return nil
}

func (r *RequestEntry) validate(vars Vars) error {
	if r.Name == "" {
		return fmt.Errorf("request name is missing")
	}

	if err := r.validateAndParseUrl(vars); err != nil {
		return err
	}

	if err := r.Method.validate(); err != nil {
		return err
	}

	if (r.Method == RequestMethodGet || r.Method == RequestMethodDelete) && r.Body != nil {
		return fmt.Errorf("body is not supported for request method %q", r.Method)
	}

	if err := r.validateAndParseQuery(vars); err != nil {
		return err
	}

	if err := r.validateAndParseHeaders(vars); err != nil {
		return err
	}

	if err := r.validateAndParseBody(vars); err != nil {
		return err
	}

	return nil
}

type Collection struct {
	Vars     Vars            `toml:"vars"` // optional.
	Requests []*RequestEntry `toml:"requests"`
}

func (c *Collection) validate() error {
	if len(c.Requests) == 0 {
		return fmt.Errorf("no requests defined in collection")
	}

	for i, r := range c.Requests {
		if err := r.validate(c.Vars); err != nil {
			return fmt.Errorf("invalid request at index %d: %w", i, err)
		}
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

func Load(filepath string, envFilepath string) (*Collection, error) {
	if !fileExists(filepath) {
		return nil, fmt.Errorf("collection file %q not found", filepath)
	}

	if fileExists(envFilepath) {
		err := godotenv.Load(envFilepath)
		if err != nil {
			return nil, fmt.Errorf("failed to load environment file %q: %w", envFilepath, err)
		}

		// TODO: remove.
		// allEnvVars := os.Environ()
		// for _, v := range allEnvVars {
		// 	pieces := strings.Split(v, "=")
		// 	if len(pieces) != 2 {
		// 		continue
		// 	}

		// 	key := strings.TrimSpace(pieces[0])
		// 	value := strings.TrimSpace(pieces[1])
		// 	if key == "" || value == "" {
		// 		continue
		// 	}

		// 	collection.Vars[key] = value
		// }
	}

	var collection Collection
	if _, err := toml.DecodeFile(filepath, &collection); err != nil {
		return nil, fmt.Errorf("invalid contents in collection: %w", err)
	}

	if err := collection.validate(); err != nil {
		return nil, fmt.Errorf("invalid collection: %w", err)
	}

	// TODO: issue warning if duplicate vars in collection.
	return &collection, nil
}

func (c *Collection) ListRequests() []string {
	requestNames := make([]string, len(c.Requests))
	for i, r := range c.Requests {
		requestNames[i] = r.Name
	}

	return requestNames
}

func extractPlaceholders(template string) []string {
	matches := placeholderRegexp.FindAllStringSubmatch(template, -1)
	placeholders := make([]string, len(matches))
	for i, match := range matches {
		placeholders[i] = strings.TrimSpace(match[1])
	}

	// TODO: ensure there are no duplicates in the result slice.
	return placeholders
}

func replacePlaceholders(vars Vars, target string, targetName string) (string, error) {
	placeholders := extractPlaceholders(target)
	if len(placeholders) == 0 {
		return target, nil
	}

	result := target
	for _, p := range placeholders {
		value, ok := vars[p]
		if ok {
			result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", p), value)
			continue
		} else {
			// if placeholder is not found in collection, we look it up in env.
			envValue := os.Getenv(p)
			if envValue == "" {
				return "", fmt.Errorf("undefined placeholder in %s: %s", targetName, p)
			}
			result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", p), envValue)
		}
	}

	return result, nil
}
