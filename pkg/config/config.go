package config

import (
	"bytes"

	"github.com/pkg/errors"
	kotsv1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	"github.com/replicatedhq/kots/kotskinds/multitype"
	"github.com/replicatedhq/kots/pkg/base"
	"github.com/replicatedhq/kots/pkg/logger"
	"github.com/replicatedhq/kots/pkg/template"
	"github.com/replicatedhq/kots/pkg/util"
	yaml "github.com/replicatedhq/yaml/v3"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

func TemplateConfig(log *logger.Logger, configSpecData string, configValuesData string, localRegistry template.LocalRegistry) (string, error) {
	return templateConfig(log, configSpecData, configValuesData, localRegistry, MarshalConfig)
}

func templateConfig(log *logger.Logger, configSpecData string, configValuesData string, localRegistry template.LocalRegistry, marshalFunc func(config *kotsv1beta1.Config) (string, error)) (string, error) {
	// This function will
	// 1. unmarshal config
	// 2. replace all item values with values that already exist
	// 3. re-marshal it (with an unlimited line length)
	// 4. put new config yaml through templating engine
	// This process will re-order items and discard comments, so it should not be saved.

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(configSpecData), nil, nil) // TODO fix decode of boolstrings
	if err != nil {
		return "", errors.Wrap(err, "failed to decode config data")
	}
	config := obj.(*kotsv1beta1.Config)

	builder := template.Builder{}
	builder.AddCtx(template.StaticCtx{})

	// get template context from config values
	templateContext, err := base.UnmarshalConfigValuesContent([]byte(configValuesData))
	if err != nil {
		log.Error(err)
		templateContext = map[string]template.ItemValue{}
	}

	// add config context
	configCtx, err := builder.NewConfigContext(config.Spec.Groups, templateContext, localRegistry, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create config context")
	}

	ApplyValuesToConfig(config, configCtx.ItemValues)
	configDocWithData, err := marshalFunc(config)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal config")
	}

	builder.AddCtx(configCtx)

	rendered, err := builder.RenderTemplate("config", string(configDocWithData))
	if err != nil {
		return "", errors.Wrap(err, "failed to render config template")
	}

	return rendered, nil
}

func ApplyValuesToConfig(config *kotsv1beta1.Config, values map[string]template.ItemValue) {
	for idxG, g := range config.Spec.Groups {
		for idxI, i := range g.Items {
			value, ok := values[i.Name]
			if ok {
				config.Spec.Groups[idxG].Items[idxI].Value = multitype.FromString(value.ValueStr())
				config.Spec.Groups[idxG].Items[idxI].Default = multitype.FromString(value.DefaultStr())
			}
			for idxC, c := range i.Items {
				value, ok := values[c.Name]
				if ok {
					config.Spec.Groups[idxG].Items[idxI].Items[idxC].Value = multitype.FromString(value.ValueStr())
					config.Spec.Groups[idxG].Items[idxI].Items[idxC].Default = multitype.FromString(value.DefaultStr())
				}
			}
		}
	}
}

// MarshalConfig runs the same code path as the k8s json->yaml serializer, but uses a different yaml library for those parts
// first, the object is marshalled to json
// second, the json is unmarshalled to an object as yaml
// third, the object is marshalled as yaml
func MarshalConfig(config *kotsv1beta1.Config) (string, error) {
	s := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme.Scheme,
		scheme.Scheme,
		json.SerializerOptions{Yaml: false, Pretty: true, Strict: false},
	)

	var marshalledJSON bytes.Buffer
	if err := s.Encode(config, &marshalledJSON); err != nil {
		return "", errors.Wrap(err, "failed to marshal config as json")
	}

	var unmarshalledYAML interface{}
	if err := yaml.Unmarshal(marshalledJSON.Bytes(), &unmarshalledYAML); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal config as yaml")
	}

	marshalledYAML, err := util.MarshalIndent(2, unmarshalledYAML)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal config as yaml")
	}

	return string(marshalledYAML), nil
}
