package commons

import (
	"bytes"
	"github.com/pkg/errors"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"text/template"
)

func GetInt64Value(value int64, defalutValue int64) int64 {
	if value != 0 {
		return value
	} else {
		return defalutValue
	}
}

func GetIntValue(value int32, defalutValue int32) int32 {
	if value != 0 {
		return value
	} else {
		return defalutValue
	}
}

func GetStringValue(value string, defalutValue string) string {
	if value != "" {
		return value
	} else {
		return defalutValue
	}
}

func GetImagePullPolicy(policy v1.PullPolicy) v1.PullPolicy {
	if policy != "" {
		return policy
	} else {
		return v1.PullIfNotPresent
	}
}

func LoadTemplate(name, templatePath string, p interface{}) (string, error) {
	b, err := ioutil.ReadFile(filepath.Clean(templatePath))
	if err != nil {
		return "", err
	}
	data := string(b)
	var writer bytes.Buffer
	t := template.New(name)
	t, err = t.Parse(data)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse template %v", name)
	}
	err = t.Execute(&writer, p)
	return writer.String(), err
}

