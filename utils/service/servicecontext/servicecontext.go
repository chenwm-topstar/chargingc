package servicecontext

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

const serviceVarPrefix = "var/"

// ServiceContext help choose the tag to call the right services.
// New(2021 May): this context also serves the service variables instead of just service tags.
// The underlying map combines two kinds of context:
//   Service Tag: the key with no speicified prefix
//   Service Variables: the key with prefix "var/"
type ServiceContext map[string]string

type key int

const serviceContextKey key = iota

// Derive extracts service context from src context, to merge it into dest context.
func Derive(src, dest context.Context) context.Context {
	sc := FromContext(src)
	if sc == nil {
		return dest
	}

	return With(dest, sc)
}

// With embeds service context into general context
func With(ctx context.Context, sc ServiceContext) context.Context {
	return context.WithValue(ctx, serviceContextKey, sc)
}

// FromContext extracts service context from general context.
func FromContext(ctx context.Context) ServiceContext {
	sc, ok := ctx.Value(serviceContextKey).(ServiceContext)
	if !ok {
		return nil
	}
	return sc
}

// Merge two ServiceContext together, and returns a new ServiceContext.
func Merge(a, b ServiceContext) ServiceContext {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	out := ServiceContext{}
	for name, tag := range a {
		out[name] = tag
	}
	for name, tag := range b {
		out[name] = tag
	}

	return out
}

// New service tag spcification in format {ServiceName}:{Tag} into ServiceContext.
func New(specs []string) ServiceContext {
	out := ServiceContext{}
	for _, spec := range specs {
		parts := strings.SplitN(spec, ":", 2)
		if len(parts) != 2 {
			continue
		}
		out[parts[0]] = parts[1]
	}
	return out
}

// NewWithVariables service tags and variables into ServiceContext.
func NewWithVariables(services []string, vars map[string]string) ServiceContext {
	out := ServiceContext{}
	for _, spec := range services {
		parts := strings.SplitN(spec, ":", 2)
		if len(parts) != 2 {
			continue
		}
		out[parts[0]] = parts[1]
	}
	for k, v := range vars {
		out[serviceVarPrefix+k] = v
	}
	return out
}

// Specs returns a wired format to pass as string slices.
func (sc ServiceContext) Specs() []string {
	out := make([]string, len(sc))
	i := 0
	for name, tag := range sc {
		out[i] = fmt.Sprintf("%s:%s", name, tag)
		i++
	}
	return out
}

// Services returns only service names included in context.
func (sc ServiceContext) Services() []string {
	var out []string
	for name := range sc {
		if strings.HasPrefix(name, serviceVarPrefix) {
			// Not a service name
			continue
		}
		out = append(out, name)
	}
	return out
}

// Tag returns the tag for a given service name.
func (sc ServiceContext) Tag(serviceName string) string {
	if sc == nil {
		return ""
	}
	if strings.HasPrefix(serviceName, serviceVarPrefix) {
		// No a service name
		return ""
	}
	return sc[serviceName]
}

func (sc ServiceContext) Variable(key string) (string, bool) {
	if sc == nil {
		return "", false
	}

	val, ok := sc[serviceVarPrefix+key]
	return val, ok
}

func (sc ServiceContext) Hash() string {
	if sc == nil {
		return ""
	}
	var needles []string
	for k, v := range sc {
		needles = append(needles, k, v)
	}
	sort.Strings(needles)
	h := md5.New()
	for _, v := range needles {
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// InjectRequest embeds the service context into an http request,
// which will be recognized by CM MicroService framework.
func (sc ServiceContext) InjectRequest(r *http.Request) {
	for _, s := range sc.Specs() {
		r.Header.Add(ServiceContextMetadataKey, s)
	}
}

// FromRequest extracts the service context embeded in the http request.
func FromRequest(r *http.Request) ServiceContext {
	return New(r.Header.Values(ServiceContextMetadataKey))
}
