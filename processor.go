package main

import (
	"errors"
	"io"

	yaml "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type fileFormat string

const (
	FormatYAML = fileFormat(".yaml")
)

type processor struct {
	Resources map[string]*unstructured.Unstructured
}

func newProcessor() *processor {
	return &processor{
		Resources: make(map[string]*unstructured.Unstructured),
	}
}

func (p *processor) ReadResources(input io.Reader, format fileFormat) error {
	if format != FormatYAML {
		return errors.New("format not supported")
	}

	r := yaml.NewDecoder(input)

	var err error
	for err == nil {
		var obj unstructured.Unstructured
		err = r.Decode(&obj.Object)
		if err != nil {
			break
		}
		p.Resources[generateResourceID(&obj)] = &obj
	}

	if err == io.EOF {
		err = nil
	}

	return err
}

func generateResourceID(u *unstructured.Unstructured) string {
	return u.GetKind() + "/" + u.GetNamespace() + "/" + u.GetName()
}

func (p *processor) RenderResources() ([]*unstructured.Unstructured, error) {
	var r []*unstructured.Unstructured
	for _, v := range p.Resources {
		r = append(r, v)
	}
	return r, nil
}
